<script setup lang="ts">
import PromptsPanel from "@/components/PromptsPanel.vue";
import { useWorkspaceContextStore } from "@/stores/workspaceContext";
import { onMounted } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const ctx = useWorkspaceContextStore();

onMounted(() => {
  void ctx.reloadNamespaces().catch(() => {
    /* surfaced in namespacesError */
  });
});
</script>

<template>
  <div class="page">
    <section class="hero" aria-labelledby="dash-welcome">
      <h2 id="dash-welcome" class="hero-title">{{ t("workspace.welcomeTitle") }}</h2>
      <p class="hero-lede">{{ t("workspace.welcomeLead") }}</p>
    </section>

    <PromptsPanel />
  </div>
</template>

<style scoped>
.page {
  max-width: 1120px;
}

.hero {
  margin-bottom: 1rem;
}

.hero-title {
  margin: 0 0 0.35rem;
  font-size: 1.25rem;
  font-weight: 750;
  letter-spacing: -0.03em;
}

.hero-lede {
  margin: 0;
  max-width: 52ch;
  color: var(--fg-muted);
  line-height: 1.5;
  font-size: 0.875rem;
}
</style>
