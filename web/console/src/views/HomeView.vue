<script setup lang="ts">
import { ApiError } from "@/api/client";
import LocaleSwitcher from "@/components/LocaleSwitcher.vue";
import InkforgeLogoMark from "@/components/InkforgeLogoMark.vue";
import { me, type MeResp } from "@/api/user";
import { useAuthStore } from "@/stores/auth";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const auth = useAuthStore();
const { t } = useI18n();

const loading = ref(true);
const profile = ref<MeResp | null>(null);
const loadError = ref<string | null>(null);
/** True when /me returned 401 — show sign-in prompt, no redirect. */
const unauthenticated = ref(false);
const busy = ref(false);

const docBase =
  typeof import.meta.env.VITE_DOCS_URL === "string" &&
  import.meta.env.VITE_DOCS_URL.length > 0
    ? import.meta.env.VITE_DOCS_URL.replace(/\/$/, "")
    : "";

onMounted(async () => {
  loading.value = true;
  loadError.value = null;
  unauthenticated.value = false;
  profile.value = null;
  try {
    profile.value = await me();
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      auth.setAccess(null);
      unauthenticated.value = true;
    } else {
      loadError.value =
        e instanceof ApiError ? e.message : t("home.profileError");
    }
  } finally {
    loading.value = false;
  }
});

