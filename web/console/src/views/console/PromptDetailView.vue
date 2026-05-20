<script setup lang="ts">
import {
  createVersion,
  diffVersions,
  getChannelPointer,
  getPrompt,
  listVersions,
  patchChannelPointer,
  putDraft,
  type ChannelPointerResp,
  type PromptVersionItem,
} from "@/api/prompts";
import { ApiError } from "@/api/client";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import {
  expandBodyForDisplay,
  extractPlaceholderNames,
  formatPlaceholder,
  getAtomicPlaceholderDeleteRange,
  getPlaceholderTrigger,
  hasDebugSampleValues,
  isValidVarName,
  mapDisplaySelectionToDraft,
  mapDraftSelectionToDisplay,
  normalizeDraftBodyForSave,
  normalizeSchemaJson,
  renderBodyHighlightHtml,
  sanitizeVarName,
} from "@/lib/prompt-placeholders";
import { formatVersionDisplay, suggestNextVersionLabel } from "@/lib/prompt-version";
import { usePromptBodyHistory } from "@/composables/use-prompt-body-history";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { onBeforeRouteLeave, useRoute, useRouter } from "vue-router";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const ws = useWorkspaceContextStore();

const ns = computed(() => ws.selectedNsSlug);
const promptKey = computed(() => String(route.params.promptKey ?? ""));
const isArchived = computed(() => ws.selectedNamespace?.status === "archived");

// ── Navigation & data ──────────────────────────────────────────────
const detail = ref<Awaited<ReturnType<typeof getPrompt>> | null>(null);
const loading = ref(false);
const err = ref<string | null>(null);

// ── Draft state ────────────────────────────────────────────────────
const draftBody = ref("");
const draftSchema = ref("");
const baselineBody = ref("");
const baselineSchema = ref("");
const saveBusy = ref(false);
const saveMsg = ref<string | null>(null);
const saveStatus = ref<"saved" | "dirty" | "saving">("saved");
let saveTimeout: ReturnType<typeof setTimeout> | null = null;

// ── Schema variable rows ──────────────────────────────────────────
interface SchemaVarRow {
  name: string;
  type: "string" | "number" | "boolean";
  required: boolean;
  description: string;
}

const schemaVars = ref<SchemaVarRow[]>([]);
const showSchemaJson = ref(false);
const schemaJsonErr = ref<string | null>(null);
/** Skip body→schema sync while hydrating from API (avoids false dirty state). */
const hydrating = ref(false);

function loadSchemaIntoVars() {
  const raw = draftSchema.value.trim() || "[]";
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      schemaJsonErr.value = t("prompts.schemaNotArray");
      return;
    }
    schemaVars.value = parsed.map((e: any) => ({
      name: String(e?.name ?? "").trim(),
      type: (["string", "number", "boolean"].includes(e?.type) ? e.type : "string") as SchemaVarRow["type"],
      required: Boolean(e?.required),
      description: String(e?.description ?? ""),
    }));
    schemaJsonErr.value = null;
  } catch {
    schemaJsonErr.value = t("prompts.schemaInvalidJson");
  }
}

function syncVarsToJson() {
  try {
    const arr = schemaVars.value.map((v) => {
      const o: any = { name: v.name };
      if (v.type !== "string") o.type = v.type;
      if (v.required) o.required = true;
      if (v.description.trim()) o.description = v.description.trim();
      return o;
    });
    draftSchema.value = JSON.stringify(arr);
    schemaJsonErr.value = null;
  } catch {
    // ignore
  }
}

/** Sync schema JSON from {{placeholders}} in the draft body (preserves type/required/description). */
function syncSchemaFromBody() {
  if (showSchemaJson.value) return;
  const detected = extractPlaceholderNames(draftBody.value);
  const byName = new Map(
    schemaVars.value.filter((v) => v.name).map((v) => [v.name, v] as const),
  );
  const next: SchemaVarRow[] = detected.map(
    (name) => byName.get(name) ?? { name, type: "string", required: false, description: "" },
  );
  const same =
    next.length === schemaVars.value.length &&
    next.every((v, i) => v.name === schemaVars.value[i]?.name);
  if (!same) schemaVars.value = next;
  syncVarsToJson();
}

function onToggleSchemaJson() {
  if (!showSchemaJson.value) {
    syncVarsToJson();
    showSchemaJson.value = true;
  } else {
    loadSchemaIntoVars();
    showSchemaJson.value = false;
    syncSchemaFromBody();
  }
}

watch(draftBody, () => {
  if (hydrating.value) return;
  syncSchemaFromBody();
  recomputeSampleKeys();
});

const detectedVarNames = computed(() => extractPlaceholderNames(draftBody.value));

// ── Sample values for in-editor debug preview ─────────────────────
const sampleValues = ref<Record<string, string>>({});

function recomputeSampleKeys() {
  const names = detectedVarNames.value;
  const next: Record<string, string> = {};
  for (const n of names) {
    if (sampleValues.value[n] !== undefined) next[n] = sampleValues.value[n];
  }
  sampleValues.value = next;
}

// ── Body editor refs & helpers ─────────────────────────────────────
const bodyTextareaRef = ref<HTMLTextAreaElement | null>(null);
const bodyHighlightRef = ref<HTMLPreElement | null>(null);
const lineNumsRef = ref<HTMLDivElement | null>(null);

const lineCount = computed(() => draftBody.value.split("\n").length);

const bodyHighlightHtml = computed(() =>
  renderBodyHighlightHtml(draftBody.value, sampleValues.value),
);

const hasDebugPreview = computed(() => hasDebugSampleValues(sampleValues.value));

const editorLayoutBody = computed({
  get() {
    if (!hasDebugPreview.value) return draftBody.value;
    return expandBodyForDisplay(draftBody.value, sampleValues.value);
  },
  set(v: string) {
    draftBody.value = normalizeDraftBodyForSave(v);
  },
});

const bodyHistory = usePromptBodyHistory({
  getBody: () => draftBody.value,
  setBody: (body) => {
    draftBody.value = body;
  },
  getTextarea: () => bodyTextareaRef.value,
  hasDebugPreview: () => hasDebugPreview.value,
  getSampleValues: () => sampleValues.value,
});

function setEditorBody(next: string, selStart: number, selEnd = selStart) {
  if (hasDebugPreview.value) {
    editorLayoutBody.value = next;
  } else {
    draftBody.value = next;
  }
  void nextTick(() => {
    const ta = bodyTextareaRef.value;
    if (!ta) return;
    ta.selectionStart = selStart;
    ta.selectionEnd = selEnd;
  });
}

function syncEditorScroll() {
  const ta = bodyTextareaRef.value;
  if (!ta) return;
  if (lineNumsRef.value) lineNumsRef.value.scrollTop = ta.scrollTop;
  if (bodyHighlightRef.value) {
    bodyHighlightRef.value.scrollTop = ta.scrollTop;
    bodyHighlightRef.value.scrollLeft = ta.scrollLeft;
  }
}

function onBodyKeydown(ev: KeyboardEvent) {
  const ta = bodyTextareaRef.value;
  if (!ta) return;

  if (varSuggestOpen.value && varSuggestItems.value.length) {
    if (ev.key === "ArrowDown") {
      ev.preventDefault();
      varSuggestIndex.value = (varSuggestIndex.value + 1) % varSuggestItems.value.length;
      return;
    }
    if (ev.key === "ArrowUp") {
      ev.preventDefault();
      const len = varSuggestItems.value.length;
      varSuggestIndex.value = (varSuggestIndex.value - 1 + len) % len;
      return;
    }
    if (ev.key === "Enter" || ev.key === "Tab") {
      ev.preventDefault();
      applyVarSuggestAt(varSuggestIndex.value);
      return;
    }
    if (ev.key === "Escape") {
      ev.preventDefault();
      closeVarSuggest();
      return;
    }
  }

  if (ev.key === "Backspace" || ev.key === "Delete") {
    const start = ta.selectionStart;
    const end = ta.selectionEnd;
    const debugPreview = hasDebugPreview.value;
    const val = ta.value;
    const range = getAtomicPlaceholderDeleteRange(val, start, end, ev.key, debugPreview);
    if (range) {
      ev.preventDefault();
      bodyHistory.noteBeforeProgrammaticEdit();
      const next = val.slice(0, range.start) + val.slice(range.end);
      setEditorBody(next, range.start);
      closeVarSuggest();
      return;
    }
  }

  if (ev.key === "Tab") {
    ev.preventDefault();
    bodyHistory.noteBeforeProgrammaticEdit();
    const start = ta.selectionStart;
    const end = ta.selectionEnd;
    const val = ta.value;
    setEditorBody(val.slice(0, start) + "  " + val.slice(end), start + 2);
  } else if (ev.key === "Enter") {
    ev.preventDefault();
    bodyHistory.noteBeforeProgrammaticEdit();
    const start = ta.selectionStart;
    const val = ta.value;
    const lineStart = val.lastIndexOf("\n", start - 1) + 1;
    const indent = val.slice(lineStart).match(/^(\s*)/)?.[1] ?? "";
    const nextPos = start + 1 + indent.length;
    setEditorBody(val.slice(0, start) + "\n" + indent + val.slice(start), nextPos);
  }
}

