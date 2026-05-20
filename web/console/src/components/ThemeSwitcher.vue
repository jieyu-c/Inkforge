<script setup lang="ts">
import type { AppTheme } from "@/theme/constants";
import { DEFAULT_THEME } from "@/theme/constants";
import { setAppTheme } from "@/theme";
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const theme = ref<AppTheme>(DEFAULT_THEME);

onMounted(() => {
  const current = document.documentElement.getAttribute("data-theme");
  theme.value = current === "light" ? "light" : "dark";
});

const rootClass = computed(() => ({
  "theme-switcher": true,
}));

function pick(next: AppTheme) {
  if (theme.value !== next) {
    theme.value = next;
    setAppTheme(next);
  }
}
</script>

<template>
  <div
    :class="rootClass"
    role="radiogroup"
    :aria-label="t('themeSwitcher.aria')"
  >
    <button
      type="button"
      class="theme-switcher__seg"
      :class="{ 'is-active': theme === 'light' }"
      role="radio"
      :aria-checked="theme === 'light'"
      @click="pick('light')"
    >
      {{ t("themeSwitcher.light") }}
    </button>
    <button
      type="button"
      class="theme-switcher__seg"
      :class="{ 'is-active': theme === 'dark' }"
      role="radio"
      :aria-checked="theme === 'dark'"
      @click="pick('dark')"
    >
      {{ t("themeSwitcher.dark") }}
    </button>
  </div>
</template>

<style scoped>
.theme-switcher {
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

.theme-switcher__seg {
  border: none;
  border-radius: 999px;
  padding: 0.42rem 0.88rem;
  min-width: 3.25rem;
  font: inherit;
  font-size: 0.77rem;
  font-weight: 650;
  letter-spacing: 0.02em;
  color: var(--fg-soft);
  background: transparent;
  cursor: pointer;
  transition:
    background 0.16s ease,
    color 0.16s ease,
    box-shadow 0.16s ease,
    transform 0.08s ease;
}

.theme-switcher__seg:hover:not(.is-active) {
  color: var(--fg-muted);
}

.theme-switcher__seg:active {
  transform: scale(0.98);
}

.theme-switcher__seg.is-active {
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
  .theme-switcher__seg {
    transition: none;
    transform: none;
  }
}
</style>
