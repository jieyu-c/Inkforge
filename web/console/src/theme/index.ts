import type { AppTheme } from "./constants";
import {
  DEFAULT_THEME,
  THEME_STORAGE_KEY,
  isAppTheme,
} from "./constants";

export function initialTheme(): AppTheme {
  if (typeof localStorage === "undefined") {
    return DEFAULT_THEME;
  }
  const raw = localStorage.getItem(THEME_STORAGE_KEY);
  if (!raw || !isAppTheme(raw)) {
    return DEFAULT_THEME;
  }
  return raw;
}

export function persistTheme(theme: AppTheme): void {
  localStorage.setItem(THEME_STORAGE_KEY, theme);
}

/** Keep `<html data-theme>` and `color-scheme` aligned with the active palette. */
export function syncDocumentTheme(theme: AppTheme): void {
  document.documentElement.setAttribute("data-theme", theme);
  document.documentElement.style.colorScheme = theme;
}

export function setAppTheme(theme: AppTheme): void {
  persistTheme(theme);
  syncDocumentTheme(theme);
}

export { type AppTheme, DEFAULT_THEME };