function insertVarAtCursor(varName: string) {
  const ta = bodyTextareaRef.value;
  if (!ta) return;
  bodyHistory.noteBeforeProgrammaticEdit();
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  const val = ta.value;
  const insert = formatPlaceholder(varName);
  setEditorBody(val.slice(0, start) + insert + val.slice(end), start + insert.length);
  closeVarSuggest();
  ta.focus();
}

function replaceVarSuggest(name: string) {
  const ta = bodyTextareaRef.value;
  if (!ta) return;
  bodyHistory.noteBeforeProgrammaticEdit();
  const start = varSuggestStart.value;
  const end = hasDebugPreview.value
    ? mapDisplaySelectionToDraft(
        draftBody.value,
        sampleValues.value,
        ta.selectionStart,
        ta.selectionStart,
      ).selectionStart
    : ta.selectionStart;
  const insert = formatPlaceholder(name);
  const nextBody = draftBody.value.slice(0, start) + insert + draftBody.value.slice(end);
  const nextDraftCursor = start + insert.length;
  draftBody.value = nextBody;
  closeVarSuggest();
  void nextTick(() => {
    if (!ta) return;
    const sel = hasDebugPreview.value
      ? mapDraftSelectionToDisplay(
          nextBody,
          sampleValues.value,
          nextDraftCursor,
          nextDraftCursor,
        )
      : { start: nextDraftCursor, end: nextDraftCursor };
    ta.selectionStart = sel.start;
    ta.selectionEnd = sel.end;
    ta.focus();
  });
}

// ── Variable picker & {{ autocomplete ─────────────────────────────
const varPickerOpen = ref(false);
const newVarName = ref("");
const newVarInputRef = ref<HTMLInputElement | null>(null);

const varSuggestOpen = ref(false);
const varSuggestStart = ref(0);
const varSuggestQuery = ref("");
const varSuggestIndex = ref(0);

type VarSuggestItem = { kind: "existing" | "create"; name: string };

const varSuggestItems = computed<VarSuggestItem[]>(() => {
  const q = varSuggestQuery.value;
  const qLower = q.toLowerCase();
  const existing = detectedVarNames.value.filter((n) => n.toLowerCase().startsWith(qLower));
  const items: VarSuggestItem[] = existing.map((name) => ({ kind: "existing", name }));
  const trimmed = q.trim();
  if (trimmed && isValidVarName(trimmed) && !detectedVarNames.value.includes(trimmed)) {
    items.push({ kind: "create", name: trimmed });
  }
  return items;
});

const newVarNameValid = computed(() => isValidVarName(newVarName.value));

function closeVarSuggest() {
  varSuggestOpen.value = false;
  varSuggestQuery.value = "";
  varSuggestIndex.value = 0;
}

function refreshVarSuggest() {
  const ta = bodyTextareaRef.value;
  if (!ta || showSchemaJson.value || isArchived.value) {
    closeVarSuggest();
    return;
  }
  const body = draftBody.value;
  const cursor = hasDebugPreview.value
    ? mapDisplaySelectionToDraft(body, sampleValues.value, ta.selectionStart, ta.selectionStart)
        .selectionStart
    : ta.selectionStart;
  const trigger = getPlaceholderTrigger(body, cursor);
  if (!trigger) {
    closeVarSuggest();
    return;
  }
  const prevQuery = varSuggestQuery.value;
  const wasOpen = varSuggestOpen.value;
  varSuggestOpen.value = true;
  varSuggestStart.value = trigger.start;
  varSuggestQuery.value = trigger.query;
  if (!wasOpen || prevQuery !== trigger.query) {
    varSuggestIndex.value = 0;
    return;
  }
  const len = varSuggestItems.value.length;
  if (len === 0) {
    varSuggestIndex.value = 0;
  } else if (varSuggestIndex.value >= len) {
    varSuggestIndex.value = len - 1;
  }
}

function onBodyKeyup(ev: KeyboardEvent) {
  if (
    ev.key === "Escape" ||
    (varSuggestOpen.value &&
      (ev.key === "ArrowUp" || ev.key === "ArrowDown" || ev.key === "Enter" || ev.key === "Tab"))
  ) {
    return;
  }
  refreshVarSuggest();
}

function openVarPicker() {
  if (isArchived.value) return;
  varPickerOpen.value = true;
  const ta = bodyTextareaRef.value;
  if (ta && ta.selectionStart !== ta.selectionEnd) {
    newVarName.value = sanitizeVarName(ta.value.slice(ta.selectionStart, ta.selectionEnd));
  } else {
    newVarName.value = "";
  }
  void nextTick(() => newVarInputRef.value?.focus());
}

function closeVarPicker() {
  varPickerOpen.value = false;
  newVarName.value = "";
}

function confirmNewVar() {
  const name = sanitizeVarName(newVarName.value);
  if (!isValidVarName(name)) return;
  insertVarAtCursor(name);
  closeVarPicker();
}

function applyVarSuggestAt(index: number) {
  const item = varSuggestItems.value[index];
  if (!item) return;
  replaceVarSuggest(item.name);
}

function onVarSuggestMouseDown(index: number) {
  applyVarSuggestAt(index);
}

// ── Live validation ───────────────────────────────────────────────
interface ValidationIssue { level: "error" | "warn"; message: string }

const validationIssues = computed<ValidationIssue[]>(() => {
  const issues: ValidationIssue[] = [];
  const nameRe = /^[a-zA-Z0-9_]+$/;
  const validPlaceholder = /^\{\{\s*[a-zA-Z0-9_]+\s*\}\}$/;

  for (const raw of draftBody.value.match(/\{\{[^}]*\}\}/g) ?? []) {
    if (!validPlaceholder.test(raw)) {
      issues.push({
        level: "error",
        message: t("prompts.invalidPlaceholder", { placeholder: raw }),
      });
    }
  }

  for (const name of detectedVarNames.value) {
    if (!nameRe.test(name)) {
      issues.push({ level: "error", message: t("prompts.schemaInvalidName", { name }) });
    }
  }

  return issues;
});

const hasErrors = computed(() => validationIssues.value.some((i) => i.level === "error"));

// ── Dirty detection ───────────────────────────────────────────────
const isDraftDirty = computed(
  () =>
    !isArchived.value &&
    (draftBody.value !== baselineBody.value ||
      normalizeSchemaJson(draftSchema.value) !== normalizeSchemaJson(baselineSchema.value)),
);

watch(isDraftDirty, (dirty) => {
  if (dirty && saveStatus.value === "saved") saveStatus.value = "dirty";
});

// ── Versions ──────────────────────────────────────────────────────
const versions = ref<PromptVersionItem[]>([]);
const verLoading = ref(false);
const snapNote = ref("");
const snapCustomVersion = ref("");
const snapBusy = ref(false);

const suggestedNextVersion = computed(() =>
  suggestNextVersionLabel(versions.value.map((v) => v.version)),
);

// ── Diff ──────────────────────────────────────────────────────────
const diffBaseId = ref("");
const diffCompareId = ref("");
const diffOut = ref<{ body_diff: string; schema_diff: string } | null>(null);
const diffBusy = ref(false);
const diffErr = ref<string | null>(null);

const versionsNewestFirst = computed(() => [...versions.value].reverse());

function versionLabelById(id: string): string {
  return versions.value.find((v) => v.id === id)?.version ?? "";
}

function primeDiffSelection() {
  if (diffBaseId.value || diffCompareId.value) return;
  const vs = versions.value;
  if (vs.length >= 2) {
    diffBaseId.value = vs[vs.length - 2]!.id;
    diffCompareId.value = vs[vs.length - 1]!.id;
  } else if (vs.length === 1) {
    diffBaseId.value = vs[0]!.id;
    diffCompareId.value = vs[0]!.id;
  }
}

