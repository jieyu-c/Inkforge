const STORAGE_KEY = "inkforge_access";

export function readStoredAccessToken(): string | null {
  try {
    return sessionStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

export function writeAccessToken(token: string | null): void {
  try {
    if (token) sessionStorage.setItem(STORAGE_KEY, token);
    else sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    /* ignore quota / private mode */
  }
}
