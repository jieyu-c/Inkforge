<script setup lang="ts">
import { ApiError } from "@/api/client";
import { useAuthStore } from "@/stores/auth";
import {
  canonicalPhone,
  MIN_PASSWORD_LENGTH,
} from "@/lib/phone";
import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const phoneRaw = ref("");
const password = ref("");
const error = ref<string | null>(null);
const banner = ref<string | null>(null);
const busy = ref(false);

watch(
  () => route.query.registered,
  (v) => {
    banner.value = v === "1" ? "Registration succeeded. Please sign in." : null;
  },
  { immediate: true },
);

async function submit() {
  error.value = null;
  const phone = canonicalPhone(phoneRaw.value);
  if (!phone) {
    error.value = "Enter a valid mainland China mobile number.";
    return;
  }
  if (password.value.length < MIN_PASSWORD_LENGTH) {
    error.value = `Password must be at least ${MIN_PASSWORD_LENGTH} characters.`;
    return;
  }

  busy.value = true;
  try {
    await auth.login(phone, password.value);
    const redirect = route.query.redirect;
    const path =
      typeof redirect === "string" && redirect.startsWith("/")
        ? redirect
        : "/home";
    await router.replace(path);
  } catch (e) {
    if (e instanceof ApiError) {
      error.value = e.message;
    } else {
      error.value = "Could not sign in.";
    }
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="shell">
    <div class="card">
      <p class="eyebrow">Inkforge</p>
      <h1>Sign in</h1>
      <p class="muted">Phone number and password for the console.</p>

      <form class="form" @submit.prevent="submit">
        <label class="field">
          <span>Phone</span>
          <input
            v-model="phoneRaw"
            type="text"
            autocomplete="username"
            inputmode="tel"
            placeholder="11-digit mobile or +86…"
          />
        </label>
        <label class="field">
          <span>Password</span>
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
          />
        </label>

        <p v-if="banner" class="ok">{{ banner }}</p>

        <p v-if="error" class="err">{{ error }}</p>

        <button type="submit" class="btn primary" :disabled="busy">
          {{ busy ? "Signing in…" : "Sign in" }}
        </button>
      </form>

      <p class="footer">
        No account?
        <RouterLink :to="{ name: 'register' }">Create one</RouterLink>
      </p>
    </div>
  </div>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}

.card {
  width: 100%;
  max-width: 22rem;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 2rem 1.75rem;
  box-shadow: var(--shadow);
}

.eyebrow {
  margin: 0 0 0.25rem;
  font-size: 0.75rem;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--muted);
}

h1 {
  margin: 0 0 0.35rem;
  font-family: "Iowan Old Style", "Palatino Linotype", Palatino, Georgia, serif;
  font-size: 1.65rem;
  font-weight: 600;
  letter-spacing: -0.02em;
}

.muted {
  margin: 0 0 1.5rem;
  color: var(--muted);
  font-size: 0.92rem;
  line-height: 1.45;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.88rem;
}

.field span {
  color: var(--muted);
}

input {
  font: inherit;
  padding: 0.55rem 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface-2);
  color: var(--ink);
}

input:focus {
  outline: 2px solid color-mix(in srgb, var(--accent) 35%, transparent);
  border-color: var(--accent);
}

.err {
  margin: 0;
  font-size: 0.88rem;
  color: var(--danger);
}

.ok {
  margin: 0;
  font-size: 0.88rem;
  color: var(--ok);
}

.btn {
  font: inherit;
  margin-top: 0.25rem;
  cursor: pointer;
  border-radius: 8px;
  padding: 0.6rem 1rem;
  border: none;
}

.btn.primary {
  background: var(--accent);
  color: #fff;
}

.btn.primary:hover:not(:disabled) {
  filter: brightness(1.05);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.footer {
  margin: 1.5rem 0 0;
  font-size: 0.9rem;
  color: var(--muted);
}

.footer a {
  color: var(--accent);
  font-weight: 500;
}
</style>