function setDiffBase(id: string) {
  diffBaseId.value = id;
}

function setDiffCompare(id: string) {
  diffCompareId.value = id;
}

function compareWithPrevious(indexInList: number) {
  if (indexInList <= 0) return;
  diffBaseId.value = versions.value[indexInList - 1]!.id;
  diffCompareId.value = versions.value[indexInList]!.id;
}

// ── Publish ───────────────────────────────────────────────────────
const publishChannel = computed(
  () => ws.selectedNamespace?.default_channel_slug?.trim() || "production",
);
const pointer = ref<ChannelPointerResp | null>(null);
const pointerLoading = ref(false);
const pointerVersionId = ref("");
const pointerBusy = ref(false);
const publishVersionSearch = ref("");

const publishVersionOptions = computed(() => {
  const q = publishVersionSearch.value.trim().toLowerCase();
  let list = versions.value;
  if (q) {
    list = list.filter((v) => {
      const label = v.version.toLowerCase();
      const display = formatVersionDisplay(v.version).toLowerCase();
      const note = v.change_note?.toLowerCase() ?? "";
      return label.includes(q) || display.includes(q) || note.includes(q);
    });
  }
  if (
    pointerVersionId.value &&
    !list.some((v) => v.id === pointerVersionId.value)
  ) {
    const current = versions.value.find((v) => v.id === pointerVersionId.value);
    if (current) list = [current, ...list];
  }
  return list;
});

// ── Release tool dialogs ───────────────────────────────────────────
const snapshotDlgEl = ref<HTMLDialogElement | null>(null);
const publishDlgEl = ref<HTMLDialogElement | null>(null);
const versionsDlgEl = ref<HTMLDialogElement | null>(null);

function openSnapshotDlg() {
  if (isArchived.value) return;
  snapshotDlgEl.value?.showModal();
}

function closeSnapshotDlg() {
  snapshotDlgEl.value?.close();
}

function onSnapshotDlgCancel(ev: Event) {
  if (snapBusy.value) ev.preventDefault();
}

function openPublishDlg() {
  if (isArchived.value) return;
  void loadReleaseData();
  publishDlgEl.value?.showModal();
}

function closePublishDlg() {
  publishDlgEl.value?.close();
}

function onPublishDlgCancel(ev: Event) {
  if (pointerBusy.value) ev.preventDefault();
}

function openVersionsDlg() {
  void loadVersions().then(() => {
    primeDiffSelection();
    void runDiffIfReady();
  });
  versionsDlgEl.value?.showModal();
}

function closeVersionsDlg() {
  versionsDlgEl.value?.close();
}

// ── Unsaved-leave dialog ──────────────────────────────────────────
const unsavedDlgOpen = ref(false);
const pendingLeaveFullPath = ref<string | null>(null);
/** When leave was triggered by a namespace switch, revert selection if user stays. */
const pendingLeavePrevNs = ref<string | null>(null);
const leaveBusy = ref(false);
const allowRouteLeaveOnce = ref(false);

// ── Data loading ──────────────────────────────────────────────────
async function loadDetail() {
  const slug = ns.value;
  const key = promptKey.value;
  if (!slug || !key) return;
  loading.value = true;
  err.value = null;
  try {
    const d = await getPrompt(slug, key);
    detail.value = d;
    hydrating.value = true;
    draftBody.value = d.draft_body ?? "";
    draftSchema.value = d.draft_schema ?? "[]";
    loadSchemaIntoVars();
    syncSchemaFromBody();
    baselineBody.value = draftBody.value;
    baselineSchema.value = draftSchema.value;
    saveStatus.value = "saved";
    recomputeSampleKeys();
    bodyHistory.reset();
  } catch (e) {
    detail.value = null;
    err.value =
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.detailLoadError");
  } finally {
    hydrating.value = false;
    loading.value = false;
  }
}

function primeDiffFromVersions() {
  primeDiffSelection();
}

async function runDiffIfReady() {
  const slug = ns.value;
  const key = promptKey.value;
  const baseId = diffBaseId.value;
  const compareId = diffCompareId.value;
  if (!slug || !key || !baseId || !compareId) return;
  if (baseId === compareId) {
    diffOut.value = null;
    diffErr.value = null;
    return;
  }
  const a = versionLabelById(baseId);
  const b = versionLabelById(compareId);
  if (!a || !b) return;
  diffBusy.value = true;
  diffErr.value = null;
  diffOut.value = null;
  try {
    diffOut.value = await diffVersions(slug, key, a, b);
  } catch (e) {
    diffOut.value = null;
    diffErr.value =
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.diffError");
  } finally {
    diffBusy.value = false;
  }
}

async function loadVersions() {
  const slug = ns.value;
  const key = promptKey.value;
  if (!slug || !key) return;
  verLoading.value = true;
  try {
    const res = await listVersions(slug, key, { page: 1, page_size: 100 });
    versions.value = res.items ?? [];
    primeDiffFromVersions();
  } catch {
    versions.value = [];
  } finally {
    verLoading.value = false;
  }
}

async function loadReleaseData() {
  await Promise.all([loadVersions(), loadPointer()]);
}

async function loadPointer() {
  const slug = ns.value;
  const key = promptKey.value;
  const ch = publishChannel.value;
  if (!slug || !key || !ch) return;
  pointerLoading.value = true;
  pointer.value = null;
  try {
    pointer.value = await getChannelPointer(slug, key, ch);
    pointerVersionId.value = pointer.value.version_id;
  } catch (e) {
    pointer.value = null;
    if (e instanceof ApiError && e.status === 404) {
      /* no pointer yet */
    }
  } finally {
    pointerLoading.value = false;
  }
}

// ── Save ──────────────────────────────────────────────────────────
async function onSaveDraft(): Promise<boolean> {
  const slug = ns.value;
  const key = promptKey.value;
  if (!slug || !key || isArchived.value) return false;
  saveBusy.value = true;
  saveStatus.value = "saving";
  saveMsg.value = null;
  try {
    const bodyToSave = normalizeDraftBodyForSave(draftBody.value);
    if (bodyToSave !== draftBody.value) {
      draftBody.value = bodyToSave;
    }
    const res = await putDraft(slug, key, {
      body: bodyToSave,
      schema: draftSchema.value,
    });
    saveMsg.value =
      res.warnings?.length
        ? `${t("prompts.savedWithWarnings")}: ${res.warnings.join("; ")}`
        : t("prompts.saved");
    await loadDetail();
    baselineBody.value = draftBody.value;
    baselineSchema.value = draftSchema.value;
    saveStatus.value = "saved";
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(() => {
      saveMsg.value = null;
    }, 3000);
    return true;
  } catch (e) {
    saveMsg.value =
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.saveError");
    saveStatus.value = "dirty";
    return false;
  } finally {
    saveBusy.value = false;
  }
}

function onGlobalKeydown(ev: KeyboardEvent) {
  if ((ev.ctrlKey || ev.metaKey) && ev.key === "s") {
    ev.preventDefault();
    void onSaveDraft();
    return;
  }
  if ((ev.ctrlKey || ev.metaKey) && ev.shiftKey && ev.key.toLowerCase() === "v") {
    ev.preventDefault();
    openVarPicker();
    return;
  }

  const ta = bodyTextareaRef.value;
  const inBodyEditor = ta && document.activeElement === ta && !showSchemaJson.value;
  if (!inBodyEditor || isArchived.value) return;

  const key = ev.key.toLowerCase();
  if ((ev.ctrlKey || ev.metaKey) && key === "z") {
    ev.preventDefault();
    if (ev.shiftKey) bodyHistory.redo();
    else bodyHistory.undo();
    return;
  }
  if ((ev.ctrlKey || ev.metaKey) && key === "y") {
    ev.preventDefault();
    bodyHistory.redo();
  }
}

// ── Snapshot / Diff / Publish ─────────────────────────────────────
async function onSnapshot() {
  const slug = ns.value;
  const key = promptKey.value;
  if (!slug || !key || isArchived.value) return;
  snapBusy.value = true;
  try {
    const body: { change_note?: string; version?: string } = {};
    const note = snapNote.value.trim();
    if (note) body.change_note = note;
    const customVer = snapCustomVersion.value.trim();
    if (customVer) body.version = customVer;
    const created = await createVersion(slug, key, body);
    snapNote.value = "";
    snapCustomVersion.value = "";
    await loadVersions();
    if (created?.id) pointerVersionId.value = created.id;
    closeSnapshotDlg();
  } catch (e) {
    window.alert(
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.snapshotError"),
    );
  } finally {
    snapBusy.value = false;
  }
}

