<script setup lang="ts">
import { ApiError } from "@/api/client";
import type { NamespaceDetail } from "@/api/namespaces";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const ctx = useWorkspaceContextStore();

const selected = computed(() => ctx.selectedNamespace);

function quotaLine(ns: NamespaceDetail): string {
  if (ns.quota_prompts_max > 0) {
    return t("workspace.nsQuotaUsed", {
      count: ns.prompt_count,
      cap: ns.quota_prompts_max,
    });
  }
  return t("workspace.nsQuotaUnlimited", { count: ns.prompt_count });
}

const createSlug = ref("");
const createDisplayName = ref("");
const createDesc = ref("");
const createTags = ref("");
const createQuotaStr = ref("");
const createBusy = ref(false);
const createErr = ref<string | null>(null);
const createDialogEl = ref<HTMLDialogElement | null>(null);

const archiveDialogOpen = ref(false);
const restoreDialogOpen = ref(false);
const confirmBusy = ref(false);

function openCreateDialog() {
  createErr.value = null;
  createDialogEl.value?.showModal();
}

function closeCreateDialog() {
  createDialogEl.value?.close();
}

function onCreateDialogCancel(ev: Event) {
  if (createBusy.value) {
    ev.preventDefault();
  }
}

async function submitCreateNs() {
  createErr.value = null;
  createBusy.value = true;
  try {
    const tags = createTags.value
      .split(",")
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
    const quotaNum = Number.parseInt(createQuotaStr.value.trim(), 10);
    const quota =
      createQuotaStr.value.trim() !== "" && !Number.isNaN(quotaNum) && quotaNum > 0
        ? quotaNum
        : undefined;
    await ctx.createAndSelect({
      ns_slug: createSlug.value.trim(),
      display_name: createDisplayName.value.trim(),
      description: createDesc.value.trim() || undefined,
      tags: tags.length > 0 ? tags : undefined,
      quota_prompts_max: quota,
    });
    createSlug.value = "";
    createDisplayName.value = "";
    createDesc.value = "";
    createTags.value = "";
    createQuotaStr.value = "";
    closeCreateDialog();
  } catch (e) {
    const msg =
      e instanceof ApiError ? `${e.code}: ${e.message}` : String((e as Error)?.message ?? e);
    createErr.value = msg;
  } finally {
    createBusy.value = false;
  }
}

onMounted(() => {
  void ctx.reloadNamespaces().catch(() => {
    /* surfaced in namespacesError */
  });
});

function requestArchive() {
  if (!ctx.selectedNsSlug) return;
  archiveDialogOpen.value = true;
}

function requestRestore() {
  if (!ctx.selectedNsSlug) return;
  restoreDialogOpen.value = true;
}

async function onArchiveConfirmed() {
  const ns = ctx.selectedNsSlug;
  if (!ns) return;
  confirmBusy.value = true;
  try {
    await ctx.archiveSelected(ns);
    archiveDialogOpen.value = false;
  } catch (_e: unknown) {
    void ctx.reloadNamespaces().catch(() => undefined);
  } finally {
    confirmBusy.value = false;
  }
}

async function onRestoreConfirmed() {
  const ns = ctx.selectedNsSlug;
  if (!ns) return;
  confirmBusy.value = true;
  try {
    await ctx.restoreSelected(ns);
    restoreDialogOpen.value = false;
  } catch (_e: unknown) {
    void ctx.reloadNamespaces().catch(() => undefined);
  } finally {
    confirmBusy.value = false;
  }
}
</script>

<template>
  <div class="page">
    <section class="hero" aria-labelledby="dash-welcome">
      <h2 id="dash-welcome" class="hero-title">{{ t("workspace.welcomeTitle") }}</h2>
      <p class="hero-lede">{{ t("workspace.welcomeLead") }}</p>

      <p class="hero-create-cta">
        <button type="button" class="btn-primary btn-open-create-ns" @click="openCreateDialog">
          {{ t("workspace.nsCreateTitle") }}
        </button>
      </p>

      <div v-if="selected" class="ns-summary" aria-live="polite">
        <div class="ns-summary-main">
          <span class="ns-summary-k">{{ t("workspace.nsSelectedTitle") }}</span>
          <span class="ns-summary-row">
            <span class="quota-text">{{ quotaLine(selected) }}</span>
            <span v-if="selected.archived_at" class="ns-arch-hint">{{ t("workspace.nsArchivedBadge") }}</span>
          </span>
        </div>
        <div class="ns-actions">
          <button
            v-if="selected.status === 'active'"
            type="button"
            class="btn-archive"
            :disabled="ctx.namespacesLoading"
            @click="requestArchive"
          >
            {{ t("workspace.nsArchiveButton") }}
          </button>
          <button
            v-else
            type="button"
            class="btn-restore"
            :disabled="ctx.namespacesLoading"
            @click="requestRestore"
          >
            {{ t("workspace.nsRestoreButton") }}
          </button>
        </div>
      </div>

      <dialog
        ref="createDialogEl"
        class="ns-create-dialog"
        aria-labelledby="ns-create-heading"
        @cancel="onCreateDialogCancel"
      >
        <div class="create-dialog-surface">
          <header class="create-dialog-head">
            <h3 id="ns-create-heading" class="panel-title">{{ t("workspace.nsCreateTitle") }}</h3>
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
          <form class="create-form" @submit.prevent="submitCreateNs">
            <label class="fld">
              <span>{{ t("workspace.nsCreateSlug") }}</span>
              <input v-model.trim="createSlug" class="ctx-input mono" type="text" autocomplete="off" />
            </label>
            <label class="fld">
              <span>{{ t("workspace.nsCreateDisplayName") }}</span>
              <input v-model.trim="createDisplayName" type="text" autocomplete="organization" />
            </label>
            <label class="fld span-2">
              <span>{{ t("workspace.nsCreateDescription") }}</span>
              <input v-model.trim="createDesc" type="text" />
            </label>
            <label class="fld span-2">
              <span>{{ t("workspace.nsCreateTags") }}</span>
              <input v-model.trim="createTags" type="text" />
            </label>
            <label class="fld">
              <span>{{ t("workspace.nsCreateQuota") }}</span>
              <input v-model.trim="createQuotaStr" inputmode="numeric" class="mono" />
              <small class="help">{{ t("workspace.nsCreateQuotaHelp") }}</small>
            </label>
            <div class="create-dialog-footer">
              <button type="button" class="btn-dialog-muted" :disabled="createBusy" @click="closeCreateDialog">
                {{ t("common.close") }}
              </button>
              <button type="submit" class="btn-primary" :disabled="createBusy || !createSlug || !createDisplayName">
                {{ createBusy ? t("workspace.nsBusy") : t("workspace.nsCreateButton") }}
              </button>
            </div>
            <p v-if="createErr" class="form-err">{{ createErr }}</p>
          </form>
        </div>
      </dialog>
    </section>

    <ConfirmDialog
      v-model="archiveDialogOpen"
      :title="t('workspace.nsArchiveDialogTitle')"
      :message="t('workspace.nsArchiveDialogMessage')"
      confirm-variant="danger"
      :busy="confirmBusy"
      @confirm="onArchiveConfirmed"
    />
    <ConfirmDialog
      v-model="restoreDialogOpen"
      :title="t('workspace.nsRestoreDialogTitle')"
      :message="t('workspace.nsRestoreDialogMessage')"
      :busy="confirmBusy"
      @confirm="onRestoreConfirmed"
    />
  </div>
