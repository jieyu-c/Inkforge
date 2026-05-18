import { useAuthStore } from "@/stores/auth";
import HomeView from "@/views/HomeView.vue";
import LoginView from "@/views/LoginView.vue";
import RegisterView from "@/views/RegisterView.vue";
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
    meta: { requiresAuth: true },
  },
];

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

function wantsAuth(
  m: RouteLocationNormalized["meta"],
): m is { requiresAuth: true } {
  return m.requiresAuth === true;
}

function wantsGuest(
  m: RouteLocationNormalized["meta"],
): m is { guestOnly: true } {
  return m.guestOnly === true;
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
      return { name: "home" };
    }
  }

  if (wantsAuth(to.meta)) {
    const ok = await auth.ensureSession();
    if (!ok) {
      return { name: "login", query: { redirect: to.fullPath } };
    }
  }

  return true;
});