async function onPublishPointer() {
  const slug = ns.value;
  const key = promptKey.value;
  const ch = publishChannel.value;
  const vid = pointerVersionId.value.trim();
  if (!slug || !key || !vid || isArchived.value) return;
  pointerBusy.value = true;
  try {
    pointer.value = await patchChannelPointer(slug, key, ch, { version_id: vid });
    pointerVersionId.value = pointer.value.version_id;
    closePublishDlg();
  } catch (e) {
    window.alert(
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.pointerError"),
    );
  } finally {
    pointerBusy.value = false;
  }
}

// ── Unsaved-leave dialog ──────────────────────────────────────────
const leaveDlgRef = ref<HTMLDialogElement | null>(null);

watch(unsavedDlgOpen, async (open) => {
  await nextTick();
  const el = leaveDlgRef.value;
  if (!el) return;
  if (open && !el.open) el.showModal();
  if (!open && el.open) el.close();
});

function leaveDlgClose() {
  const revertNs = pendingLeavePrevNs.value;
  pendingLeaveFullPath.value = null;
  pendingLeavePrevNs.value = null;
  unsavedDlgOpen.value = false;
  if (revertNs) {
    ws.setSelectedNsSlug(revertNs);
  }
}

function openUnsavedLeavePath(fullPath: string) {
  pendingLeaveFullPath.value = fullPath;
  unsavedDlgOpen.value = true;
}

onBeforeRouteLeave((to) => {
  if (allowRouteLeaveOnce.value) {
    allowRouteLeaveOnce.value = false;
    return true;
  }
  if (!isDraftDirty.value) return true;
  openUnsavedLeavePath(to.fullPath);
  return false;
});

function onBeforeUnload(ev: BeforeUnloadEvent) {
  if (isDraftDirty.value) {
    ev.preventDefault();
  }
}

async function confirmSaveAndExit() {
  const path = pendingLeaveFullPath.value;
  if (!path) {
    leaveDlgClose();
    return;
  }
  leaveBusy.value = true;
  try {
    const ok = await onSaveDraft();
    if (!ok) return;
    pendingLeavePrevNs.value = null;
    pendingLeaveFullPath.value = null;
    unsavedDlgOpen.value = false;
    allowRouteLeaveOnce.value = true;
    await router.push(path);
  } finally {
    leaveBusy.value = false;
  }
}

function confirmDiscardAndExit() {
  const path = pendingLeaveFullPath.value;
  pendingLeavePrevNs.value = null;
  pendingLeaveFullPath.value = null;
  unsavedDlgOpen.value = false;
  if (!path) return;
  allowRouteLeaveOnce.value = true;
  void router.push(path);
}

function leaveDetailOnNamespaceChange(prevNs: string) {
  const workspacePath = router.resolve({ name: "workspace" }).fullPath;
  if (!isDraftDirty.value) {
    allowRouteLeaveOnce.value = true;
    void router.push({ name: "workspace" });
    return;
  }
  pendingLeavePrevNs.value = prevNs;
  openUnsavedLeavePath(workspacePath);
}

function back() {
  if (!isDraftDirty.value) {
    void router.push({ name: "workspace" });
    return;
  }
  openUnsavedLeavePath(router.resolve({ name: "workspace" }).fullPath);
}

// ── Watchers & lifecycle ──────────────────────────────────────────
watch(ns, (next, prev) => {
  if (prev === next || prev === "") return;
  leaveDetailOnNamespaceChange(prev);
});

watch(promptKey, async () => {
  await loadDetail();
  await loadReleaseData();
});

watch(publishChannel, () => void loadPointer());

watch([diffBaseId, diffCompareId], () => {
  void runDiffIfReady();
});

onMounted(async () => {
  window.addEventListener("beforeunload", onBeforeUnload);
  window.addEventListener("keydown", onGlobalKeydown);
  await loadDetail();
  await loadReleaseData();
});

onUnmounted(() => {
  window.removeEventListener("beforeunload", onBeforeUnload);
  window.removeEventListener("keydown", onGlobalKeydown);
});
</script>

