import ConsoleLayout from "@/layouts/ConsoleLayout.vue";
import { useAuthStore } from "@/stores/auth";
import ComingSoonView from "@/views/console/ComingSoonView.vue";
import SettingsView from "@/views/console/SettingsView.vue";
import HomeView from "@/views/HomeView.vue";
import LoginView from "@/views/LoginView.vue";
import RegisterView from "@/views/RegisterView.vue";
import WorkspaceView from "@/views/WorkspaceView.vue";
import {
  createRouter,
  createWebHistory,
  type RouteLocationNormalized,
} from "vue-router";

const routes = [
  { path: "/", redirect: "/home" },
  {
    path: "/login",
    name: "login",
    component: LoginView,
    meta: { guestOnly: true },
  },
  {
    path: "/register",
    name: "register",
    component: RegisterView,
    meta: { guestOnly: true },
  },
  {
    path: "/home",
    name: "home",
    component: HomeView,
  },
  {
    path: "/workspace",
    component: ConsoleLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        name: "workspace",
        component: WorkspaceView,
        meta: {
          titleI18nKey: "workspace.pageTitle",
        },
      },
      {
        path: "prompts",
        name: "console-prompts",
        component: ComingSoonView,
        meta: {
          titleI18nKey: "workspace.comingSoonPromptsTitle",
        },
      },
      {
        path: "settings",
        name: "console-settings",
        component: SettingsView,
        meta: {
          titleI18nKey: "workspace.settingsPageTitle",
        },
      },
    ],
  },
];

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

function wantsGuest(
  m: RouteLocationNormalized["meta"],
): m is { guestOnly: true } {
  return m.guestOnly === true;
}

function requiresAuthMatched(to: RouteLocationNormalized): boolean {
  return to.matched.some((r) => r.meta.requiresAuth === true);
}

router.beforeEach(async (to) => {
  const auth = useAuthStore();

  if (wantsGuest(to.meta)) {
    if (!auth.accessToken) {
      try {
        await auth.refreshWithCookie();
      } catch {
        /* no cookie session */
      }
    }
    if (auth.accessToken) {
      return { name: "workspace" };
    }
  }

  if (requiresAuthMatched(to)) {
    const ok = await auth.ensureSession();
    if (!ok) {
      return { name: "login", query: { redirect: to.fullPath } };
    }
  }

  return true;
});
