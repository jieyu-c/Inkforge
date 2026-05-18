<script setup lang="ts">
import { ApiError } from "@/api/client";
import { me, type MeResp } from "@/api/user";
import { useAuthStore } from "@/stores/auth";
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

const auth = useAuthStore();
const router = useRouter();

const profile = ref<MeResp | null>(null);
const loadError = ref<string | null>(null);
const busy = ref(false);

onMounted(async () => {
  loadError.value = null;
  try {
    profile.value = await me();
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      auth.setAccess(null);
      await router.replace({ name: "login" });
      return;
    }
    loadError.value =
      e instanceof ApiError ? e.message : "Could not load profile.";
  }
});

async function onLogout() {
  busy.value = true;
  try {
    await auth.logout();
    await router.replace({ name: "login" });
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="shell">
    <header class="top">
      <h1 class="brand">Inkforge</h1>
      <button
        type="button"
        class="btn ghost"
        :disabled="busy"
        @click="onLogout"
      >
        Sign out
      </button>
    </header>

    <main class="card">
      <h2>Console</h2>
      <p class="muted">Signed-in session from your phone account.</p>

      <p v-if="loadError" class="err">{{ loadError }}</p>

      <dl v-else-if="profile" class="grid">
        <dt>User ID</dt>
        <dd>{{ profile.user_id }}</dd>
        <dt>Phone</dt>
        <dd>{{ profile.phone }}</dd>
      </dl>

      <p v-else class="muted">Loading…</p>
    </main>
  </div>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  padding: 1.75rem 1.25rem 3rem;
  max-width: 40rem;
  margin: 0 auto;
}

.top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2rem;
}

.brand {
  font-family: "Iowan Old Style", "Palatino Linotype", Palatino, Georgia, serif;
  font-size: 1.35rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  margin: 0;
}

.card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.75rem 1.5rem;
  box-shadow: var(--shadow);
}

h2 {
  margin: 0 0 0.35rem;
  font-size: 1.25rem;
}

.muted {
  margin: 0 0 1.25rem;
  color: var(--muted);
  font-size: 0.95rem;
}

.grid {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.5rem 1.25rem;
  margin: 0;
  font-size: 0.95rem;
}

dt {
  margin: 0;
  color: var(--muted);
  font-weight: 500;
}

dd {
  margin: 0;
  word-break: break-all;
}

.err {
  color: var(--danger);
  margin: 0;
}

.btn {
  font: inherit;
  cursor: pointer;
  border-radius: 8px;
  padding: 0.45rem 0.9rem;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--ink);
}

.btn.ghost:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
