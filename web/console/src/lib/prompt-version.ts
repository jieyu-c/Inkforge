/** Semver label helpers (align with services/console/internal/pkg/promptversion/label.go). */

const semverCore = /^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z][0-9A-Za-z.-]*))?$/;

export function normalizeVersionLabel(raw: string): string {
  let s = raw.trim();
  if (s.startsWith("v") || s.startsWith("V")) s = s.slice(1);
  return s;
}

export function isValidVersionLabel(raw: string): boolean {
  return semverCore.test(normalizeVersionLabel(raw));
}

function parseTriple(label: string): [number, number, number] | null {
  const m = normalizeVersionLabel(label).match(semverCore);
  if (!m) return null;
  return [Number(m[1]), Number(m[2]), Number(m[3])];
}

function tripleLess(a: [number, number, number], b: [number, number, number]): boolean {
  if (a[0] !== b[0]) return a[0] < b[0];
  if (a[1] !== b[1]) return a[1] < b[1];
  return a[2] < b[2];
}

/** Next default label: 1.0.0 or patch+1 of the greatest existing semver. */
export function suggestNextVersionLabel(existing: string[]): string {
  let best: [number, number, number] | null = null;
  for (const raw of existing) {
    const t = parseTriple(String(raw));
    if (!t) continue;
    if (!best || tripleLess(best, t)) best = t;
  }
  if (!best) return "1.0.0";
  return `${best[0]}.${best[1]}.${best[2] + 1}`;
}

export function formatVersionDisplay(version: string): string {
  const n = normalizeVersionLabel(version);
  return n ? `v${n}` : version;
}