<template>
  <div class="detail">
    <button type="button" class="back" @click="back">{{ t("prompts.backToList") }}</button>

    <header class="head">
      <h1 class="title">{{ detail?.title || promptKey }}</h1>
    </header>

    <p v-if="isArchived" class="warn">{{ t("prompts.archivedNoWrite") }}</p>
    <p v-if="loading" class="muted">{{ t("prompts.loading") }}</p>
    <p v-else-if="err" class="err">{{ err }}</p>

    <section v-if="detail" class="panel">
      <div class="work-toolbar" role="toolbar" :aria-label="t('prompts.workToolbarAria')">
        <div class="toolbar-draft" role="status" aria-live="polite">
          <span class="save-state" :class="saveStatus">
            <span
              class="save-dot"
              :class="{ 'saved-dot': saveStatus === 'saved', 'dirty-dot': saveStatus === 'dirty' }"
            />
            <span v-if="saveStatus === 'saving'">{{ t("prompts.loading") }}</span>
            <span v-else-if="saveStatus === 'saved'">{{ t("prompts.statusSaved") }}</span>
            <span v-else>{{ t("prompts.statusUnsaved") }}</span>
          </span>
          <button type="button" class="btn sm" :disabled="saveBusy || isArchived" @click="onSaveDraft">
            {{ t("prompts.saveDraft") }}
          </button>
          <span
            v-if="saveMsg"
            class="save-msg small"
            :class="saveMsg.includes('warning') || saveMsg.includes('警告') ? 'warn' : 'ok'"
          >
            {{ saveMsg }}
          </span>
        </div>

        <div class="toolbar-spacer" aria-hidden="true" />

        <div class="toolbar-release">
          <span v-if="pointerLoading" class="live-badge muted small">{{ t("prompts.loading") }}</span>
          <span v-else-if="pointer" class="live-badge mono" :title="t('prompts.currentPointer')">
            {{ formatVersionDisplay(pointer.version) }}
          </span>
          <button
            type="button"
            class="btn sm tool"
            :disabled="isArchived"
            @click="openSnapshotDlg"
          >
            {{ t("prompts.toolbarSnapshot") }}
          </button>
          <button
            type="button"
            class="btn sm tool"
            :disabled="isArchived"
            @click="openPublishDlg"
          >
            {{ t("prompts.toolbarPublish") }}
          </button>
          <button type="button" class="btn sm tool" @click="openVersionsDlg">
            {{ t("prompts.toolbarVersions") }}
          </button>
        </div>
      </div>

      <div class="edit-layout">
        <div class="col col-editor">
          <div class="panel-head">
            <h2 class="col-title">{{ t("prompts.bodyLabel") }}</h2>
            <button type="button" class="link-btn" @click="onToggleSchemaJson">
              {{ showSchemaJson ? t("prompts.hideJson") : t("prompts.schemaJson") }}
            </button>
          </div>

          <div v-if="!showSchemaJson" class="var-strip">
            <span class="var-strip-label">{{ t("prompts.varsDetected") }}</span>
            <div v-if="detectedVarNames.length" class="var-chips" role="list">
              <button
                v-for="name in detectedVarNames"
                :key="name"
                type="button"
                class="chip var-chip mono"
                role="listitem"
                :title="t('prompts.insertVarAtCursor', { name })"
                @click="insertVarAtCursor(name)"
              >
                {{ name }}
              </button>
            </div>
            <span v-else class="var-empty-hint">{{ t("prompts.varsEmpty") }}</span>
            <button
              type="button"
              class="btn sm var-add-btn"
              :disabled="isArchived"
              :title="t('prompts.insertVarToolbar')"
              @click="openVarPicker"
            >
              + {{ t("prompts.addVariable") }}
            </button>
          </div>

          <div v-if="varPickerOpen && !showSchemaJson" class="var-picker">
            <label class="var-picker-label" for="inkforge-new-var">{{ t("prompts.varName") }}</label>
            <div class="var-picker-row">
              <input
                id="inkforge-new-var"
                ref="newVarInputRef"
                v-model="newVarName"
                class="inp mono var-picker-inp"
                type="text"
                spellcheck="false"
                :placeholder="t('prompts.varNamePlaceholder')"
                @keydown.enter.prevent="confirmNewVar"
                @keydown.escape="closeVarPicker"
              />
              <button
                type="button"
                class="btn sm"
                :disabled="!newVarNameValid"
                @click="confirmNewVar"
              >
                {{ t("prompts.insertVar") }}
              </button>
              <button type="button" class="btn sm ghost" @click="closeVarPicker">
                {{ t("common.close") }}
              </button>
            </div>
            <p v-if="newVarName.trim() && !newVarNameValid" class="var-picker-err small">
              {{ t("prompts.varNameInvalid") }}
            </p>
            <p class="var-picker-hint muted small">{{ t("prompts.varPickerHint") }}</p>
          </div>

          <div v-if="!showSchemaJson && detectedVarNames.length" class="sample-row">
            <span class="var-strip-label">{{ t("prompts.sampleValuesLabel") }}</span>
            <div class="sample-grid">
              <label v-for="vn in detectedVarNames" :key="vn" class="sample-field">
                <span class="sample-label mono">{{ vn }}</span>
                <input
                  v-model="sampleValues[vn]"
                  class="inp sample-inp"
                  type="text"
                  spellcheck="false"
                  :placeholder="t('prompts.sampleValuePlaceholder')"
                />
              </label>
            </div>
            <p class="sample-hint muted small">{{ t("prompts.sampleValuesHint") }}</p>
          </div>

          <textarea
            v-if="showSchemaJson"
            v-model="draftSchema"
            class="area mono json-area"
            rows="10"
            spellcheck="false"
          />
          <p v-if="showSchemaJson && schemaJsonErr" class="err small">{{ schemaJsonErr }}</p>

          <div v-show="!showSchemaJson" class="editor-wrap">
            <div ref="lineNumsRef" class="line-nums" aria-hidden="true">
              <div v-for="n in lineCount" :key="n" class="ln">{{ n }}</div>
            </div>
            <div class="editor-stack" :class="{ 'has-debug-preview': hasDebugPreview }">
              <ul
                v-if="varSuggestOpen && varSuggestItems.length"
                class="var-suggest"
                role="listbox"
                :aria-label="t('prompts.insertVarToolbar')"
              >
                <li
                  v-for="(item, idx) in varSuggestItems"
                  :key="item.kind + item.name"
                  role="option"
                  :aria-selected="idx === varSuggestIndex"
                >
                  <button
                    type="button"
                    class="var-suggest-item"
                    :class="{ on: idx === varSuggestIndex, create: item.kind === 'create' }"
                    @mousedown.prevent="onVarSuggestMouseDown(idx)"
                  >
                    <span v-if="item.kind === 'create'" class="var-suggest-create">
                      {{ t("prompts.varSuggestCreate", { name: item.name }) }}
                    </span>
                    <span v-else class="mono">{{ item.name }}</span>
                  </button>
                </li>
              </ul>
              <textarea
                ref="bodyTextareaRef"
                v-model="editorLayoutBody"
                class="editor-layer body-area"
                rows="20"
                spellcheck="false"
                :disabled="isArchived"
                :aria-label="t('prompts.bodyLabel')"
                @scroll="syncEditorScroll"
                @keydown="onBodyKeydown"
                @beforeinput="bodyHistory.onBeforeInput"
                @compositionstart="bodyHistory.onCompositionStart"
                @compositionend="bodyHistory.onCompositionEnd"
                @input="refreshVarSuggest"
                @click="refreshVarSuggest"
                @keyup="onBodyKeyup"
              />
              <pre
                ref="bodyHighlightRef"
                class="editor-layer body-highlight"
                aria-hidden="true"
                v-html="bodyHighlightHtml"
              />
            </div>
          </div>
        </div>
      </div>

      <p
        v-if="validationIssues.length"
        class="val-inline"
        :class="{ error: hasErrors }"
        role="status"
      >
        {{ validationIssues[0].message }}
        <span v-if="validationIssues.length > 1" class="muted"> (+{{ validationIssues.length - 1 }})</span>
      </p>
    </section>

    <dialog
      ref="snapshotDlgEl"
      class="tool-dialog"
      aria-labelledby="prompt-snapshot-heading"
      @cancel="onSnapshotDlgCancel"
    >
      <div class="tool-dialog-surface">
        <header class="tool-dialog-head">
          <h2 id="prompt-snapshot-heading" class="tool-dialog-title">{{ t("prompts.toolbarSnapshot") }}</h2>
          <button
            type="button"
            class="tool-dialog-close"
            :disabled="snapBusy"
            :aria-label="t('common.close')"
            @click="closeSnapshotDlg"
          >
            ×
          </button>
        </header>
        <form class="tool-dialog-body snap" @submit.prevent="onSnapshot">
          <label class="lbl">{{ t("prompts.snapshotNote") }}</label>
          <input v-model="snapNote" class="inp full" :disabled="isArchived" />
          <label class="lbl">{{ t("prompts.snapshotVersionLabel") }}</label>
          <input
            v-model="snapCustomVersion"
            class="inp full mono"
            type="text"
            inputmode="decimal"
            :placeholder="suggestedNextVersion"
            :disabled="isArchived"
          />
          <p class="hint small muted">
            {{ t("prompts.snapshotVersionHint", { next: formatVersionDisplay(suggestedNextVersion) }) }}
          </p>
          <footer class="tool-dialog-foot">
            <button type="button" class="btn ghost sm" :disabled="snapBusy" @click="closeSnapshotDlg">
              {{ t("common.close") }}
            </button>
            <button type="submit" class="btn sm" :disabled="snapBusy || isArchived">
              {{ t("prompts.snapshotCta") }}
            </button>
          </footer>
        </form>
      </div>
    </dialog>

    <dialog
      ref="publishDlgEl"
      class="tool-dialog tool-dialog-wide"
      aria-labelledby="prompt-publish-heading"
      @cancel="onPublishDlgCancel"
    >
      <div class="tool-dialog-surface">
        <header class="tool-dialog-head">
          <h2 id="prompt-publish-heading" class="tool-dialog-title">{{ t("prompts.toolbarPublish") }}</h2>
          <button
            type="button"
            class="tool-dialog-close"
            :disabled="pointerBusy"
            :aria-label="t('common.close')"
            @click="closePublishDlg"
          >
            ×
          </button>
        </header>
        <div class="tool-dialog-body publish">
          <p class="channel-line small muted">
            {{ t("prompts.channelLabel") }}: <span class="mono">{{ publishChannel }}</span>
          </p>
          <p v-if="pointerLoading" class="muted small">{{ t("prompts.loading") }}</p>
          <p v-else-if="pointer" class="small live-line">
            {{ t("prompts.currentPointer") }}:
            <span class="mono">{{ formatVersionDisplay(pointer.version) }}</span>
          </p>
          <p v-else class="muted small">{{ t("prompts.noPointer") }}</p>

          <label class="lbl" for="inkforge-publish-ver-search">{{ t("prompts.publishVersionLabel") }}</label>
          <input
            id="inkforge-publish-ver-search"
            v-model="publishVersionSearch"
            class="inp full search"
            type="search"
            :placeholder="t('prompts.publishVersionSearchPlaceholder')"
            :disabled="isArchived"
          />
          <p v-if="verLoading" class="muted small">{{ t("prompts.loading") }}</p>
          <p v-else-if="!publishVersionOptions.length" class="muted small">
            {{ t("prompts.publishVersionEmpty") }}
          </p>
          <ul v-else class="ver-pick-list" role="listbox" :aria-label="t('prompts.publishVersionLabel')">
            <li v-for="v in publishVersionOptions" :key="v.id" role="option">
              <button
                type="button"
                class="ver-pick-item"
                :class="{ on: pointerVersionId === v.id }"
                :aria-selected="pointerVersionId === v.id"
                @click="pointerVersionId = v.id"
              >
                <span class="mono ver-pick-num">{{ formatVersionDisplay(v.version) }}</span>
                <span class="ver-pick-meta muted small">{{ v.created_at }}</span>
                <span v-if="v.change_note" class="ver-pick-note small">{{ v.change_note }}</span>
              </button>
            </li>
          </ul>
          <footer class="tool-dialog-foot">
            <button type="button" class="btn ghost sm" :disabled="pointerBusy" @click="closePublishDlg">
              {{ t("common.close") }}
            </button>
            <button
              type="button"
              class="btn sm"
              :disabled="pointerBusy || isArchived || !pointerVersionId"
              @click="onPublishPointer"
            >
              {{ t("prompts.applyPointer") }}
            </button>
          </footer>
        </div>
      </div>
    </dialog>

    <dialog
      ref="versionsDlgEl"
      class="tool-dialog tool-dialog-wide"
      aria-labelledby="prompt-versions-heading"
    >
      <div class="tool-dialog-surface">
        <header class="tool-dialog-head">
          <h2 id="prompt-versions-heading" class="tool-dialog-title">{{ t("prompts.toolbarVersions") }}</h2>
          <button
            type="button"
            class="tool-dialog-close"
            :aria-label="t('common.close')"
            @click="closeVersionsDlg"
          >
            ×
          </button>
        </header>
        <div class="tool-dialog-body">
          <p v-if="verLoading" class="muted">{{ t("prompts.loading") }}</p>
          <p v-else-if="!versions.length" class="muted small">{{ t("prompts.historyEmpty") }}</p>
          <template v-else>
            <div v-if="versions.length >= 2" class="diff-panel card">
              <div class="diff-pickers">
                <label class="diff-pick">
                  <span class="diff-pick-label">{{ t("prompts.diffBase") }}</span>
                  <select v-model="diffBaseId" class="diff-select mono">
                    <option v-for="v in versionsNewestFirst" :key="'base-' + v.id" :value="v.id">
                      {{ formatVersionDisplay(v.version) }}
                    </option>
                  </select>
                </label>
                <span class="diff-arrow" aria-hidden="true">→</span>
                <label class="diff-pick">
                  <span class="diff-pick-label">{{ t("prompts.diffCompare") }}</span>
                  <select v-model="diffCompareId" class="diff-select mono">
                    <option v-for="v in versionsNewestFirst" :key="'cmp-' + v.id" :value="v.id">
                      {{ formatVersionDisplay(v.version) }}
                    </option>
                  </select>
                </label>
              </div>
              <p class="diff-hint muted small">{{ t("prompts.diffPickHint") }}</p>
              <p v-if="diffBusy" class="muted small">{{ t("prompts.diffRunning") }}</p>
              <p v-else-if="diffErr" class="err small">{{ diffErr }}</p>
              <p v-else-if="diffBaseId === diffCompareId" class="muted small">
                {{ t("prompts.diffSameVersion") }}
              </p>
              <pre v-else-if="diffOut" class="pre">{{ diffOut.body_diff || "(no body diff)" }}

