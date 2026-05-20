<script setup lang="ts">
import { ApiError } from "@/api/client";
import { useAuthStore } from "@/stores/auth";
import {
  canonicalPhone,
  MIN_PASSWORD_LENGTH,
} from "@/lib/phone";
import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

const auth = useAuthStore();
const { t, locale } = useI18n();
const router = useRouter();
const route = useRoute();

const phoneRaw = ref("");
const password = ref("");
const error = ref<string | null>(null);
const banner = ref<string | null>(null);
const busy = ref(false);

watch(
  () => [route.query.registered, locale.value] as const,
  ([registered]) => {
    banner.value =
      registered === "1" ? t("login.bannerRegistered") : null;
  },
  { immediate: true },
);

async function submit() {
  error.value = null;
  const phone = canonicalPhone(phoneRaw.value);
  if (!phone) {
    error.value = t("login.invalidPhone");
    return;
  }
  if (password.value.length < MIN_PASSWORD_LENGTH) {
    error.value = t("login.passwordMin", { n: MIN_PASSWORD_LENGTH });
    return;
  }

  busy.value = true;
  try {
    await auth.login(phone, password.value);
    const redirect = route.query.redirect;
    const path =
      typeof redirect === "string" && redirect.startsWith("/")
        ? redirect
        : "/workspace";
    await router.replace(path);
  } catch (e) {
    if (e instanceof ApiError) {
      error.value = e.message;
    } else {
      error.value = t("login.genericError");
    }
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="shell">
    <div class="panel">
      <div class="glow" aria-hidden="true" />
      <div class="card">
        <p class="eyebrow">{{ t("common.brandEyebrow") }}</p>
        <h1>{{ t("login.title") }}</h1>
        <p class="muted">{{ t("login.subtitle") }}</p>

      <form class="form" @submit.prevent="submit">
        <label class="field">
          <span>{{ t("common.phone") }}</span>
          <input
            v-model="phoneRaw"
            type="text"
            autocomplete="username"
            inputmode="tel"
            :placeholder="t('login.placeholderPhone')"
          />
        </label>
        <label class="field">
          <span>{{ t("common.password") }}</span>
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
          />
        </label>

        <p v-if="banner" class="ok">{{ banner }}</p>

        <p v-if="error" class="err">{{ error }}</p>

        <button type="submit" class="btn primary" :disabled="busy">
          {{ busy ? t("login.submitBusy") : t("login.submit") }}
        </button>
      </form>

      <p class="footer">
        {{ t("login.footerLead") }}
        <RouterLink :to="{ name: 'register' }">{{
          t("login.footerLink")
        }}</RouterLink>
      </p>
    </div>
    </div>
  </div>
</template>

<style scoped>
.shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}

.panel {
  position: relative;
  width: 100%;
  max-width: 22rem;
}

.glow {
  position: absolute;
  z-index: 0;
  left: 50%;
  top: 35%;
  width: 140%;
  height: 100%;
  transform: translate(-50%, -50%);
  background: radial-gradient(
    ellipse closest-side,
    color-mix(in srgb, var(--accent) 28%, transparent),
    transparent 72%
  );
  filter: blur(36px);
  pointer-events: none;
  opacity: 0.9;
}

.card {
  position: relative;
  z-index: 1;
  width: 100%;
  background: var(--surface);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 2rem 1.75rem;
  box-shadow:
    var(--shadow),
    inset 0 1px 0 color-mix(in srgb, #fff 8%, transparent);
}

.eyebrow {
  margin: 0 0 0.35rem;
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.7rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--accent);
  opacity: 0.95;
}

h1 {
  margin: 0 0 0.35rem;
  font-size: 1.55rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  background: linear-gradient(
    120deg,
    var(--ink) 0%,
    color-mix(in srgb, var(--accent) 55%, var(--ink)) 100%
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.muted {
  margin: 0 0 1.5rem;
  color: var(--muted);
  font-size: 0.9rem;
  line-height: 1.5;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  font-size: 0.86rem;
}

.field span {
  color: var(--muted);
  font-weight: 500;
}

input {
  font: inherit;
  padding: 0.6rem 0.75rem;
  border-radius: 10px;
  border: 1px solid var(--border-dim);
  background: var(--surface-2);
  color: var(--ink);
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

input::placeholder {
  color: color-mix(in srgb, var(--muted) 65%, transparent);
}

input:hover {
  border-color: color-mix(in srgb, var(--accent) 35%, var(--border-dim));
}

input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow:
    0 0 0 3px var(--accent-glow),
    0 0 28px color-mix(in srgb, var(--accent) 18%, transparent);
}

.err {
  margin: 0;
  font-size: 0.86rem;
  color: var(--danger);
}

.ok {
  margin: 0;
  font-size: 0.86rem;
  color: var(--ok);
}

.btn {
  font: inherit;
  margin-top: 0.25rem;
  cursor: pointer;
  border-radius: 10px;
  padding: 0.65rem 1rem;
  border: none;
  font-weight: 600;
  transition:
    filter 0.15s ease,
    box-shadow 0.15s ease,
    transform 0.1s ease;
}

.btn.primary {
  background: linear-gradient(135deg, #818cf8 0%, #6366f1 48%, #4f46e5 100%);
  color: #fafafa;
  box-shadow:
    0 0 0 1px color-mix(in srgb, #fff 22%, transparent),
    0 4px 24px color-mix(in srgb, var(--accent) 35%, transparent);
}

.btn.primary:hover:not(:disabled) {
  filter: brightness(1.06);
  box-shadow:
    0 0 0 1px color-mix(in srgb, #fff 28%, transparent),
    0 6px 32px color-mix(in srgb, var(--accent) 45%, transparent);
}

.btn.primary:active:not(:disabled) {
  transform: translateY(1px);
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  filter: grayscale(0.2);
}

.footer {
  margin: 1.5rem 0 0;
  font-size: 0.88rem;
  color: var(--muted);
}

.footer a {
  font-weight: 600;
}
</style>
