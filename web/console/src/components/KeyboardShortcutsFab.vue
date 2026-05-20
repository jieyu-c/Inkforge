<script setup lang="ts">
import {
  CONSOLE_SHORTCUT_CATALOG,
  formatShortcutChord,
} from "@/lib/keyboard-shortcuts";
import { ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const dlgEl = ref<HTMLDialogElement | null>(null);

function openDialog() {
  dlgEl.value?.showModal();
}

function closeDialog() {
  dlgEl.value?.close();
}
</script>

<template>
  <button
    type="button"
    class="shortcuts-fab"
    :aria-label="t('shortcuts.fabAria')"
    :title="t('shortcuts.fabAria')"
    @click="openDialog"
  >
    <span class="shortcuts-fab-icon" aria-hidden="true">⌨</span>
  </button>

  <dialog
    ref="dlgEl"
    class="shortcuts-dialog"
    aria-labelledby="inkforge-shortcuts-title"
    @cancel="closeDialog"
  >
    <div class="shortcuts-dialog-surface">
      <header class="shortcuts-dialog-head">
        <h2 id="inkforge-shortcuts-title" class="shortcuts-dialog-title">
          {{ t("shortcuts.title") }}
        </h2>
        <button
          type="button"
          class="shortcuts-dialog-close"
          :aria-label="t('common.close')"
          @click="closeDialog"
        >
          ×
        </button>
      </header>

      <div class="shortcuts-dialog-body">
        <section
          v-for="group in CONSOLE_SHORTCUT_CATALOG"
          :key="group.titleKey"
          class="shortcuts-group"
        >
          <h3 class="shortcuts-group-title">{{ t(group.titleKey) }}</h3>
          <ul class="shortcuts-list">
            <li v-for="item in group.items" :key="item.labelKey" class="shortcuts-row">
              <span class="shortcuts-label">{{ t(item.labelKey) }}</span>
              <span class="shortcuts-keys">
                <template
                  v-for="(chord, chordIdx) in item.chords"
                  :key="`${item.labelKey}-${chordIdx}`"
                >
                  <span v-if="chordIdx > 0" class="shortcuts-or">{{ t("shortcuts.or") }}</span>
                  <kbd
                    v-for="(keyLabel, keyIdx) in formatShortcutChord(chord)"
                    :key="`${chordIdx}-${keyIdx}`"
                    class="shortcuts-kbd"
                  >
                    {{ keyLabel }}
                  </kbd>
                </template>
              </span>
            </li>
          </ul>
        </section>
        <p class="shortcuts-foot muted small">{{ t("shortcuts.scopeNote") }}</p>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.shortcuts-fab {
  position: fixed;
  right: 1.35rem;
  bottom: 1.35rem;
  z-index: 50;
  width: 3rem;
  height: 3rem;
  border: none;
  border-radius: 999px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fafafa;
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--accent) 88%, white) 0%,
    var(--accent) 55%,
    color-mix(in srgb, var(--accent) 85%, #312e81) 100%
  );
  box-shadow:
    0 0 0 1px color-mix(in srgb, #fff 28%, transparent) inset,
    0 10px 28px color-mix(in srgb, var(--accent) 38%, transparent),
    0 4px 12px rgba(0, 0, 0, 0.22);
  transition:
    transform 0.12s ease,
    box-shadow 0.16s ease;
}

.shortcuts-fab:hover {
  transform: translateY(-2px);
  box-shadow:
    0 0 0 1px color-mix(in srgb, #fff 35%, transparent) inset,
    0 14px 36px color-mix(in srgb, var(--accent) 45%, transparent),
    0 6px 16px rgba(0, 0, 0, 0.28);
}

.shortcuts-fab:active {
  transform: translateY(0) scale(0.96);
}

.shortcuts-fab-icon {
  font-size: 1.2rem;
  line-height: 1;
}

.shortcuts-dialog {
  padding: 0;
  border: none;
  border-radius: 14px;
  background: var(--elev);
  color: var(--fg);
  max-width: min(480px, calc(100vw - 2rem));
  width: 100%;
  box-shadow:
    0 0 0 1px var(--border-subtle),
    0 24px 64px rgba(0, 0, 0, 0.42);
}

.shortcuts-dialog::backdrop {
  background: rgba(0, 0, 0, 0.5);
}

.shortcuts-dialog-surface {
  overflow: hidden;
}

.shortcuts-dialog-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1rem 1.1rem 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}

.shortcuts-dialog-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
}

.shortcuts-dialog-close {
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

.shortcuts-dialog-close:hover {
  background: var(--elev-2);
  color: var(--fg);
}

.shortcuts-dialog-body {
  padding: 0.85rem 1.1rem 1.1rem;
  max-height: min(70vh, 520px);
  overflow: auto;
}

.shortcuts-group + .shortcuts-group {
  margin-top: 1rem;
}

.shortcuts-group-title {
  margin: 0 0 0.5rem;
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-soft);
}

.shortcuts-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.45rem;
}

.shortcuts-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem 0.75rem;
  padding: 0.45rem 0.55rem;
  border-radius: 10px;
  background: color-mix(in srgb, var(--elev-2) 70%, transparent);
  border: 1px solid var(--border-faint);
}

.shortcuts-label {
  font-size: 0.84rem;
  color: var(--fg);
  min-width: 0;
  flex: 1 1 10rem;
}

.shortcuts-keys {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.28rem;
  flex-shrink: 0;
}

.shortcuts-or {
  font-size: 0.72rem;
  color: var(--fg-soft);
  margin: 0 0.12rem;
}

.shortcuts-kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.55rem;
  padding: 0.18rem 0.42rem;
  border-radius: 6px;
  border: 1px solid var(--border-strong);
  background: var(--bg-canvas);
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--fg-muted);
  line-height: 1.2;
}

.shortcuts-foot {
  margin: 1rem 0 0;
  line-height: 1.45;
}

@media (max-width: 780px) {
  .shortcuts-fab {
    right: 1rem;
    bottom: 1rem;
    width: 2.75rem;
    height: 2.75rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .shortcuts-fab {
    transition: none;
    transform: none;
  }
  .shortcuts-fab:hover,
  .shortcuts-fab:active {
    transform: none;
  }
}
</style>