--- schema ---
{{ diffOut.schema_diff || "(no schema diff)" }}</pre>
            </div>

            <ul class="vlist ver-history">
              <li
                v-for="(v, idx) in versions"
                :key="v.id"
                class="vrow card ver-history-row"
                :class="{
                  'is-base': diffBaseId === v.id,
                  'is-compare': diffCompareId === v.id,
                }"
              >
                <div class="ver-history-main">
                  <span class="mono ver-history-label">{{ formatVersionDisplay(v.version) }}</span>
                  <span class="small">{{ v.created_at }}</span>
                  <span v-if="v.change_note" class="note">{{ v.change_note }}</span>
                </div>
                <div class="ver-history-actions">
                  <button
                    type="button"
                    class="diff-slot-btn"
                    :class="{ on: diffBaseId === v.id }"
                    @click="setDiffBase(v.id)"
                  >
                    {{ t("prompts.diffBaseShort") }}
                  </button>
                  <button
                    type="button"
                    class="diff-slot-btn"
                    :class="{ on: diffCompareId === v.id }"
                    @click="setDiffCompare(v.id)"
                  >
                    {{ t("prompts.diffCompareShort") }}
                  </button>
                  <button
                    v-if="idx > 0"
                    type="button"
                    class="diff-quick-btn"
                    @click="compareWithPrevious(idx)"
                  >
                    {{ t("prompts.diffVsPrev") }}
                  </button>
                </div>
              </li>
            </ul>
          </template>

          <footer class="tool-dialog-foot">
            <button type="button" class="btn ghost sm" @click="closeVersionsDlg">
              {{ t("common.close") }}
            </button>
          </footer>
        </div>
      </div>
    </dialog>

    <!-- Unsaved-leave dialog -->
    <dialog
      ref="leaveDlgRef"
      class="unsaved-dlg"
      role="alertdialog"
      aria-labelledby="inkforge-unsaved-title"
      @cancel.prevent="leaveDlgClose"
    >
      <div class="unsaved-surface">
        <header class="unsaved-head">
          <h3 id="inkforge-unsaved-title" class="unsaved-title">{{ t("prompts.unsavedTitle") }}</h3>
          <button
            type="button"
            class="unsaved-close"
            :disabled="leaveBusy || saveBusy"
            :aria-label="t('common.close')"
            @click="leaveDlgClose"
          >
            ×
          </button>
        </header>
        <p class="unsaved-msg">{{ t("prompts.unsavedMessage") }}</p>
        <footer class="unsaved-foot">
          <button
            type="button"
            class="unsaved-btn ghost"
            :disabled="leaveBusy || saveBusy"
            @click="leaveDlgClose"
          >
            {{ t("prompts.unsavedStay") }}
          </button>
          <button
            type="button"
            class="unsaved-btn warn"
            :disabled="leaveBusy || saveBusy"
            @click="confirmDiscardAndExit"
          >
            {{ t("prompts.unsavedDiscard") }}
          </button>
          <button
            type="button"
            class="unsaved-btn primary"
            :disabled="leaveBusy || saveBusy || isArchived"
            @click="confirmSaveAndExit"
          >
            {{
              leaveBusy || saveBusy ? t("prompts.unsavedSaving") : t("prompts.unsavedSaveExit")
            }}
          </button>
        </footer>
      </div>
    </dialog>
  </div>
</template>

<style scoped>
.detail {
  max-width: 72rem;
}
.back {
  font: inherit;
  font-weight: 650;
  background: none;
  border: none;
  color: var(--accent);
  cursor: pointer;
  padding: 0 0 0.75rem;
}
.head {
  margin-bottom: 1rem;
}
.title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 800;
}
.sub {
  margin: 0.35rem 0 0;
  color: var(--fg-muted);
}
.warn {
  padding: 0.75rem 1rem;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, #f97316 35%, var(--border-subtle));
  background: color-mix(in srgb, #f97316 8%, var(--elev-2));
  font-size: 0.9rem;
  color: var(--fg-muted);
}

/* ── Work toolbar ───────────────────────────────────────────────── */
.work-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.55rem 0.75rem;
  margin-bottom: 0.85rem;
  padding: 0.55rem 0.75rem;
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 88%, transparent);
}
.toolbar-draft,
.toolbar-release {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.55rem;
}
.toolbar-spacer {
  flex: 1 1 0.5rem;
  min-width: 0.25rem;
}
.live-badge {
  font-size: 0.72rem;
  font-weight: 650;
  padding: 0.18rem 0.45rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, #4ade80 35%, var(--border-subtle));
  background: color-mix(in srgb, #4ade80 10%, var(--elev-2));
  color: var(--fg-muted);
}
.btn.tool {
  background: var(--elev-2);
  border-color: var(--border-strong);
  font-weight: 600;
}
.btn.tool:hover:not(:disabled) {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 10%, var(--elev-2));
}

.tool-dialog {
  padding: 0;
  border: none;
  border-radius: 14px;
  background: var(--elev);
  color: var(--fg);
  max-width: min(440px, calc(100vw - 2rem));
  width: 100%;
  box-shadow:
    0 0 0 1px var(--border-subtle),
    0 24px 64px rgba(0, 0, 0, 0.42);
}
.tool-dialog-wide {
  max-width: min(520px, calc(100vw - 2rem));
}
.tool-dialog::backdrop {
  background: rgba(0, 0, 0, 0.5);
}
.tool-dialog-surface {
  overflow: hidden;
}
.tool-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}
.tool-dialog-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
}
.tool-dialog-close {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  margin: -0.2rem -0.15rem 0 0;
  padding: 0;
  border: none;
  border-radius: 8px;
  font-size: 1.35rem;
  line-height: 1;
  color: var(--fg-soft);
  background: transparent;
  cursor: pointer;
}
.tool-dialog-close:hover:not(:disabled) {
  background: var(--elev-2);
  color: var(--fg);
}
.tool-dialog-close:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.tool-dialog-body {
  padding: 1rem 1.1rem;
  display: grid;
  gap: 0.45rem;
  max-height: min(70vh, 640px);
  overflow: auto;
}
.tool-dialog-foot {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.35rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--border-faint);
}
.vlist.compact {
  max-height: 12rem;
  overflow: auto;
}

