export const THEME_STORAGE_KEY = "inkforge.theme";

export const SUPPORTED_THEMES = ["dark", "light"] as const;

export type AppTheme = (typeof SUPPORTED_THEMES)[number];

export const DEFAULT_THEME: AppTheme = "light";

export function isAppTheme(value: string): value is AppTheme {
  return (SUPPORTED_THEMES as readonly string[]).includes(value);
}
