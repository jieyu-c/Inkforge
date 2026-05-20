<script setup lang="ts">
import { ApiError } from "@/api/client";
import type { NamespaceDetail } from "@/api/namespaces";
import ArchivedNamespacesTool from "@/components/ArchivedNamespacesTool.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import { computed, ref } from "vue";
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
      display_name: createDisplayName.value.trim(),
      description: createDesc.value.trim() || undefined,
      tags: tags.length > 0 ? tags : undefined,
      quota_prompts_max: quota,
    });
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
  <div class="ns-actions" role="group" :aria-label="t('workspace.nsActionsAria')">
    <ArchivedNamespacesTool />
    <template v-if="selected">
      <span class="ns-quota" :title="quotaLine(selected)">{{ quotaLine(selected) }}</span>
      <span v-if="selected.archived_at" class="ns-arch-badge">{{ t("workspace.nsArchivedBadge") }}</span>
      <button
        v-if="selected.status === 'active'"
        type="button"
        class="btn-ns-muted"
        :disabled="ctx.namespacesLoading"
        @click="requestArchive"
      >
        {{ t("workspace.nsArchiveButton") }}
      </button>
      <button
        v-else
        type="button"
        class="btn-ns-muted"
        :disabled="ctx.namespacesLoading"
        @click="requestRestore"
      >
        {{ t("workspace.nsRestoreButton") }}
      </button>
    </template>

    <button type="button" class="btn-ns-create" @click="openCreateDialog">
      {{ t("workspace.nsCreateTitle") }}
    </button>

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
          <label class="fld span-2">
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
            <button type="submit" class="btn-primary" :disabled="createBusy || !createDisplayName">
              {{ createBusy ? t("workspace.nsBusy") : t("workspace.nsCreateButton") }}
            </button>
          </div>
          <p v-if="createErr" class="form-err">{{ createErr }}</p>
        </form>
      </div>
    </dialog>

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
.ns-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem 0.55rem;
}

.ns-quota {
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: color-mix(in srgb, var(--fg) 88%, var(--fg-muted));
  white-space: nowrap;
}

.ns-arch-badge {
  font-size: 0.68rem;
  font-weight: 650;
  color: var(--fg-soft);
  white-space: nowrap;
}

.btn-ns-muted,
.btn-ns-create {
  cursor: pointer;
  border-radius: 8px;
  padding: 0.32rem 0.62rem;
  font-weight: 600;
  font-size: 0.75rem;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg-muted);
  white-space: nowrap;
}

.btn-ns-muted:hover:not(:disabled),
.btn-ns-create:hover:not(:disabled) {
  color: var(--fg);
  border-color: var(--accent-soft);
}

.btn-ns-create {
  color: var(--fg);
  border-color: color-mix(in srgb, var(--accent) 35%, var(--border-strong));
  background: color-mix(in srgb, var(--accent) 10%, var(--elev-2));
}

.btn-ns-muted:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.ns-create-dialog .create-form {
  padding: 1rem 1.1rem 1.15rem;
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
  padding: 0.5rem 1rem;
  border-radius: 10px;
  font-weight: 650;
  font-size: 0.8575rem;
  background: linear-gradient(145deg, #818cf8 0%, #6366f1 52%, #4f46e5 100%);
  color: #fafafa;
  border: 1px solid color-mix(in srgb, #fff 42%, transparent);
  cursor: pointer;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.06);
}

.btn-primary:disabled {
  opacity: 0.58;
  cursor: not-allowed;
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
</style>