/* ── Edit toolbar (legacy aliases) ──────────────────────────────── */
.edit-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.85rem;
}
.diff-summary {
  cursor: pointer;
  font-weight: 650;
  color: var(--fg);
  user-select: none;
}
.diff-summary::-webkit-details-marker {
  color: var(--accent);
}
.live-line {
  margin: 0;
}
.inp.full {
  width: 100%;
  max-width: none;
  box-sizing: border-box;
}
.save-state {
  font-size: 0.75rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--fg-muted);
}
.save-msg {
  color: var(--fg-muted);
}
.save-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--fg-soft);
}
.saved-dot {
  background: #4ade80;
}
.dirty-dot {
  background: #f97316;
  box-shadow: 0 0 6px color-mix(in srgb, #f97316 50%, transparent);
}
.ok {
  color: #4ade80;
}

/* ── Edit layout ────────────────────────────────────────────────── */
.edit-layout {
  display: block;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.45rem;
}
.link-btn {
  font: inherit;
  font-size: 0.72rem;
  color: var(--fg-soft);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.link-btn:hover {
  color: var(--accent);
}
.col-title {
  margin: 0;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--fg-soft);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

/* ── Auto-detected variables ────────────────────────────────────── */
.var-strip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem 0.55rem;
  margin-bottom: 0.5rem;
  min-height: 1.5rem;
}
.var-strip-label {
  font-size: 0.68rem;
  font-weight: 650;
  color: var(--fg-soft);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.var-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}
.var-chip {
  font-size: 0.72rem;
  padding: 0.15rem 0.45rem;
  cursor: pointer;
}
.var-empty-hint {
  font-size: 0.78rem;
  color: var(--fg-soft);
}
.var-add-btn {
  margin-left: auto;
  font-size: 0.72rem;
}
.var-picker {
  display: grid;
  gap: 0.35rem;
  margin-bottom: 0.55rem;
  padding: 0.55rem 0.65rem;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--accent) 28%, var(--border-subtle));
  background: color-mix(in srgb, var(--accent) 6%, var(--elev-2));
}
.var-picker-label {
  font-size: 0.68rem;
  font-weight: 650;
  color: var(--fg-soft);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.var-picker-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
}
.var-picker-inp {
  flex: 1 1 8rem;
  min-width: 0;
  font-size: 0.82rem;
}
.var-picker-err {
  margin: 0;
  color: color-mix(in srgb, #f87171 70%, var(--fg));
}
.var-picker-hint {
  margin: 0;
}
.var-suggest {
  position: absolute;
  top: 0.45rem;
  left: 0.55rem;
  right: 0.55rem;
  z-index: 4;
  list-style: none;
  margin: 0;
  padding: 0.28rem;
  display: grid;
  gap: 0.12rem;
  max-height: 10rem;
  overflow: auto;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  background: var(--elev);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.28);
}
.var-suggest-item {
  width: 100%;
  text-align: left;
  font: inherit;
  font-size: 0.82rem;
  padding: 0.38rem 0.55rem;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--fg);
  cursor: pointer;
}
.var-suggest-item:hover,
.var-suggest-item.on {
  background: color-mix(in srgb, var(--accent) 14%, var(--elev-2));
}
.var-suggest-item.create {
  color: var(--accent);
  font-weight: 600;
}
.var-suggest-create {
  font-size: 0.82rem;
}
.sample-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.45rem 0.65rem;
  margin-bottom: 0.55rem;
  padding: 0.55rem 0.65rem;
  border-radius: 10px;
  border: 1px dashed var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 65%, transparent);
}
.sample-hint {
  flex: 1 1 100%;
  margin: 0;
  line-height: 1.45;
}
.json-area {
  width: 100%;
  min-height: 12rem;
  margin-bottom: 0.5rem;
}
.chip {
  border-radius: 6px;
  border: 1px solid var(--border-subtle);
  background: var(--elev-2);
  color: var(--fg-muted);
}
.var-chip:hover {
  border-color: var(--accent-soft);
  color: var(--fg);
}

/* ── Editor with line numbers ───────────────────────────────────── */
.editor-wrap {
  display: flex;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  overflow: hidden;
  background: var(--elev-2);
}
.editor-stack {
  position: relative;
  flex: 1;
  min-width: 0;
  min-height: 24rem;
}
.editor-layer {
  margin: 0;
  padding: 0.55rem 0.65rem;
  border: none;
  box-sizing: border-box;
  width: 100%;
  min-height: 24rem;
  line-height: 1.5;
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.875rem;
  font-weight: 400;
  font-variant-ligatures: none;
  letter-spacing: normal;
  tab-size: 2;
  white-space: pre-wrap;
  overflow-wrap: break-word;
  word-break: normal;
}
.body-highlight {
  position: absolute;
  inset: 0;
  z-index: 2;
  height: 100%;
  overflow: auto;
  pointer-events: none;
  color: transparent;
  background: transparent;
  scrollbar-width: none;
}
.body-highlight::-webkit-scrollbar {
  display: none;
}
.line-nums {
  flex-shrink: 0;
  width: 2.75rem;
  overflow: hidden;
  background: color-mix(in srgb, var(--bg-canvas) 70%, var(--elev-2));
  border-right: 1px solid var(--border-faint);
  padding: 0.55rem 0.35rem;
  text-align: right;
  user-select: none;
}
.ln {
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.8rem;
  line-height: 1.5;
  color: var(--fg-soft);
}
.body-area {
  position: relative;
  z-index: 1;
  height: 100%;
  flex: 1;
  border-radius: 0;
  background: transparent;
  color: var(--fg);
  caret-color: var(--fg);
  resize: none;
}
.body-area:disabled {
  color: var(--fg-muted);
  opacity: 1;
}
.body-area:focus {
  outline: none;
}
.editor-wrap:focus-within {
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 35%, transparent);
}