async function onLogout() {
  busy.value = true;
  try {
    await auth.logout();
    profile.value = null;
    unauthenticated.value = true;
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="page">
    <a class="skip" href="#session">{{ t("home.skipToSession") }}</a>

    <nav class="nav" :aria-label="t('home.navPrimary')">
      <div class="nav-inner">
        <RouterLink class="logo-link" :to="{ name: 'home' }">
          <span class="logo-mark-wrap" aria-hidden="true">
            <InkforgeLogoMark class="logo-mark" />
          </span>
          <span class="logo-word">Inkforge</span>
        </RouterLink>

        <div class="nav-links">
          <RouterLink
            v-if="auth.isAuthenticated"
            class="nav-a"
            :to="{ name: 'workspace' }"
          >
            {{ t("home.navWorkspace") }}
          </RouterLink>
          <a class="nav-a" href="#features">{{ t("home.navFeatures") }}</a>
          <a class="nav-a" href="#session">{{ t("home.navSession") }}</a>
          <a
            class="nav-a"
            :href="docBase ? `${docBase}/inkforge.md` : '#features'"
            :rel="docBase ? 'noreferrer noopener' : undefined"
            :target="docBase ? '_blank' : undefined"
          >
            {{ t("home.navDocs") }}
          </a>
        </div>

        <LocaleSwitcher variant="inline" class="locale-in-nav" />

        <div class="nav-actions">
          <template v-if="auth.isAuthenticated">
            <button
              type="button"
              class="btn btn-outline"
              :disabled="busy"
              @click="onLogout"
            >
              {{ t("common.signOut") }}
            </button>
          </template>
          <template v-else>
            <RouterLink class="btn btn-ghost" :to="{ name: 'login' }">
              {{ t("common.signIn") }}
            </RouterLink>
            <RouterLink class="btn btn-primary" :to="{ name: 'register' }">
              {{ t("common.register") }}
            </RouterLink>
          </template>
        </div>
      </div>
    </nav>

    <header class="hero">
      <div class="hero-grid" aria-hidden="true">
        <div class="hero-orb hero-orb-a" />
        <div class="hero-orb hero-orb-b" />
      </div>
      <div class="hero-inner">
        <p class="hero-eyebrow">{{ t("home.heroEyebrow") }}</p>
        <h1 class="hero-title">
          {{ t("home.heroLine1") }}
          <span class="hero-title-accent">{{ t("home.heroLine2") }}</span>
        </h1>
        <p class="hero-lede">
          {{ t("home.heroLead") }}
        </p>
        <div class="hero-cta">
          <RouterLink
            v-if="unauthenticated || !auth.isAuthenticated"
            class="btn btn-primary btn-large"
            :to="{ name: 'login' }"
          >
            {{ t("home.ctaSignInConsole") }}
          </RouterLink>
          <RouterLink
            v-else
            class="btn btn-primary btn-large"
            :to="{ name: 'workspace' }"
          >
            {{ t("home.ctaOpenWorkspace") }}
          </RouterLink>
          <a class="btn btn-outline btn-large" href="#features">
            {{ t("home.ctaExplore") }}
          </a>
        </div>
        <ul class="hero-meta" :aria-label="t('home.metaLabel')">
          <li><span class="dot" aria-hidden="true" />{{ t("home.metaGoZero") }}</li>
          <li><span class="dot" aria-hidden="true" />{{ t("home.metaSession") }}</li>
          <li><span class="dot" aria-hidden="true" />{{ t("home.metaVue") }}</li>
        </ul>
      </div>
    </header>

    <section id="features" class="section features">
      <div class="section-head">
        <h2>{{ t("home.sectionFeaturesTitle") }}</h2>
        <p>
          {{ t("home.sectionFeaturesLead") }}
        </p>
      </div>
      <div class="feature-grid">
        <article class="feature-card">
          <div class="feature-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
              <path
                d="M4 7a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H6a2 2 0 01-2-2V7z"
                stroke="currentColor"
                stroke-width="1.5"
              />
              <path
                d="M8 11h8M8 15h5"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
              />
            </svg>
          </div>
          <h3>{{ t("home.feat1Title") }}</h3>
          <p>
            {{ t("home.feat1Body") }}
          </p>
        </article>
        <article class="feature-card">
          <div class="feature-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
              <path
                d="M12 3l8 4v6c0 5-3.5 9-8 10-4.5-1-8-5-8-10V7l8-4z"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linejoin="round"
              />
              <path
                d="M9 12l2 2 4-5"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <h3>{{ t("home.feat2Title") }}</h3>
          <p>
            {{ t("home.feat2Body") }}
          </p>
        </article>
        <article class="feature-card">
          <div class="feature-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
              <path
                d="M13 4H7a3 3 0 00-3 3v10a3 3 0 003 3h10a3 3 0 003-3v-6"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linecap="round"
              />
              <path
                d="M21 7l-7.5 7.5L10 17l-.5-3.5L17 7h4z"
                stroke="currentColor"
                stroke-width="1.5"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <h3>{{ t("home.feat3Title") }}</h3>
          <p>
            {{ t("home.feat3Body") }}
          </p>
        </article>
      </div>
    </section>

    <section id="session" class="section session-section">
      <div class="section-head">
        <h2>{{ t("home.sessionTitle") }}</h2>
        <p class="session-lead">
          {{ t("home.sessionLeadPrefix") }}
          <code>/me</code>
          {{ t("home.sessionLeadSuffix") }}
        </p>
      </div>

      <div class="session-layout">
        <aside class="code-rail" aria-hidden="true">
          <pre class="code-snippet"><code><span class="tok-k">curl</span> -sSf <span class="tok-s">\</span>
  -H <span class="tok-s">"Authorization: Bearer &lt;access&gt;"</span> <span class="tok-s">\</span>
  <span class="tok-s">"&lt;/me endpoint&gt;"</span></code></pre>
        </aside>

        <div class="session-panel">
          <p v-if="loadError" class="err">{{ loadError }}</p>

          <div v-else-if="unauthenticated" class="session-callout">
            <div>
              <h3>{{ t("home.signedOutTitle") }}</h3>
              <p class="muted">
                {{ t("home.signedOutBody") }}
              </p>
            </div>
            <RouterLink class="btn btn-primary" :to="{ name: 'login' }">
              {{ t("common.signIn") }}
            </RouterLink>
          </div>

          <dl v-else-if="profile" class="profile-grid">
            <div class="profile-row">
              <dt>{{ t("home.profileUserId") }}</dt>
              <dd class="mono">{{ profile.user_id }}</dd>
            </div>
            <div class="profile-row">
              <dt>{{ t("home.profilePhone") }}</dt>
              <dd class="mono">{{ profile.phone }}</dd>
            </div>
          </dl>

          <p v-else-if="loading" class="muted loading">{{ t("home.loadingProfile") }}</p>
        </div>
      </div>
    </section>

    <footer class="footer">
      <div class="footer-inner">
        <span class="footer-brand">Inkforge</span>
        <span class="footer-note">{{ t("home.footerTagline") }}</span>
        <span class="footer-meta">{{
          t("home.footerCopyright", { year: new Date().getFullYear() })
        }}</span>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.skip {
  position: absolute;
  left: -9999px;
  top: 0;
  padding: 0.5rem 1rem;
  background: var(--accent);
  color: #fafafa;
  font-weight: 600;
  z-index: 100;
}
.skip:focus {
  left: 1rem;
  top: 1rem;
}

/* —— Nav —— */
.nav {
  position: sticky;
  top: 0;
  z-index: 50;
  border-bottom: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--bg-canvas) 72%, transparent);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
}

.nav-inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0.65rem 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo-link {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  font-weight: 700;
  font-size: 1rem;
  letter-spacing: -0.02em;
  color: var(--fg);
  text-decoration: none;
}
.logo-link:hover {
  color: var(--accent);
  text-decoration: none;
  text-shadow: none;
}

