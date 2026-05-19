import type { AppLocale } from "./constants";

/** Keep `<html lang>` aligned with ICU locale hints (a11y + font rendering). */
export function syncDocumentLang(locale: AppLocale): void {
  document.documentElement.lang = locale === "zh" ? "zh-CN" : "en";
}