/* ── Sample values ─────────────────────────────────────────────── */
.sample-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 0.65rem;
  flex: 1 1 12rem;
  min-width: 0;
}
.sample-field {
  display: grid;
  grid-template-columns: auto minmax(5rem, 10rem);
  align-items: center;
  gap: 0.35rem;
}
.sample-label {
  font-size: 0.72rem;
  color: var(--fg-soft);
  max-width: 7rem;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sample-inp {
  font-size: 0.78rem;
  padding: 0.28rem 0.45rem;
  min-width: 0;
}
.body-highlight :deep(.hl-plain) {
  color: transparent;
}
.editor-stack.has-debug-preview .body-area {
  color: transparent;
  caret-color: var(--fg);
}
.editor-stack.has-debug-preview .body-highlight {
  color: var(--fg);
}
.editor-stack.has-debug-preview .body-highlight :deep(.hl-plain) {
  color: var(--fg);
}
.body-highlight :deep(.var-ph-tag) {
  display: inline;
  border-radius: 4px;
  padding: 0 0.1em;
  background: color-mix(in srgb, var(--accent) 22%, transparent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 38%, transparent);
}
.body-highlight :deep(.p-open),
.body-highlight :deep(.p-close) {
  color: inherit;
}
.body-highlight :deep(.p-name) {
  color: var(--fg);
  font-weight: 600;
}
.body-highlight :deep(.p-val) {
  color: color-mix(in srgb, var(--fg-muted) 78%, transparent);
  opacity: 0.62;
  font-size: 0.84em;
  font-style: italic;
  font-weight: 400;
  letter-spacing: 0.01em;
}
.body-highlight :deep(.var-ph) {
  display: inline;
  margin: 0;
  padding: 0;
  color: transparent;
  background: color-mix(in srgb, var(--accent) 26%, transparent);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 42%, transparent);
  border-radius: 4px;
  font: inherit;
}
.val-inline {
  margin: 0.65rem 0 0;
  font-size: 0.78rem;
  color: color-mix(in srgb, #f97316 65%, var(--fg-muted));
}
.val-inline.error {
  color: color-mix(in srgb, #f87171 70%, var(--fg));
}

/* ── Validation bar ─────────────────────────────────────────────── */
.val-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.55rem 0.75rem;
  border-radius: 10px;
  border: 1px solid var(--border-subtle);
  font-size: 0.78rem;
  line-height: 1.45;
}
.val-bar.ok {
  border-color: color-mix(in srgb, #4ade80 30%, var(--border-subtle));
  background: color-mix(in srgb, #4ade80 6%, var(--elev-2));
  color: color-mix(in srgb, #4ade80 70%, var(--fg));
}
.val-bar.warn {
  border-color: color-mix(in srgb, #f97316 30%, var(--border-subtle));
  background: color-mix(in srgb, #f97316 6%, var(--elev-2));
  color: color-mix(in srgb, #f97316 70%, var(--fg));
}
.val-bar.error {
  border-color: color-mix(in srgb, #f87171 30%, var(--border-subtle));
  background: color-mix(in srgb, #f87171 6%, var(--elev-2));
  color: color-mix(in srgb, #f87171 70%, var(--fg));
}
.val-icon {
  flex-shrink: 0;
  font-weight: 700;
  font-size: 0.85rem;
  padding-top: 0.05rem;
}
.val-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.val-list li {
  margin: 0;
}
.val-list li.warn {
  color: color-mix(in srgb, #f97316 70%, var(--fg));
}
.val-list li.error {
  color: color-mix(in srgb, #f87171 70%, var(--fg));
}
.val-msg {
  color: inherit;
}

/* ── Versions tab (unchanged) ───────────────────────────────────── */
.card {
  padding: 1rem 1.1rem;
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 88%, transparent);
  margin-bottom: 1rem;
}
.snap {
  display: grid;
  gap: 0.5rem;
}
.vlist {
  list-style: none;
  margin: 0 0 1rem;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}
.vrow {
  display: grid;
  gap: 0.2rem;
  margin: 0;
  padding: 0.65rem 0.85rem;
}
.ver-pick-list {
  list-style: none;
  margin: 0 0 0.85rem;
  padding: 0;
  max-height: 14rem;
  overflow: auto;
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  background: color-mix(in srgb, var(--elev-2) 70%, transparent);
}
.ver-pick-item {
  width: 100%;
  display: grid;
  gap: 0.1rem;
  text-align: left;
  padding: 0.55rem 0.7rem;
  border: none;
  border-bottom: 1px solid var(--border-faint);
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}
.ver-pick-list li:last-child .ver-pick-item {
  border-bottom: none;
}
.ver-pick-item:hover {
  background: var(--pill-hover);
}
.ver-pick-item.on {
  background: color-mix(in srgb, var(--accent) 14%, var(--elev-2));
  box-shadow: inset 3px 0 0 var(--accent);
}
.ver-pick-num {
  font-weight: 650;
  font-size: 0.88rem;
}
.channel-line {
  margin: 0 0 0.75rem;
}
.note {
  font-size: 0.82rem;
  color: var(--fg-muted);
}
.diff-panel {
  display: grid;
  gap: 0.55rem;
  margin-bottom: 0.85rem;
}
.diff-pickers {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.5rem 0.65rem;
}
.diff-pick {
  display: grid;
  gap: 0.28rem;
  flex: 1 1 8rem;
  min-width: 0;
}
.diff-pick-label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-soft);
}
.diff-select {
  width: 100%;
  font: inherit;
  font-size: 0.88rem;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  padding: 0.42rem 0.55rem;
}
.diff-arrow {
  color: var(--fg-soft);
  font-size: 1.1rem;
  line-height: 1;
  padding-bottom: 0.45rem;
}
.diff-hint {
  margin: 0;
}
.ver-history {
  max-height: 14rem;
  overflow: auto;
}
.ver-history-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem 0.65rem;
  padding: 0.55rem 0.7rem;
  border-left: 3px solid transparent;
}
.ver-history-row.is-base {
  border-left-color: color-mix(in srgb, #60a5fa 70%, transparent);
}
.ver-history-row.is-compare {
  border-left-color: color-mix(in srgb, #4ade80 70%, transparent);
}
.ver-history-row.is-base.is-compare {
  border-left-color: color-mix(in srgb, var(--accent) 70%, transparent);
}
.ver-history-main {
  display: grid;
  gap: 0.12rem;
  min-width: 0;
  flex: 1 1 10rem;
}
.ver-history-label {
  font-weight: 650;
}
.ver-history-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  flex-shrink: 0;
}
.diff-slot-btn,
.diff-quick-btn {
  font: inherit;
  font-size: 0.68rem;
  font-weight: 650;
  cursor: pointer;
  border-radius: 999px;
  padding: 0.18rem 0.48rem;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg-muted);
}
.diff-slot-btn.on {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 14%, var(--elev-2));
  color: var(--fg);
}
.diff-quick-btn {
  border-style: dashed;
}
.diff-quick-btn:hover,
.diff-slot-btn:hover:not(.on) {
  border-color: var(--accent-soft);
  color: var(--fg);
}
.diffcard .h2 {
  margin: 0 0 0.5rem;
  font-size: 1rem;
}
.diffnums {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.65rem;
}
.pre {
  margin: 0;
  padding: 0.75rem;
  border-radius: 10px;
  background: var(--bg-canvas);
  border: 1px solid var(--border-faint);
  font-size: 0.72rem;
  overflow: auto;
  white-space: pre-wrap;
}
.hint {
  font-size: 0.75rem;
  color: var(--fg-muted);
  margin: 0.25rem 0 0.75rem;
}

/* ── Shared form elements ───────────────────────────────────────── */
.lbl {
  display: block;
  font-size: 0.72rem;
  font-weight: 650;
  color: var(--fg-soft);
  margin: 0.5rem 0 0.35rem;
}
.lbl.inline {
  display: inline;
  margin: 0 0.35rem 0 0;
}
.inp {
  font: inherit;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  padding: 0.45rem 0.65rem;
  width: min(100%, 28rem);
}
.inp.num {
  width: 5rem;
}
.area {
  width: 100%;
  font: inherit;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  padding: 0.55rem 0.65rem;
  resize: vertical;
  box-sizing: border-box;
}
.btn {
  font: inherit;
  font-weight: 650;
  cursor: pointer;
  border-radius: 10px;
  padding: 0.45rem 0.85rem;
  border: 1px solid var(--border-strong);
  background: color-mix(in srgb, var(--accent) 14%, var(--elev-2));
  color: var(--fg);
}
.btn:hover:not(:disabled) {
  border-color: var(--accent-soft);
}
.btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.btn.ghost {
  background: transparent;
  border-color: var(--border-subtle);
  font-weight: 500;
}
.btn.ghost:hover:not(:disabled) {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}
.btn.sm {
  font-size: 0.75rem;
  padding: 0.3rem 0.55rem;
}
.chip {
  margin: 0;
  padding: 0.2rem 0.45rem;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  background: color-mix(in srgb, var(--accent) 14%, var(--elev-2));
  border: 1px solid color-mix(in srgb, var(--accent) 30%, var(--border-subtle));
  color: var(--fg);
}
.small {
  font-size: 0.78rem;
}
.muted {
  color: var(--fg-muted);
}
.err {
  color: color-mix(in srgb, #f97316 65%, var(--fg));
  font-size: 0.85rem;
  margin: 0.35rem 0 0;
}
.mono {
  font-family: "JetBrains Mono", ui-monospace, monospace;
}

/* ── Unsaved dialog (unchanged) ─────────────────────────────────── */
.unsaved-dlg {
  padding: 0;
  border: none;
  border-radius: 14px;
  background: var(--elev);
  color: var(--fg);
  max-width: min(460px, calc(100vw - 2rem));
  width: 100%;
  box-shadow:
    0 0 0 1px var(--border-subtle),
    0 24px 64px rgba(0, 0, 0, 0.42);
}
.unsaved-dlg::backdrop {
  background: rgba(0, 0, 0, 0.5);
}
.unsaved-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.65rem;
  border-bottom: 1px solid var(--border-subtle);
}
.unsaved-title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
}
.unsaved-close {
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  border: none;
  border-radius: 8px;
  font-size: 1.35rem;
  line-height: 1;
  color: var(--fg-soft);
  background: transparent;
  cursor: pointer;
}
.unsaved-close:hover:not(:disabled) {
  background: var(--elev-2);
  color: var(--fg);
}
.unsaved-close:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.unsaved-msg {
  margin: 0;
  padding: 1rem 1.1rem;
  font-size: 0.875rem;
  line-height: 1.55;
  color: var(--fg-muted);
}
.unsaved-foot {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0 1.1rem 1.1rem;
}
.unsaved-btn {
  font: inherit;
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  border-radius: 10px;
  padding: 0.5rem 0.9rem;
  border: 1px solid var(--border-strong);
}
.unsaved-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.unsaved-btn.ghost {
  background: var(--elev-2);
  color: var(--fg-muted);
}
.unsaved-btn.warn {
  background: color-mix(in srgb, #f97316 12%, var(--elev-2));
  border-color: color-mix(in srgb, #f97316 35%, var(--border-strong));
  color: var(--fg);
}
.unsaved-btn.primary {
  background: linear-gradient(145deg, #818cf8 0%, #6366f1 52%, #4f46e5 100%);
  color: #fafafa;
  border-color: color-mix(in srgb, #fff 42%, transparent);
}
</style>