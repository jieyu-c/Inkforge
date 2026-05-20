<script setup lang="ts">
import type { NamespaceDetail } from "@/api/namespaces";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const ctx = useWorkspaceContextStore();

const dialogEl = ref<HTMLDialogElement | null>(null);
const restoreDialogOpen = ref(false);
const restoreTarget = ref<NamespaceDetail | null>(null);
const busy = ref(false);

const archived = computed(() => ctx.archivedNamespaces);
const showEntry = computed(() => archived.value.length > 0);

function quotaLine(ns: NamespaceDetail): string {
  if (ns.quota_prompts_max > 0) {
    return t("workspace.nsQuotaUsed", { count: ns.prompt_count, cap: ns.quota_prompts_max });
  }
  return t("workspace.nsQuotaUnlimited", { count: ns.prompt_count });
}

function openDialog() {
  dialogEl.value?.showModal();
}

function closeDialog() {
  dialogEl.value?.close();
}

function onDialogCancel(ev: Event) {
  if (busy.value) ev.preventDefault();
}

function switchTo(ns: NamespaceDetail) {
  ctx.setSelectedNsSlug(ns.ns_slug);
  closeDialog();
}

function requestRestore(ns: NamespaceDetail) {
  restoreTarget.value = ns;
  restoreDialogOpen.value = true;
}

async function onRestoreConfirmed() {
  const ns = restoreTarget.value;
  if (!ns) return;
  busy.value = true;
  try {
    await ctx.restoreSelected(ns.ns_slug);
    restoreDialogOpen.value = false;
    restoreTarget.value = null;
    if (ctx.archivedNamespaces.length === 0) {
      closeDialog();
    }
  } catch (_e: unknown) {
    void ctx.reloadNamespaces().catch(() => undefined);
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <button
    v-if="showEntry"
    type="button"
    class="archived-entry"
    :disabled="ctx.namespacesLoading"
    @click="openDialog"
  >
    {{ t("workspace.archivedToolLabel", { count: archived.length }) }}
  </button>

  <dialog
    ref="dialogEl"
    class="archived-dialog"
    aria-labelledby="archived-ns-heading"
    @cancel="onDialogCancel"
  >
    <div class="archived-surface">
      <header class="archived-head">
        <div>
          <h3 id="archived-ns-heading" class="archived-title">{{ t("workspace.archivedDialogTitle") }}</h3>
          <p class="archived-lede">{{ t("workspace.archivedDialogLead") }}</p>
        </div>
        <button
          type="button"
          class="archived-close"
          :disabled="busy"
          :aria-label="t('common.close')"
          @click="closeDialog"
        >
          ×
        </button>
      </header>

      <p v-if="archived.length === 0" class="archived-empty muted">{{ t("workspace.archivedDialogEmpty") }}</p>

      <ul v-else class="archived-list" role="list">
        <li v-for="ns in archived" :key="ns.ns_slug" class="archived-row">
          <div class="archived-row-main">
            <span class="archived-name">{{ ns.display_name }}</span>
            <span class="archived-meta mono muted">{{ ns.ns_slug }} · {{ quotaLine(ns) }}</span>
          </div>
          <div class="archived-row-actions">
            <button
              type="button"
              class="btn-row"
              :disabled="busy || ctx.selectedNsSlug === ns.ns_slug"
              @click="switchTo(ns)"
            >
              {{
                ctx.selectedNsSlug === ns.ns_slug
                  ? t("workspace.archivedCurrent")
                  : t("workspace.archivedSwitch")
              }}
            </button>
            <button type="button" class="btn-row primary" :disabled="busy" @click="requestRestore(ns)">
              {{ t("workspace.nsRestoreButton") }}
            </button>
          </div>
        </li>
      </ul>
    </div>
  </dialog>

  <ConfirmDialog
    v-model="restoreDialogOpen"
    :title="t('workspace.nsRestoreDialogTitle')"
    :message="t('workspace.nsRestoreDialogMessage')"
    :busy="busy"
    @confirm="onRestoreConfirmed"
  />
</template>

<style scoped>
.archived-entry {
  cursor: pointer;
  border-radius: 8px;
  padding: 0.32rem 0.62rem;
  font-weight: 600;
  font-size: 0.75rem;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg-soft);
  white-space: nowrap;
}

.archived-entry:hover:not(:disabled) {
  color: var(--fg-muted);
  border-color: var(--accent-soft);
}

.archived-entry:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.archived-dialog {
  padding: 0;
  border: none;
  border-radius: 14px;
  background: var(--elev);
  color: var(--fg);
  max-width: min(520px, calc(100vw - 2rem));
  width: 100%;
  box-shadow:
    0 0 0 1px var(--border-subtle),
    0 24px 64px rgba(0, 0, 0, 0.42);
}

.archived-dialog::backdrop {
  background: rgba(0, 0, 0, 0.5);
}

.archived-surface {
  overflow: hidden;
}

.archived-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.85rem;
  border-bottom: 1px solid var(--border-subtle);
}

.archived-title {
  margin: 0 0 0.25rem;
  font-size: 1rem;
  font-weight: 700;
}

.archived-lede {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--fg-muted);
  line-height: 1.45;
  max-width: 36ch;
}

.archived-close {
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

.archived-close:hover:not(:disabled) {
  background: var(--elev-2);
  color: var(--fg);
}

.archived-empty {
  margin: 0;
  padding: 1.25rem 1.1rem;
  font-size: 0.875rem;
}

.archived-list {
  list-style: none;
  margin: 0;
  padding: 0.65rem 0.55rem 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  max-height: min(20rem, 55vh);
  overflow-y: auto;
}

.archived-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem 0.75rem;
  padding: 0.65rem 0.7rem;
  border-radius: 10px;
  border: 1px solid var(--border-faint);
  background: color-mix(in srgb, var(--elev-2) 90%, transparent);
}

.archived-row-main {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.archived-name {
  font-weight: 650;
  font-size: 0.9rem;
}

.archived-meta {
  font-size: 0.72rem;
}

.archived-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  flex-shrink: 0;
}

.btn-row {
  cursor: pointer;
  border-radius: 8px;
  padding: 0.32rem 0.62rem;
  font-weight: 600;
  font-size: 0.75rem;
  border: 1px solid var(--border-strong);
  background: var(--elev);
  color: var(--fg-muted);
}

.btn-row:hover:not(:disabled) {
  color: var(--fg);
}

.btn-row.primary {
  border-color: color-mix(in srgb, var(--accent) 35%, var(--border-strong));
  color: var(--fg);
}

.btn-row:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.mono {
  font-family: "JetBrains Mono", ui-monospace, monospace;
}

.muted {
  color: var(--fg-muted);
}
</style>
