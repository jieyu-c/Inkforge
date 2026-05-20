<script setup lang="ts">
import type { AppLocale } from "@/i18n/constants";
import { setAppLocale } from "@/i18n";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { locale, t } = useI18n();

const rootClass = computed(() => ({
  "locale-switcher": true,
}));

function pick(next: AppLocale) {
  if (locale.value !== next) {
    setAppLocale(next);
  }
}
</script>

<template>
  <div
    :class="rootClass"
    role="radiogroup"
    :aria-label="t('localeSwitcher.aria')"
  >
    <button
      type="button"
      class="locale-switcher__seg"
      :class="{ 'is-active': locale === 'en' }"
      role="radio"
      :aria-checked="locale === 'en'"
      @click="pick('en')"
    >
      {{ t("localeSwitcher.enShort") }}
    </button>
    <button
      type="button"
      class="locale-switcher__seg"
      :class="{ 'is-active': locale === 'zh' }"
      role="radio"
      :aria-checked="locale === 'zh'"
      @click="pick('zh')"
    >
      {{ t("localeSwitcher.zhShort") }}
    </button>
  </div>
</template>

<style scoped>
.locale-switcher {
  display: inline-flex;
  align-items: stretch;
  padding: 3px;
  gap: 0;
  border-radius: 999px;
  border: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 88%, transparent);
  box-shadow:
    0 0 0 1px color-mix(in srgb, #000 18%, transparent) inset,
    0 1px 14px rgba(0, 0, 0, 0.18);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.locale-switcher__seg {
  border: none;
  border-radius: 999px;
  padding: 0.42rem 0.88rem;
  min-width: 2.7rem;
  font: inherit;
  font-size: 0.77rem;
  font-weight: 650;
  letter-spacing: 0.04em;
  color: var(--fg-soft);
  background: transparent;
  cursor: pointer;
  transition:
    background 0.16s ease,
    color 0.16s ease,
    box-shadow 0.16s ease,
    transform 0.08s ease;
}

.locale-switcher__seg:hover:not(.is-active) {
  color: var(--fg-muted);
}

.locale-switcher__seg:active {
  transform: scale(0.98);
}

.locale-switcher__seg.is-active {
  color: #fafafa;
  background: linear-gradient(
    148deg,
    color-mix(in srgb, var(--accent) 92%, white) 0%,
    var(--accent) 100%
  );
  box-shadow:
    0 0 0 1px color-mix(in srgb, #fff 35%, transparent) inset,
    0 10px 24px color-mix(in srgb, var(--accent) 42%, transparent);
}

@media (prefers-reduced-motion: reduce) {
  .locale-switcher__seg {
    transition: none;
    transform: none;
  }
}
</style>
