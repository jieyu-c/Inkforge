<script setup lang="ts">
import LocaleSwitcher from "@/components/LocaleSwitcher.vue";
import InkforgeLogoMark from "@/components/InkforgeLogoMark.vue";
import { ApiError } from "@/api/client";
import { me, type MeResp } from "@/api/user";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import { useAuthStore } from "@/stores/auth";
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const route = useRoute();
const vueRouter = useRouter();
const auth = useAuthStore();
const wsCtx = useWorkspaceContextStore();

const profile = ref<MeResp | null>(null);
const profileError = ref<string | null>(null);
const profileLoading = ref(false);
const logoutBusy = ref(false);

const pageTitle = computed(() => {
  const key = route.meta.titleI18nKey;
  return typeof key === "string" ? t(key) : t("console.shell.defaultTitle");
});

const userLine = computed(() => {
  if (profileLoading.value) return t("console.shell.profileLoading");
  if (profileError.value) return t("console.shell.profileShort");
  const p = profile.value;
  if (!p) return t("console.shell.profileShort");
  const phone = p.phone.length > 7 ? `${p.phone.slice(0, 3)}…${p.phone.slice(-4)}` : p.phone;
  return phone;
});

function truncId(id: string) {
  return id.length > 12 ? `${id.slice(0, 6)}…${id.slice(-4)}` : id;
}

async function loadProfile() {
  profileLoading.value = true;
  profileError.value = null;
  try {
    profile.value = await me();
  } catch (e) {
    profile.value = null;
    if (!(e instanceof ApiError && e.status === 401)) {
      profileError.value = t("home.profileError");
    }
  } finally {
    profileLoading.value = false;
  }
}

async function onLogout() {
  logoutBusy.value = true;
  try {
    await auth.logout();
    profile.value = null;
    await vueRouter.replace({ name: "login" });
  } finally {
    logoutBusy.value = false;
  }
}

onMounted(() => {
  void loadProfile();
  void wsCtx.reloadNamespaces().catch(() => {});
});
</script>

<template>
  <div class="shell">
    <a class="skip" href="#console-main">{{ t("console.shell.skipMain") }}</a>

    <nav class="rail" :aria-label="t('console.shell.railAria')">
      <div class="rail-brand">
        <RouterLink class="brand-link" :to="{ name: 'workspace' }">
          <span class="brand-mark-wrap" aria-hidden="true">
            <InkforgeLogoMark class="brand-mark" />
          </span>
          <span class="brand-word">{{ t("common.brandConsoleShort") }}</span>
        </RouterLink>
      </div>

      <p class="rail-group">{{ t("console.nav.groupWork") }}</p>
      <ul class="rail-list">
        <li>
          <RouterLink
            class="rail-link"
            :class="{ 'is-active': route.name === 'workspace' }"
            :to="{ name: 'workspace' }"
          >
            {{ t("console.nav.dashboard") }}
          </RouterLink>
        </li>
        <li>
          <RouterLink
            class="rail-link"
            :class="{ 'is-active': route.name === 'console-prompts' }"
            :to="{ name: 'console-prompts' }"
          >
            {{ t("console.nav.prompts") }}
          </RouterLink>
        </li>
      </ul>

      <p class="rail-group">{{ t("console.nav.groupAssist") }}</p>
      <ul class="rail-list">
        <li>
          <span
            class="rail-placeholder"
            tabindex="-1"
            role="presentation"
            :title="t('console.nav.tooltipSoon')"
          >
            <span>{{ t("console.nav.inkscribe") }}</span>
            <span class="rail-badge">{{ t("console.nav.badgeSoon") }}</span>
          </span>
        </li>
      </ul>

      <p class="rail-group">{{ t("console.nav.groupExtend") }}</p>
      <ul class="rail-list">
        <li>
          <span class="rail-placeholder" tabindex="-1" role="presentation" :title="t('console.nav.skillHint')">
            <span>{{ t("console.nav.skill") }}</span>
            <span class="rail-badge">{{ t("console.nav.badgeSoon") }}</span>
          </span>
        </li>
        <li>
          <span class="rail-placeholder" tabindex="-1" role="presentation" :title="t('console.nav.mcpHint')">
            <span>{{ t("console.nav.mcp") }}</span>
            <span class="rail-badge">{{ t("console.nav.badgeSoon") }}</span>
          </span>
        </li>
      </ul>

      <p class="rail-group">{{ t("console.nav.groupAccount") }}</p>
      <ul class="rail-list">
        <li>
          <RouterLink
            class="rail-link"
            :class="{ 'is-active': route.name === 'console-settings' }"
            :to="{ name: 'console-settings' }"
          >
            {{ t("console.nav.settings") }}
          </RouterLink>
        </li>
      </ul>
    </nav>

    <div class="main-stack">
      <header class="topbar">
        <div class="topbar-left">
          <h1 class="topbar-title">{{ pageTitle }}</h1>
          <div class="ns-bar" :title="wsCtx.namespacesError ?? undefined">
            <label class="ns-bar-inner" :for="'inkforge-ns-shell'">
              <span class="ns-bar-label">{{ t("workspace.topbarNsLabel") }}</span>
              <select
                id="inkforge-ns-shell"
                class="ns-select mono"
                :disabled="wsCtx.namespacesLoading"
                :aria-busy="wsCtx.namespacesLoading"
                :value="wsCtx.selectedNsSlug"
                @change="
                  wsCtx.setSelectedNsSlug(($event.target as HTMLSelectElement).value)
                "
              >
                <option value="">{{ t("workspace.topbarNsEmpty") }}</option>
                <option v-for="n in wsCtx.namespaces" :key="n.ns_slug" :value="n.ns_slug">
                  {{ n.display_name }} · {{ n.ns_slug }}
                </option>
              </select>
              <span v-if="wsCtx.namespacesLoading" class="ns-bar-status">{{
                t("workspace.topbarNsLoading")
              }}</span>
              <span v-else-if="wsCtx.namespacesError" class="ns-bar-status err">{{
                t("workspace.topbarNsError")
              }}</span>
            </label>
          </div>
        </div>
        <div class="topbar-right">
          <LocaleSwitcher variant="inline" class="locale" />
          <div
            v-if="profile"
            class="user-meta"
            :title="`${profile.user_id} · ${profile.phone}`"
          >
            <span class="user-id mono">{{ truncId(profile.user_id) }}</span>
            <span class="user-phone mono">{{ userLine }}</span>
          </div>
          <div v-else class="user-meta muted" :title="profileError ?? undefined">
            {{ userLine }}
          </div>
          <button
            type="button"
            class="btn-outline"
            :disabled="logoutBusy"
            @click="onLogout"
          >
            {{ t("common.signOut") }}
          </button>
        </div>
      </header>

      <main id="console-main" class="main-pane" tabindex="-1">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 238px minmax(0, 1fr);
}