.logo-mark-wrap {
  display: inline-flex;
  width: 1.625rem;
  height: 1.625rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  color: inherit;
}

.logo-mark-wrap .logo-mark {
  width: 100%;
  height: 100%;
}

.logo-word {
  font-variant-numeric: tabular-nums;
}

.nav-links {
  display: none;
  flex: 1;
  gap: 0.35rem;
  margin-left: 1.75rem;
}
@media (min-width: 780px) {
  .nav-links {
    display: flex;
  }
}

.nav-a {
  padding: 0.42rem 0.72rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--fg-muted);
  text-decoration: none;
  transition: background 0.12s ease, color 0.12s ease;
}
.nav-a:hover {
  background: var(--pill-hover);
  color: var(--fg);
  text-decoration: none;
  text-shadow: none;
}

.nav-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-shrink: 0;
}

.locale-in-nav {
  flex-shrink: 0;
}

/* Shared buttons —— */
.btn {
  font: inherit;
  cursor: pointer;
  border: none;
  border-radius: 10px;
  padding: 0.45rem 0.95rem;
  font-weight: 600;
  font-size: 0.84rem;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  transition:
    transform 0.12s ease,
    box-shadow 0.12s ease,
    filter 0.12s ease,
    background 0.12s ease,
    border-color 0.12s ease;
}
.btn:active:not(:disabled) {
  transform: translateY(1px);
}

.btn-large {
  padding: 0.72rem 1.25rem;
  font-size: 0.95rem;
}

