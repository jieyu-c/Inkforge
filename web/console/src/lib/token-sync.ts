let syncHandler: (() => void) | null = null;

/** Registered from main.ts after pinia is installed; avoids circular imports. */
export function registerAccessTokenSync(handler: () => void): void {
  syncHandler = handler;
}

export function notifyAccessTokenChanged(): void {
  syncHandler?.();
}