</template>

<style scoped>
.page {
  max-width: 1120px;
}

.hero {
  margin-bottom: 1.75rem;
}

.hero-title {
  margin: 0 0 0.45rem;
  font-size: 1.45rem;
  font-weight: 750;
  letter-spacing: -0.03em;
}

.hero-lede {
  margin: 0 0 1rem;
  max-width: 52ch;
  color: var(--fg-muted);
  line-height: 1.5;
  font-size: 0.9125rem;
}

.hero-create-cta {
  margin: 0 0 1rem;
}

.btn-open-create-ns {
  cursor: pointer;
}

.ns-create-dialog {
  padding: 0;
  border: none;
  border-radius: 14px;
  background: var(--elev);
  color: var(--fg);
  max-width: min(540px, calc(100vw - 2rem));
  width: 100%;
  box-shadow:
    0 0 0 1px var(--border-subtle),
    0 24px 64px rgba(0, 0, 0, 0.42);
}

.ns-create-dialog::backdrop {
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

.ns-create-dialog .create-form {
  padding: 1rem 1.1rem 1.15rem;
}

.create-dialog-footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
  gap: 0.55rem;
  grid-column: 1 / -1;
  margin-top: 0.15rem;
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
  border-color: var(--border-strong);
  color: var(--fg);
}

.btn-dialog-muted:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.ns-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.65rem 1rem;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  background: var(--elev);
}

.ns-summary-main {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.ns-summary-k {
  margin: 0;
  font-size: 0.6875rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--accent);
}

.ns-summary-row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.45rem;
  font-size: 0.875rem;
  color: var(--fg-muted);
}

.quota-text {
  margin: 0;
}

.ns-arch-hint {
  font-size: 0.7825rem;
  color: var(--fg-soft);
}

.ns-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  flex-shrink: 0;
}

.btn-archive {
  cursor: pointer;
  border-radius: 9px;
  padding: 0.4rem 0.82rem;
  font-weight: 650;
  font-size: 0.8rem;
  border: 1px solid color-mix(in srgb, #f97316 35%, transparent);
  background: color-mix(in srgb, #f97316 14%, var(--elev-2));
}

.btn-restore {
  cursor: pointer;
  border-radius: 9px;
  padding: 0.4rem 0.82rem;
  font-weight: 650;
  font-size: 0.8rem;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
}

.panel-title {
  font-size: 1rem;
  font-weight: 700;
}

.create-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

.fld.span-2 {
  grid-column: span 2;
}

@media (max-width: 640px) {
  .fld.span-2 {
    grid-column: span 1;
  }
}

.fld input {
  padding: 0.45rem 0.62rem;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  font-size: 0.85rem;
}

.mono {
  font-family: "JetBrains Mono", ui-monospace, monospace;
}

.help {
  font-weight: 500;
  text-transform: none;
  letter-spacing: 0;
  color: var(--fg-soft);
}

.ctx-input {
  font-size: 0.84rem;
}

.form-err {
  grid-column: 1 / -1;
  margin: 0;
  padding: 0.45rem;
  border-radius: 8px;
  border: 1px solid color-mix(in srgb, #f97316 30%, transparent);
  background: color-mix(in srgb, #f97316 12%, transparent);
  color: var(--fg-muted);
  font-size: 0.8125rem;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 1rem;
  border-radius: 10px;
  font-weight: 650;
  font-size: 0.8575rem;
  background: linear-gradient(145deg, #818cf8 0%, #6366f1 52%, #4f46e5 100%);
  color: #fafafa;
  border: 1px solid color-mix(in srgb, #fff 42%, transparent);
  cursor: pointer;
}

.btn-primary:hover {
  filter: brightness(1.06);
}

.btn-primary:disabled {
  opacity: 0.58;
  cursor: not-allowed;
}
</style>
