<script setup lang="ts">
import { ApiError } from "@/api/client";
import LocaleSwitcher from "@/components/LocaleSwitcher.vue";
import { useAuthStore } from "@/stores/auth";
import {
  canonicalPhone,
  MIN_PASSWORD_LENGTH,
} from "@/lib/phone";
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

const auth = useAuthStore();
const { t } = useI18n();
const router = useRouter();

const phoneRaw = ref("");
const password = ref("");
const password2 = ref("");
const error = ref<string | null>(null);
const busy = ref(false);

async function submit() {
  error.value = null;

  const phone = canonicalPhone(phoneRaw.value);
  if (!phone) {
    error.value = t("register.invalidPhone");
    return;
  }
  if (password.value.length < MIN_PASSWORD_LENGTH) {
    error.value = t("register.passwordMin", { n: MIN_PASSWORD_LENGTH });
    return;
  }
  if (password.value !== password2.value) {
    error.value = t("register.passwordMismatch");
    return;
  }

  busy.value = true;
  try {
    await auth.register(phone, password.value);
    await router.replace({ name: "login", query: { registered: "1" } });
  } catch (e) {
    if (e instanceof ApiError) {
      error.value = e.message;
    } else {
      error.value = t("register.genericError");
    }
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="shell">
    <LocaleSwitcher variant="corner" />
    <div class="panel">
      <div class="glow" aria-hidden="true" />
      <div class="card">
        <p class="eyebrow">{{ t("common.brandEyebrow") }}</p>
        <h1>{{ t("register.title") }}</h1>
        <p class="muted">
          {{ t("register.subtitle", { n: MIN_PASSWORD_LENGTH }) }}
        </p>

      <form class="form" @submit.prevent="submit">
        <label class="field">
          <span>{{ t("common.phone") }}</span>
          <input
            v-model="phoneRaw"
            type="text"
            autocomplete="tel"
            inputmode="tel"
            :placeholder="t('register.placeholderPhone')"
          />
        </label>
        <label class="field">
          <span>{{ t("common.password") }}</span>
          <input
            v-model="password"
            type="password"
            autocomplete="new-password"
          />
        </label>
        <label class="field">
          <span>{{ t("common.confirmPassword") }}</span>
          <input
            v-model="password2"
            type="password"
            autocomplete="new-password"
          />
        </label>

        <p v-if="error" class="err">{{ error }}</p>

        <button type="submit" class="btn primary" :disabled="busy">
          {{
            busy ? t("register.submitBusy") : t("register.submit")
          }}
        </button>
      </form>

      <p class="footer">
        {{ t("register.footerLead") }}
        <RouterLink :to="{ name: 'login' }">{{
          t("register.footerLink")
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
