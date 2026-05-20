import { createI18n } from "vue-i18n";
import type { AppLocale } from "./constants";
import { LOCALE_STORAGE_KEY, isAppLocale } from "./constants";
import en from "@/locales/en.json";
import zh from "@/locales/zh.json";
import { syncDocumentLang } from "./sync-document-lang";

function detectBrowserLocale(): AppLocale {
  if (typeof navigator === "undefined") {
    return "en";
  }
  const langs =
    navigator.languages?.length > 0 ? navigator.languages : [navigator.language];
  for (const lang of langs) {
    const normalized = lang.toLowerCase();
    if (normalized === "zh" || normalized.startsWith("zh-")) {
      return "zh";
    }
  }
  return "en";
}

function initialLocale(): AppLocale {
  if (typeof localStorage !== "undefined") {
    const raw = localStorage.getItem(LOCALE_STORAGE_KEY);
    if (raw && isAppLocale(raw)) {
      return raw;
    }
  }
  return detectBrowserLocale();
}

export const i18n = createI18n({
  legacy: false,
  locale: initialLocale(),
  fallbackLocale: "en",
  messages: {
    en,
    zh,
  },
  missingWarn: false,
  fallbackWarn: false,
});

export function persistLocale(locale: AppLocale): void {
  localStorage.setItem(LOCALE_STORAGE_KEY, locale);
}

export function setAppLocale(locale: AppLocale): void {
  const i18nGlobal = i18n.global as { locale: { value: AppLocale } };
  i18nGlobal.locale.value = locale;
  persistLocale(locale);
  syncDocumentLang(locale);
}

export { syncDocumentLang, type AppLocale };
