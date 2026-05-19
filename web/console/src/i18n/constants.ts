export const LOCALE_STORAGE_KEY = "inkforge.locale";

export const SUPPORTED_LOCALES = ["en", "zh"] as const;

export type AppLocale = (typeof SUPPORTED_LOCALES)[number];

export function isAppLocale(value: string): value is AppLocale {
  return (SUPPORTED_LOCALES as readonly string[]).includes(value);
}
