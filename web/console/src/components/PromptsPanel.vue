<script setup lang="ts">
import {
  createPrompt,
  listPrompts,
  type PromptSummary,
} from "@/api/prompts";
import { ApiError } from "@/api/client";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

const { t } = useI18n();
const ws = useWorkspaceContextStore();
const router = useRouter();

const items = ref<PromptSummary[]>([]);
const total = ref(0);
const loading = ref(false);
const error = ref<string | null>(null);
const q = ref("");
const newTitle = ref("");
const createBusy = ref(false);
const createErr = ref<string | null>(null);
const createDialogEl = ref<HTMLDialogElement | null>(null);
const createTitleInputEl = ref<HTMLInputElement | null>(null);

const ns = computed(() => ws.selectedNsSlug);
const isArchived = computed(() => ws.selectedNamespace?.status === "archived");

function promptLabel(p: PromptSummary): string {
  return p.title?.trim() || p.prompt_key;
}

async function load() {
  const slug = ns.value;
  if (!slug) {
    items.value = [];
    total.value = 0;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    const res = await listPrompts(slug, { q: q.value.trim() || undefined });
    items.value = res.items ?? [];
    total.value = res.total ?? 0;
  } catch (e) {
    items.value = [];
    total.value = 0;
    error.value =
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.loadError");
  } finally {
    loading.value = false;
  }
}

function openPrompt(key: string) {
  void router.push({
    name: "console-prompt-detail",
    params: { promptKey: key },
  });
}

function openCreateDialog() {
  createErr.value = null;
  newTitle.value = "";
  createDialogEl.value?.showModal();
  void nextTick(() => createTitleInputEl.value?.focus());
}

function closeCreateDialog() {
  createDialogEl.value?.close();
}

function onCreateDialogCancel(ev: Event) {
  if (createBusy.value) {
    ev.preventDefault();
  }
}

async function submitCreate() {
  const slug = ns.value;
  const title = newTitle.value.trim();
  if (!slug || title.length === 0) return;
  createErr.value = null;
  createBusy.value = true;
  try {
    const created = await createPrompt(slug, { title });
    newTitle.value = "";
    closeCreateDialog();
    await load();
    if (created?.prompt_key) openPrompt(created.prompt_key);
  } catch (e) {
    createErr.value =
      e instanceof ApiError ? `${e.code}: ${e.message}` : t("prompts.createError");
  } finally {
    createBusy.value = false;
  }
}

onMounted(() => void load());
watch(ns, () => void load());

let searchTimer: ReturnType<typeof setTimeout> | null = null;
watch(q, () => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => void load(), 280);
});
</script>

<template>
  <section class="prompts-panel" :aria-label="t('prompts.title')">
    <div class="toolbar" role="toolbar" :aria-label="t('prompts.toolbarAria')">
      <input
        v-model="q"
        class="inp search"
        type="search"
        :placeholder="t('prompts.searchPlaceholder')"
        :disabled="!ns"
      />
      <span v-if="ns" class="count">{{ total }} · {{ t("prompts.total") }}</span>
      <button
        v-if="ns && !isArchived"
        type="button"
        class="btn-primary btn-create"
        @click="openCreateDialog"
      >
        {{ t("prompts.createOpenCta") }}
      </button>
    </div>

    <p v-if="!ns" class="warn">{{ t("prompts.pickNs") }}</p>
    <p v-else-if="isArchived" class="warn warn-compact">{{ t("prompts.archivedReadOnly") }}</p>

    <dialog
      ref="createDialogEl"
      class="prompt-create-dialog"
      aria-labelledby="prompt-create-heading"
      @cancel="onCreateDialogCancel"
    >
      <div class="create-dialog-surface">
        <header class="create-dialog-head">
          <h2 id="prompt-create-heading" class="panel-title">{{ t("prompts.createDialogTitle") }}</h2>
          <button
            type="button"
            class="create-dialog-close"
            :disabled="createBusy"
            :aria-label="t('common.close')"
            @click="closeCreateDialog"
          >
            ×
          </button>
        </header>
        <form class="create-form" @submit.prevent="submitCreate">
          <label class="fld span-2" for="inkforge-new-prompt-title">
            <span>{{ t("prompts.newTitleLabel") }}</span>
            <input
              id="inkforge-new-prompt-title"
              ref="createTitleInputEl"
              v-model="newTitle"
              type="text"
              autocomplete="off"
              :placeholder="t('prompts.newTitlePlaceholder')"
            />
          </label>
          <div class="create-dialog-footer">
            <button type="button" class="btn-dialog-muted" :disabled="createBusy" @click="closeCreateDialog">
              {{ t("common.close") }}
            </button>
            <button type="submit" class="btn-primary" :disabled="createBusy || newTitle.trim().length === 0">
              {{ createBusy ? t("prompts.loading") : t("prompts.createCta") }}
            </button>
          </div>
          <p v-if="createErr" class="form-err">{{ createErr }}</p>
        </form>
      </div>
    </dialog>

    <p v-if="error" class="err">{{ error }}</p>
    <p v-if="loading" class="muted">{{ t("prompts.loading") }}</p>

    <ul v-else-if="ns" class="list">
      <li v-for="p in items" :key="p.prompt_key" class="row">
        <button type="button" class="row-main" @click="openPrompt(p.prompt_key)">
          <span class="key">{{ promptLabel(p) }}</span>
          <span class="muted small">{{ p.updated_at }}</span>
        </button>
      </li>
      <li v-if="!loading && items.length === 0" class="empty muted">{{ t("prompts.emptyList") }}</li>
    </ul>
  </section>
