<script setup lang="ts">
import { RouterView } from "vue-router";
</script>

<template>
  <RouterView />
</template>

<style>
:root {
  color-scheme: light;
  --bg: #f3efe8;
  --surface: #fffefb;
  --surface-2: #faf7f2;
  --ink: #1c1917;
  --muted: #78716c;
  --border: #e7e1d8;
  --accent: #1d4a47;
  --danger: #b4533c;
  --ok: #3f6f46;
  --shadow: 0 18px 40px rgba(28, 25, 23, 0.06);
  font-family:
    system-ui,
    -apple-system,
    "Segoe UI",
    Roboto,
    "Helvetica Neue",
    Arial,
    sans-serif;
  line-height: 1.5;
  background: var(--bg);
  color: var(--ink);
}

*,
*::before,
*::after {
  box-sizing: border-box;
}

body {
  margin: 0;
}

a {
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
