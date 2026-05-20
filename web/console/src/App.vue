<script setup lang="ts">
import { RouterView } from "vue-router";
</script>

<template>
  <RouterView />
</template>

<style>
:root,
[data-theme="dark"] {
  color-scheme: dark;
  --bg: #070a12;
  --bg-elevated: #0d111c;
  --surface: color-mix(in srgb, #121a2a 78%, transparent);
  --surface-2: color-mix(in srgb, #0a1628 92%, transparent);
  --accent: #6366f1;
  --accent-dim: color-mix(in srgb, var(--accent) 55%, #0f172a);
  --accent-glow: color-mix(in srgb, var(--accent) 45%, transparent);
  --border: color-mix(in srgb, var(--accent) 16%, #1e293b 84%);
  --border-dim: color-mix(in srgb, #94a3b8 22%, transparent);
  --ink: #e8eef7;
  --muted: #8b9cb3;
  --danger: #f87171;
  --ok: #4ade80;
  --shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 12%, transparent),
    0 24px 48px rgba(0, 0, 0, 0.45);

  /* Landing / chrome tokens (aligned to GitHub-style dark UI) */
  --bg-canvas: var(--bg);
  --fg: var(--ink);
  --fg-muted: var(--muted);
  --fg-soft: color-mix(in srgb, var(--muted) 88%, transparent);
  --border-subtle: color-mix(in srgb, #30363d 55%, transparent);
  --border-faint: color-mix(in srgb, #30363d 35%, transparent);
  --border-strong: color-mix(in srgb, #8b9cb3 32%, var(--border-dim));
  --accent-soft: color-mix(in srgb, var(--accent) 40%, transparent);
  --elev: var(--surface);
  --elev-2: var(--surface-2);
  --pill-hover: color-mix(in srgb, var(--muted) 12%, transparent);
}

[data-theme="light"] {
  color-scheme: light;
  --bg: #f4f6fb;
  --bg-elevated: #ffffff;
  --surface: color-mix(in srgb, #ffffff 94%, #e2e8f0);
  --surface-2: #eef2f9;
  --accent: #6366f1;
  --accent-dim: color-mix(in srgb, var(--accent) 32%, #f1f5f9);
  --accent-glow: color-mix(in srgb, var(--accent) 28%, transparent);
  --border: color-mix(in srgb, var(--accent) 10%, #cbd5e1 90%);
  --border-dim: color-mix(in srgb, #64748b 16%, transparent);
  --ink: #0f172a;
  --muted: #64748b;
  --danger: #dc2626;
  --ok: #16a34a;
  --shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 8%, transparent),
    0 18px 40px rgba(15, 23, 42, 0.08);

  --bg-canvas: var(--bg);
  --fg: var(--ink);
  --fg-muted: var(--muted);
  --fg-soft: color-mix(in srgb, var(--muted) 82%, transparent);
  --border-subtle: color-mix(in srgb, #cbd5e1 72%, transparent);
  --border-faint: color-mix(in srgb, #e2e8f0 88%, transparent);
  --border-strong: color-mix(in srgb, #64748b 24%, var(--border-dim));
  --accent-soft: color-mix(in srgb, var(--accent) 35%, transparent);
  --elev: var(--surface);
  --elev-2: var(--surface-2);
  --pill-hover: color-mix(in srgb, var(--muted) 8%, transparent);
}

:root {
  font-family: "DM Sans", system-ui, -apple-system, "Segoe UI", sans-serif;
  line-height: 1.5;
  background-color: var(--bg);
  color: var(--ink);
}

*,
*::before,
*::after {
  box-sizing: border-box;
}

body {
  margin: 0;
  min-height: 100vh;
  background: radial-gradient(
      ellipse 120% 80% at 50% -20%,
      color-mix(in srgb, #3730a3 34%, transparent),
      transparent 55%
    ),
    radial-gradient(
      ellipse 70% 50% at 100% 100%,
      color-mix(in srgb, #1e3a5f 25%, transparent),
      transparent 45%
    ),
    var(--bg);
}

[data-theme="light"] body {
  background: radial-gradient(
      ellipse 120% 80% at 50% -20%,
      color-mix(in srgb, #c7d2fe 55%, transparent),
      transparent 55%
    ),
    radial-gradient(
      ellipse 70% 50% at 100% 100%,
      color-mix(in srgb, #dbeafe 45%, transparent),
      transparent 45%
    ),
    var(--bg);
}

body::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background-image:
    linear-gradient(
      color-mix(in srgb, var(--accent) 6%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--accent) 6%, transparent) 1px,
      transparent 1px
    );
  background-size: 48px 48px;
  mask-image: radial-gradient(ellipse 85% 70% at 50% 40%, black 15%, transparent 75%);
  opacity: 0.55;
}

[data-theme="light"] body::before {
  opacity: 0.22;
}

#app {
  position: relative;
  z-index: 1;
  min-height: 100vh;
}

a {
  color: var(--accent);
  text-decoration: none;
  transition: color 0.15s ease, text-shadow 0.15s ease;
}

a:hover {
  color: color-mix(in srgb, var(--accent) 85%, white);
  text-shadow: 0 0 20px var(--accent-glow);
  text-decoration: none;
}

*:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}
</style>
