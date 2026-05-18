import * as authApi from "@/api/auth";
import { readStoredAccessToken, writeAccessToken } from "@/lib/auth-token";
import { defineStore } from "pinia";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    accessToken: readStoredAccessToken() as string | null,
  }),
  getters: {
    isAuthenticated: (s) => !!s.accessToken,
  },
  actions: {
    setAccess(token: string | null) {
      writeAccessToken(token);
      this.accessToken = token;
    },
    syncFromStorage() {
      this.accessToken = readStoredAccessToken();
    },
    async login(phone: string, password: string) {
      const data = await authApi.login({ phone, password });
      this.setAccess(data.access_token);
      return data;
    },
    async register(phone: string, password: string) {
      await authApi.register({ phone, password });
    },
    async refreshWithCookie() {
      const data = await authApi.refresh();
      this.setAccess(data.access_token);
      return data;
    },
    async logout() {
      try {
        await authApi.logout();
      } finally {
        this.setAccess(null);
      }
    },
    async ensureSession() {
      if (this.accessToken) return true;
      try {
        await this.refreshWithCookie();
        return true;
      } catch {
        return false;
      }
    },
  },
});
