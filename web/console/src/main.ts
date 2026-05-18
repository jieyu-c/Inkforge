import { createApp } from "vue";
import App from "./App.vue";
import { registerAccessTokenSync } from "@/lib/token-sync";
import { router } from "@/router";
import { useAuthStore } from "@/stores/auth";
import { createPinia } from "pinia";

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
registerAccessTokenSync(() => {
  useAuthStore().syncFromStorage();
});
app.use(router);
app.mount("#app");
