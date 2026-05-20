<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";

export type NamespaceSelectOption = {
  value: string;
  label: string;
};

const props = withDefaults(
  defineProps<{
    id?: string;
    modelValue: string;
    options: NamespaceSelectOption[];
    emptyLabel: string;
    /** Override trigger label (e.g. archived namespace not in dropdown options). */
    selectedLabel?: string;
    disabled?: boolean;
    ariaBusy?: boolean;
  }>(),
  {
    disabled: false,
    ariaBusy: false,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const open = ref(false);
const rootEl = ref<HTMLElement | null>(null);

const selectedLabel = computed(() => {
  if (props.selectedLabel) return props.selectedLabel;
  if (!props.modelValue) return props.emptyLabel;
  return (
    props.options.find((option) => option.value === props.modelValue)?.label ??
    props.modelValue
  );
});

function toggle() {
  if (props.disabled) return;
  open.value = !open.value;
}

function pick(value: string) {
  emit("update:modelValue", value);
  open.value = false;
}

function onDocPointerDown(event: PointerEvent) {
  if (!open.value) return;
  const root = rootEl.value;
  if (root && !root.contains(event.target as Node)) {
    open.value = false;
  }
}

function onTriggerKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    open.value = false;
    return;
  }
  if (event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    toggle();
  }
}

onMounted(() => {
  document.addEventListener("pointerdown", onDocPointerDown);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", onDocPointerDown);
});
</script>

<template>
  <div
    ref="rootEl"
    class="ns-select-wrap"
    :class="{ 'is-open': open, 'is-disabled': disabled }"
  >
    <button
      :id="id"
      type="button"
      class="ns-select-trigger mono"
      :disabled="disabled"
      :aria-busy="ariaBusy"
      aria-haspopup="listbox"
      :aria-expanded="open"
      @click="toggle"
      @keydown="onTriggerKeydown"
    >
      <span class="ns-select-value">{{ selectedLabel }}</span>
      <span class="ns-select-chevron" aria-hidden="true">▾</span>
    </button>

    <ul v-show="open" class="ns-select-menu mono" role="listbox" :aria-labelledby="id">
      <li role="presentation">
        <button
          type="button"
          class="ns-select-option"
          role="option"
          :aria-selected="modelValue === ''"
          @click="pick('')"
        >
          {{ emptyLabel }}
        </button>
      </li>
      <li v-for="option in options" :key="option.value" role="presentation">
        <button
          type="button"
          class="ns-select-option"
          role="option"
          :aria-selected="modelValue === option.value"
          @click="pick(option.value)"
        >
          {{ option.label }}
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ns-select-wrap {
  position: relative;
  min-width: min(420px, 78vw);
  max-width: 100%;
}

.ns-select-trigger,
.ns-select-option {
  box-sizing: border-box;
  width: 100%;
  margin: 0;
  padding: 0.32rem 0.52rem;
  border: none;
  border-radius: 0;
  background: transparent;
  color: var(--fg);
  font: inherit;
  font-size: 0.7825rem;
  line-height: 1.5;
  text-align: left;
}

.ns-select-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  cursor: pointer;
  transition:
    border-color 0.12s ease,
    background 0.12s ease;
}

.ns-select-trigger:hover:not(:disabled) {
  border-color: var(--accent-soft);
}

.ns-select-wrap.is-open .ns-select-trigger {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 6%, var(--elev-2));
}

.ns-select-trigger:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.ns-select-value {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ns-select-chevron {
  flex-shrink: 0;
  font-size: 0.72rem;
  color: var(--fg-soft);
  line-height: 1;
}

.ns-select-menu {
  position: absolute;
  top: calc(100% + 2px);
  left: 0;
  right: 0;
  z-index: 80;
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: min(16rem, 50vh);
  overflow-y: auto;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  box-shadow: var(--shadow);
}

.ns-select-menu li + li {
  border-top: 1px solid var(--border-faint);
}

.ns-select-option {
  cursor: pointer;
  transition: background 0.12s ease;
}

.ns-select-option:hover {
  background: var(--pill-hover);
}

.ns-select-option[aria-selected="true"] {
  background: color-mix(in srgb, var(--accent) 12%, var(--elev-2));
}
</style>