.btn-primary {
  background: linear-gradient(
    145deg,
    #818cf8 0%,
    #6366f1 52%,
    #4f46e5 100%
  );
  color: #fafafa;
  border: 1px solid color-mix(in srgb, #fff 42%, transparent);
  box-shadow:
    0 1px 0 color-mix(in srgb, #fff 42%, transparent) inset,
    0 3px 20px color-mix(in srgb, var(--accent) 42%, transparent);
}
.btn-primary:hover {
  filter: brightness(1.06);
  text-decoration: none;
}

.btn-outline {
  background: var(--elev-2);
  color: var(--fg);
  border: 1px solid var(--border-strong);
}
.btn-outline:hover {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 12%, var(--elev-2));
  text-decoration: none;
}

.btn-ghost {
  background: transparent;
  color: var(--fg-muted);
}
.btn-ghost:hover {
  color: var(--fg);
  background: var(--pill-hover);
}

.btn-outline:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* —— Hero —— */
.hero {
  position: relative;
  padding: clamp(2.75rem, 6vw, 4.75rem) 1.25rem 3.5rem;
  overflow: hidden;
}

.hero-grid {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.hero-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.42;
}
.hero-orb-a {
  width: min(560px, 90vw);
  height: min(560px, 90vw);
  background: radial-gradient(
    circle,
    color-mix(in srgb, var(--accent) 55%, transparent),
    transparent 68%
  );
  top: -18%;
  right: -12%;
}
.hero-orb-b {
  width: min(420px, 70vw);
  height: min(420px, 70vw);
  background: radial-gradient(circle, #6366f1 38%, transparent 70%);
  bottom: -20%;
  left: -14%;
  opacity: 0.26;
}

.hero-inner {
  position: relative;
  max-width: 1120px;
  margin: 0 auto;
}

.hero-eyebrow {
  margin: 0 0 0.75rem;
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.73rem;
  font-weight: 500;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--accent);
}

.hero-title {
  margin: 0;
  font-size: clamp(2rem, 4.8vw, 3.05rem);
  font-weight: 700;
  line-height: 1.07;
  letter-spacing: -0.035em;
  max-width: min(42rem, 100%);
}

.hero-title-accent {
  display: inline;
  background: linear-gradient(
    118deg,
    var(--accent) 0%,
    #a5b4fc 40%,
    #e0e7ff 92%
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.hero-lede {
  margin: 1.15rem 0 0;
  max-width: 52ch;
  font-size: 1.0625rem;
  line-height: 1.55;
  color: var(--fg-muted);
}

.hero-cta {
  margin-top: 1.75rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
}

.hero-meta {
  margin: 2.25rem 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-wrap: wrap;
  gap: 1rem 1.5rem;
  font-size: 0.8125rem;
  color: var(--fg-soft);
}

.hero-meta li {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: linear-gradient(
    180deg,
    var(--accent) 0%,
    color-mix(in srgb, var(--accent) 45%, #a5b4fc) 100%
  );
  box-shadow: 0 0 10px color-mix(in srgb, var(--accent) 55%, transparent);
}

/* Sections */
.section {
  max-width: 1120px;
  margin: 0 auto;
  padding: 2rem 1.25rem 2.75rem;
  scroll-margin-top: 76px;
}

.section-head {
  margin-bottom: 1.85rem;
  max-width: 56ch;
}
.section-head h2 {
  margin: 0 0 0.45rem;
  font-size: 1.375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}
.section-head p {
  margin: 0;
  color: var(--fg-muted);
  line-height: 1.52;
}

.session-lead code {
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.875em;
  padding: 0.08em 0.35em;
  border-radius: 6px;
  background: var(--elev-2);
  border: 1px solid var(--border-faint);
  color: color-mix(in srgb, var(--accent) 35%, var(--fg));
}

.feature-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr;
}
@media (min-width: 720px) {
  .feature-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.feature-card {
  padding: 1.35rem 1.3rem;
  border-radius: 14px;
  border: 1px solid var(--border-subtle);
  background: var(--elev);
  transition:
    border-color 0.16s ease,
    box-shadow 0.16s ease,
    transform 0.16s ease;
}

.feature-card:hover {
  border-color: var(--border-strong);
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--accent) 18%, transparent),
    0 18px 40px rgba(0, 0, 0, 0.28);
  transform: translateY(-2px);
}

.feature-card h3 {
  margin: 0.65rem 0 0.4rem;
  font-size: 1.02rem;
  font-weight: 650;
}

.feature-card p {
  margin: 0;
  color: var(--fg-muted);
  font-size: 0.9075rem;
  line-height: 1.5;
}

.feature-icon {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(
    155deg,
    color-mix(in srgb, var(--accent) 22%, transparent) 0%,
    color-mix(in srgb, #a5b4fc 18%, transparent) 100%
  );
  border: 1px solid color-mix(in srgb, var(--accent) 38%, transparent);
  color: var(--accent);
}

.features {
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-faint);
}

/* Session block */
.session-section {
  padding-top: 0.75rem;
}

.session-layout {
  display: grid;
  gap: 1rem;
  align-items: start;
}
@media (min-width: 860px) {
  .session-layout {
    grid-template-columns: minmax(0, 340px) minmax(0, 1fr);
  }
}

.code-rail {
  border-radius: 14px;
  border: 1px solid var(--border-subtle);
  background: var(--elev-2);
  padding: 0.95rem;
  overflow: auto;
}

.code-snippet {
  margin: 0;
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.75rem;
  line-height: 1.62;
}
.tok-k {
  color: #fda4af;
}
.tok-s {
  color: color-mix(in srgb, var(--accent) 78%, white);
}

.session-panel {
  border-radius: 14px;
  border: 1px solid var(--border-subtle);
  background: var(--elev);
  padding: 1.35rem 1.25rem;
  min-height: 8rem;
}

.session-callout {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.session-callout h3 {
  margin: 0 0 0.35rem;
  font-size: 1rem;
}

.muted {
  margin: 0;
  color: var(--fg-muted);
  font-size: 0.9175rem;
  line-height: 1.5;
}

.profile-grid {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.profile-row {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  gap: 0.75rem;
  padding: 0.75rem;
  align-items: baseline;
  border-radius: 10px;
  background: var(--elev-2);
  border: 1px solid var(--border-faint);
}

.profile-row dt {
  margin: 0;
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.68rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-soft);
}

.profile-row dd {
  margin: 0;
  text-align: right;
  word-break: break-all;
  font-size: 0.9rem;
}

dd.mono {
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.86rem;
  color: color-mix(in srgb, var(--accent) 42%, var(--fg));
}

.err {
  color: var(--danger);
  margin: 0;
  font-size: 0.9175rem;
}

.loading {
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.8375rem;
  color: color-mix(in srgb, var(--accent) 70%, var(--fg-muted));
}

/* Footer */
.footer {
  margin-top: auto;
  border-top: 1px solid var(--border-subtle);
  padding: 2rem 1.25rem 2.5rem;
}

.footer-inner {
  max-width: 1120px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  align-items: flex-start;
}

@media (min-width: 640px) {
  .footer-inner {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    gap: 1rem;
  }
  .footer-note {
    margin-right: auto;
  }
}

.footer-brand {
  font-weight: 700;
  letter-spacing: -0.02em;
}
.footer-note {
  color: var(--fg-muted);
  font-size: 0.875rem;
}
.footer-meta {
  color: var(--fg-soft);
  font-size: 0.8rem;
}
</style>
