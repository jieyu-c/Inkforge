/** Mirrors backend `phone.Canonical`: mainland mobile → `+86…`. */
const cnTail = /^1[3-9]\d{9}$/;

function digits(raw: string): string {
  let s = "";
  for (const r of raw.trim()) {
    if (r >= "0" && r <= "9") s += r;
  }
  return s;
}

export function canonicalPhone(raw: string): string {
  const d = digits(raw);
  if (d.length === 11) {
    if (!cnTail.test(d)) return "";
    return `+86${d}`;
  }
  if (d.length === 13) {
    if (!d.startsWith("86")) return "";
    const tail = d.slice(2);
    if (!cnTail.test(tail)) return "";
    return `+${d}`;
  }
  return "";
}

/** Default matches `etc/console-api.yaml` MinPasswordLength. */
export const MIN_PASSWORD_LENGTH = 10;
