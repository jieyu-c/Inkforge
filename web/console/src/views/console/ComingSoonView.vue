<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const { t } = useI18n();
const route = useRoute();

const title = computed(() => {
  const key = route.meta.titleI18nKey;
  return typeof key === "string" ? t(key) : t("workspace.comingSoonDefaultTitle");
});

const settingsTeamHint = computed(() => route.name === "console-settings");
</script>

<template>
  <div class="wrap">
    <h1 class="title">{{ title }}</h1>
    <p class="muted">{{ t("workspace.comingSoonBody") }}</p>
    <p v-if="settingsTeamHint" class="team">{{ t("workspace.settingsTeamHint") }}</p>
    <RouterLink class="back" :to="{ name: 'workspace' }">{{
      t("workspace.backToDashboard")
    }}</RouterLink>
  </div>
</template>

<style scoped>
.wrap {
  max-width: 42rem;
}

.title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.muted {
  margin: 0 0 1.25rem;
  color: var(--fg-muted);
  line-height: 1.55;
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
