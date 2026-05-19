import {
  archiveNamespace,
  createNamespace,
  listNamespaces,
  restoreNamespace,
  type CreateNamespaceReq,
  type NamespaceDetail,
} from "@/api/namespaces";
import { ApiError } from "@/api/client";
import { defineStore } from "pinia";

const STORAGE_KEY = "inkforge.selectedNsSlug";

function readNsFromStorage(): string {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    return v != null && v.trim() !== "" ? v.trim() : "";
  } catch {
    return "";
  }
}

function writeNsToStorage(slug: string) {
  try {
    if (slug === "") {
      localStorage.removeItem(STORAGE_KEY);
    } else {
      localStorage.setItem(STORAGE_KEY, slug);
    }
  } catch {
    /* ignore quota / private mode */
  }
}

/**
 * Workspace focus: authenticated personal namespaces plus client-selected slug.
 */
export const useWorkspaceContextStore = defineStore("workspaceContext", {
  state: () => ({
    selectedNsSlug: readNsFromStorage(),
    namespaces: [] as NamespaceDetail[],
    namespacesLoading: false,
    namespacesError: null as string | null,
  }),
  getters: {
    hasSelectedNs: (s) => s.selectedNsSlug.length > 0,
    selectedNamespace: (s) =>
      s.namespaces.find((n) => n.ns_slug === s.selectedNsSlug) ?? null,
  },
  actions: {
    setSelectedNsSlug(slug: string) {
      const next = slug.trim();
      this.selectedNsSlug = next;
      writeNsToStorage(next);
    },
    /** After NS list loads, drop invalid persisted selection or default to first. */
    syncSelectionAgainstNsList(slugs: string[]) {
      const set = new Set(slugs);
      if (this.selectedNsSlug && !set.has(this.selectedNsSlug)) {
        const first = slugs[0] ?? "";
        this.setSelectedNsSlug(first);
        return;
      }
      if (!this.selectedNsSlug && slugs.length === 1) {
        this.setSelectedNsSlug(slugs[0]!);
      }
    },
    async reloadNamespaces(options?: { selectFirst?: boolean }) {
      this.namespacesLoading = true;
      this.namespacesError = null;
      try {
        const res = await listNamespaces();
        this.namespaces = res.namespaces ?? [];
        const slugs = this.namespaces.map((n) => n.ns_slug);
        if (options?.selectFirst === true && !this.selectedNsSlug && slugs.length > 0) {
          this.setSelectedNsSlug(slugs[0]!);
        } else {
          this.syncSelectionAgainstNsList(slugs);
        }
      } catch (e) {
        this.namespaces = [];
        let msg =
          typeof e === "object" &&
          e !== null &&
          "message" in e &&
          typeof (e as { message: unknown }).message === "string"
            ? (e as { message: string }).message
            : "Failed to load namespaces";
        if (e instanceof ApiError && e.code) msg = `${e.code}: ${e.message}`;
        this.namespacesError = msg;
        throw e;
      } finally {
        this.namespacesLoading = false;
      }
    },
    async createAndSelect(body: CreateNamespaceReq) {
      const created = await createNamespace(body);
      await this.reloadNamespaces();
      if (created?.ns_slug) this.setSelectedNsSlug(created.ns_slug);
      return created;
    },
    async archiveSelected(slug?: string) {
      const ns = slug?.trim() || this.selectedNsSlug;
      if (!ns) return;
      await archiveNamespace(ns);
      await this.reloadNamespaces();
      if (this.selectedNsSlug === ns) {
        const next = this.namespaces.find((x) => x.status === "active" && x.ns_slug !== ns);
        this.setSelectedNsSlug(next?.ns_slug ?? this.namespaces[0]?.ns_slug ?? "");
      }
    },
    async restoreSelected(slug?: string) {
      const ns = slug?.trim() || this.selectedNsSlug;
      if (!ns) return;
      await restoreNamespace(ns);
      await this.reloadNamespaces();
    },
  },
});