</template>

<style scoped>
.prompts-panel {
  margin-top: 0.5rem;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.55rem 0.75rem;
  margin-bottom: 0.85rem;
}

.inp {
  font: inherit;
  border-radius: 10px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  padding: 0.45rem 0.65rem;
  min-width: 12rem;
}

.inp.search {
  flex: 1;
  min-width: min(100%, 200px);
}

.count {
  font-size: 0.8rem;
  color: var(--fg-muted);
  white-space: nowrap;
}

.btn-create {
  margin-left: auto;
  cursor: pointer;
}

.warn {
  padding: 0.65rem 0.85rem;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, #f97316 35%, var(--border-subtle));
  background: color-mix(in srgb, #f97316 8%, var(--elev-2));
  color: var(--fg-muted);
  font-size: 0.875rem;
  margin: 0 0 0.85rem;
}

.warn-compact {
  padding: 0.45rem 0.7rem;
  font-size: 0.8125rem;
}

.prompt-create-dialog {
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

.prompt-create-dialog::backdrop {
  background: rgba(0, 0, 0, 0.5);
}

.create-dialog-surface {
  overflow: hidden;
}

.create-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}

.create-dialog-head .panel-title {
  margin: 0;
  padding-right: 0.35rem;
  line-height: 1.35;
  font-size: 1rem;
  font-weight: 700;
}

.create-dialog-close {
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

.create-dialog-close:hover:not(:disabled) {
  background: var(--elev-2);
  color: var(--fg);
}

.create-dialog-close:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.prompt-create-dialog .create-form {
  padding: 1rem 1.1rem 1.15rem;
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.fld {
  display: flex;
  flex-direction: column;
  gap: 0.38rem;
  font-size: 0.6825rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-soft);
}

.fld input {
  padding: 0.45rem 0.62rem;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  font-size: 0.9rem;
  font-weight: 400;
  text-transform: none;
  letter-spacing: 0;
}

.create-dialog-footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
  gap: 0.55rem;
}

.btn-dialog-muted {
  cursor: pointer;
  border-radius: 10px;
  padding: 0.5rem 1rem;
  font-weight: 600;
  font-size: 0.8575rem;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg-muted);
}

.btn-dialog-muted:hover:not(:disabled) {
  color: var(--fg);
}

.btn-dialog-muted:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.45rem 0.9rem;
  border-radius: 10px;
  font-weight: 650;
  font-size: 0.8575rem;
  border: none;
  background: linear-gradient(145deg, #818cf8 0%, #6366f1 52%, #4f46e5 100%);
  color: #fff;
  cursor: pointer;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.06);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-err {
  margin: 0;
  padding: 0.45rem;
  border-radius: 8px;
  border: 1px solid color-mix(in srgb, #f97316 30%, transparent);
  background: color-mix(in srgb, #f97316 12%, transparent);
  color: var(--fg-muted);
  font-size: 0.8125rem;
}

.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.row {
  margin: 0;
  padding: 0;
}

.row-main {
  width: 100%;
  text-align: left;
  display: grid;
  gap: 0.15rem;
  padding: 0.85rem 1rem;
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 88%, transparent);
  cursor: pointer;
  color: inherit;
  font: inherit;
}

.row-main:hover {
  border-color: var(--accent-soft);
}

.key {
  font-weight: 650;
  font-size: 0.92rem;
}

.small {
  font-size: 0.72rem;
}

.muted {
  color: var(--fg-muted);
}

.empty {
  padding: 1.25rem 0.5rem;
  text-align: center;
  font-size: 0.875rem;
}

.err {
  color: color-mix(in srgb, #f97316 65%, var(--fg));
  font-size: 0.85rem;
  margin: 0.35rem 0 0;
}
</style>
