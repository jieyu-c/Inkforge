<script setup lang="ts">
import { ApiError } from "@/api/client";
import { me, type MeResp } from "@/api/user";
import LocaleSwitcher from "@/components/LocaleSwitcher.vue";
import ThemeSwitcher from "@/components/ThemeSwitcher.vue";
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const profile = ref<MeResp | null>(null);
const loadError = ref<string | null>(null);
const loading = ref(false);

async function loadProfile() {
  loading.value = true;
  loadError.value = null;
  try {
    profile.value = await me();
  } catch (e) {
    profile.value = null;
    if (e instanceof ApiError && e.status === 401) {
      loadError.value = t("home.profileError");
    } else {
      loadError.value =
        e instanceof ApiError ? `${e.code}: ${e.message}` : t("home.profileError");
    }
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void loadProfile();
});
</script>

<template>
  <div class="page">
    <h1 class="title">{{ t("workspace.settingsPageTitle") }}</h1>
    <p class="lede">{{ t("workspace.settingsIntro") }}</p>

    <section class="card" aria-labelledby="account-heading">
      <h2 id="account-heading" class="card-k">{{ t("workspace.settingsAccountHeading") }}</h2>

      <p v-if="loading" class="state">{{ t("workspace.settingsLoading") }}</p>

      <template v-else-if="profile">
        <dl class="meta">
          <dt>{{ t("workspace.settingsUserId") }}</dt>
          <dd class="mono">{{ profile.user_id }}</dd>
          <dt>{{ t("workspace.settingsPhone") }}</dt>
          <dd class="mono">{{ profile.phone }}</dd>
        </dl>
      </template>

      <div v-else class="err-block">
        <p class="err-text">{{ loadError }}</p>
        <button type="button" class="btn-retry" :disabled="loading" @click="loadProfile">
          {{ t("workspace.settingsRetry") }}
        </button>
      </div>
    </section>

    <section class="card" aria-labelledby="appearance-heading">
      <h2 id="appearance-heading" class="card-k">
        {{ t("workspace.settingsAppearanceHeading") }}
      </h2>

      <div class="pref-list">
        <div class="pref-row">
          <span class="pref-label">{{ t("workspace.settingsLanguageLabel") }}</span>
          <LocaleSwitcher />
        </div>

        <div class="pref-row">
          <span class="pref-label">{{ t("workspace.settingsThemeLabel") }}</span>
          <ThemeSwitcher />
        </div>
      </div>
    </section>

    <p class="team">{{ t("workspace.settingsTeamHint") }}</p>

    <RouterLink class="back" :to="{ name: 'workspace' }">{{
      t("workspace.backToDashboard")
    }}</RouterLink>
  </div>
</template>

<style scoped>
.page {
  max-width: 42rem;
}

.title {
  margin: 0 0 0.35rem;
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.lede {
  margin: 0 0 1.15rem;
  color: var(--fg-muted);
  font-size: 0.875rem;
  line-height: 1.5;
}

.card {
  margin-bottom: 1.15rem;
  padding: 1rem 1.1rem;
  border-radius: 12px;
  border: 1px solid var(--border-subtle);
  background: var(--elev);
}

.card-k {
  margin: 0 0 0.75rem;
  font-size: 0.6875rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--accent);
}

.pref-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.pref-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1rem;
}

.pref-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--fg);
}

.state {
  margin: 0;
  font-size: 0.875rem;
  color: var(--fg-muted);
}

.meta {
  display: grid;
  grid-template-columns: minmax(0, 9rem) 1fr;
  gap: 0.45rem 0.85rem;
  margin: 0;
  font-size: 0.875rem;
}

.meta dt {
  margin: 0;
  color: var(--fg-soft);
  font-weight: 600;
}

.meta dd {
  margin: 0;
  color: var(--fg);
  word-break: break-all;
}

.mono {
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 0.8125rem;
}

.err-block {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.65rem;
}

.err-text {
  margin: 0;
  font-size: 0.875rem;
  color: color-mix(in srgb, #f97316 55%, var(--fg-muted));
  line-height: 1.45;
}

.btn-retry {
  cursor: pointer;
  border-radius: 9px;
  padding: 0.4rem 0.85rem;
  font-weight: 600;
  font-size: 0.8rem;
  font-family: inherit;
  background: var(--elev-2);
  color: var(--fg);
  border: 1px solid var(--border-strong);
}

.btn-retry:hover:not(:disabled) {
  border-color: var(--accent-soft);
  background: color-mix(in srgb, var(--accent) 10%, var(--elev-2));
}

.btn-retry:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.team {
  margin: 0 0 1.15rem;
  padding: 0.75rem 0.95rem;
  border-radius: 10px;
  border: 1px solid var(--border-subtle);
  background: color-mix(in srgb, var(--elev-2) 92%, transparent);
  font-size: 0.8675rem;
  color: var(--fg-muted);
  line-height: 1.5;
}

.back {
  font-weight: 600;
  font-size: 0.9rem;
}
</style>