@media (max-width: 780px) {
  .shell {
    grid-template-columns: 1fr;
  }

  .rail {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    border-right: none;
    border-bottom: 1px solid var(--border-subtle);
  }

  .rail-group {
    display: none;
  }

  .rail-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    width: 100%;
    margin-bottom: 0.5rem;
  }
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

.rail {
  padding: 1rem 0.85rem;
  border-right: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 94%, transparent);
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.rail-brand {
  padding: 0.25rem 0.35rem 0.85rem;
}

.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  font-size: 0.95rem;
  letter-spacing: -0.02em;
  color: var(--fg);
  text-decoration: none;
}
.brand-link:hover {
  color: var(--accent);
}

.brand-mark-wrap {
  display: inline-flex;
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
}

.brand-mark {
  width: 100%;
  height: 100%;
}

.rail-group {
  margin: 0.85rem 0.35rem 0.3rem;
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--fg-soft);
}

.rail-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.rail-link {
  display: flex;
  align-items: center;
  padding: 0.42rem 0.65rem;
  border-radius: 8px;
  font-size: 0.8675rem;
  font-weight: 550;
  color: var(--fg-muted);
  text-decoration: none;
  transition: background 0.12s ease, color 0.12s ease;
}

.rail-link:hover {
  background: var(--pill-hover);
  color: var(--fg);
}

.rail-link.is-active {
  background: color-mix(in srgb, var(--accent) 16%, transparent);
  color: var(--fg);
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 28%, transparent);
}

.rail-placeholder {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  padding: 0.42rem 0.65rem;
  border-radius: 8px;
  font-size: 0.8675rem;
  color: var(--fg-soft);
  cursor: not-allowed;
  opacity: 0.72;
}

.rail-badge {
  font-size: 0.62rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.12rem 0.35rem;
  border-radius: 6px;
  background: var(--elev-2);
  border: 1px solid var(--border-faint);
  color: var(--fg-muted);
}

.main-stack {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 100vh;
}

.topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.85rem 1.35rem;
  border-bottom: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--bg-canvas) 85%, transparent);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.topbar-title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.topbar-left {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.65rem 1.35rem;
  min-width: 0;
}

.ns-bar-inner {
  display: grid;
  grid-template-columns: minmax(0, auto);
  gap: 0.2rem;
  min-width: 0;
}

.ns-bar-label {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-soft);
}

.ns-select {
  min-width: min(420px, 78vw);
  max-width: 100%;
  padding: 0.32rem 0.52rem;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--elev-2);
  color: var(--fg);
  font-size: 0.7825rem;
}

.ns-select:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.ns-bar-status {
  font-size: 0.7rem;
  color: var(--fg-soft);
}
.ns-bar-status.err {
  color: color-mix(in srgb, #f97316 55%, var(--fg-muted));
}

.topbar-right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.55rem;
}

.locale {
  flex-shrink: 0;
}

.user-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.05rem;
  max-width: 14rem;
  font-size: 0.75rem;
  color: var(--fg-muted);
}

.user-meta.muted {
  font-style: italic;
}

.user-id,
.user-phone {
  font-size: 0.72rem;
}

.mono {
  font-family: "JetBrains Mono", ui-monospace, monospace;
}

.btn-outline {
  font: inherit;
  cursor: pointer;
  border-radius: 9px;
  padding: 0.38rem 0.75rem;
  font-weight: 600;
  font-size: 0.8rem;
  background: var(--elev-2);
  color: var(--fg);
  border: 1px solid var(--border-strong);
  transition: border-color 0.12s ease, background 0.12s ease;
}
.btn-outline:hover:not(:disabled) {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 10%, var(--elev-2));
}
.btn-outline:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.main-pane {
  flex: 1;
  padding: 1.35rem 1.35rem 2.5rem;
  outline: none;
}
</style>
