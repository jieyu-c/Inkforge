import { createApp } from "vue";
import App from "./App.vue";
import { registerAccessTokenSync } from "@/lib/token-sync";
import { router } from "@/router";
import { useAuthStore } from "@/stores/auth";
import { createPinia } from "pinia";
import { i18n, syncDocumentLang } from "./i18n";
import type { AppLocale } from "./i18n/constants";
import { initialTheme, syncDocumentTheme } from "./theme";

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
registerAccessTokenSync(() => {
  useAuthStore().syncFromStorage();
});
app.use(i18n);
syncDocumentLang((i18n.global.locale as { value: AppLocale }).value);
syncDocumentTheme(initialTheme());
app.use(router);
app.mount("#app");
