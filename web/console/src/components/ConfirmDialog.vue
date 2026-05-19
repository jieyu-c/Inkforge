<script setup lang="ts">
import { computed, nextTick, ref, watch, useId } from "vue";
import { useI18n } from "vue-i18n";

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    /** Dialog heading — pass translated string from caller */
    title: string;
    /**
     * Optional plain-text detail under the title when the default slot is empty.
     */
    message?: string;
    confirmLabel?: string;
    cancelLabel?: string;
    busy?: boolean;
    /**
     * `danger`: destructive / warm confirmations.
     * `primary`: default accent confirmations.
     */
    confirmVariant?: "primary" | "danger";
    /** Optional explicit id for `aria-labelledby` */
    titleId?: string;
  }>(),
  {
    message: "",
    confirmVariant: "primary",
  },
);

defineSlots<{
  default?: () => unknown;
  footer?: () => unknown;
}>();

const emit = defineEmits<{
  "update:modelValue": [open: boolean];
  confirm: [];
  cancel: [];
}>();

const { t } = useI18n();
const autoTitleId = useId();
const headingId = computed(() => props.titleId ?? autoTitleId);

const dlg = ref<HTMLDialogElement | null>(null);

const cancelText = computed(() => props.cancelLabel ?? t("common.confirmCancel"));
const confirmText = computed(() => props.confirmLabel ?? t("common.confirmOk"));

watch(
  () => props.modelValue,
  async (open) => {
    await nextTick();
    const el = dlg.value;
    if (!el) return;
    if (open && !el.open) el.showModal();
    if (!open && el.open) el.close();
  },
);

function onDialogCancel(ev: Event) {
  if (props.busy) {
    ev.preventDefault();
    return;
  }
  emit("cancel");
  emit("update:modelValue", false);
}

function onCancelClick() {
  if (props.busy) return;
  emit("cancel");
  emit("update:modelValue", false);
}

function onConfirmClick() {
  if (props.busy) return;
  emit("confirm");
}

const bodyVisible = computed(
  () => Boolean(props.message && props.message.length > 0),
);
</script>

<template>
  <dialog
    ref="dlg"
    class="confirm-dlg"
    role="alertdialog"
    :aria-labelledby="headingId"
    @cancel="onDialogCancel"
  >
    <div class="confirm-surface">
      <header class="confirm-head">
        <h3 :id="headingId" class="confirm-title">{{ title }}</h3>
        <button
          type="button"
          class="confirm-close"
          :disabled="busy"
          :aria-label="t('common.close')"
          @click="onCancelClick"
        >
          ×
        </button>
      </header>
      <div v-if="$slots.default || bodyVisible" class="confirm-body">
        <slot>
          <p v-if="bodyVisible" class="confirm-msg">{{ message }}</p>
        </slot>
      </div>
      <footer class="confirm-foot">
        <slot name="footer" />
        <button type="button" class="confirm-btn-cancel" :disabled="busy" @click="onCancelClick">
          {{ cancelText }}
        </button>
        <button
          type="button"
          :class="confirmVariant === 'danger' ? 'confirm-btn-danger' : 'confirm-btn-primary'"
          :disabled="busy"
          @click="onConfirmClick"
        >
          {{ confirmText }}
        </button>
      </footer>
    </div>
  </dialog>
</template>

<style scoped>
.confirm-dlg {
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

.confirm-dlg::backdrop {
  background: rgba(0, 0, 0, 0.5);
}

.confirm-surface {
  overflow: hidden;
}

.confirm-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}

.confirm-title {
  margin: 0;
  padding-right: 0.35rem;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
}

.confirm-close {
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

.confirm-close:hover:not(:disabled) {
  background: var(--elev-2);
  color: var(--fg);
}

.confirm-close:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.confirm-body {
  padding: 0.85rem 1.1rem 0;
}

.confirm-msg {
  margin: 0;
  padding-bottom: 1rem;
  font-size: 0.8675rem;
  line-height: 1.5;
  color: var(--fg-muted);
}

.confirm-foot {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: center;
  gap: 0.55rem;
  padding: 1rem 1.1rem 1.05rem;
  border-top: 1px solid var(--border-subtle);
}

.confirm-btn-cancel {
  cursor: pointer;
  border-radius: 10px;
  padding: 0.5rem 1rem;
  font-weight: 600;
  font-size: 0.8575rem;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg-muted);
  font-family: inherit;
}

.confirm-btn-cancel:hover:not(:disabled) {
  color: var(--fg);
}

.confirm-btn-cancel:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.confirm-btn-primary {
  cursor: pointer;
  border-radius: 10px;
  padding: 0.5rem 1rem;
  font-weight: 650;
  font-size: 0.8575rem;
  background: linear-gradient(145deg, #818cf8 0%, #6366f1 52%, #4f46e5 100%);
  color: #fafafa;
  border: 1px solid color-mix(in srgb, #fff 42%, transparent);
  font-family: inherit;
}

.confirm-btn-primary:hover:not(:disabled) {
  filter: brightness(1.06);
}

.confirm-btn-danger {
  cursor: pointer;
  border-radius: 10px;
  padding: 0.5rem 1rem;
  font-weight: 650;
  font-size: 0.8575rem;
  border: 1px solid color-mix(in srgb, #f97316 42%, transparent);
  background: color-mix(in srgb, #f97316 18%, var(--elev-2));
  color: color-mix(in srgb, #fed7aa 55%, var(--fg));
  font-family: inherit;
}

.confirm-btn-danger:hover:not(:disabled) {
  filter: brightness(1.05);
}

.confirm-btn-primary:disabled,
.confirm-btn-danger:disabled {
  opacity: 0.58;
  cursor: not-allowed;
}
</style>
