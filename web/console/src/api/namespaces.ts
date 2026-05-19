import { apiRequest } from "@/api/client";

export type NamespaceDetail = {
  ns_slug: string;
  display_name: string;
  description?: string;
  status: string;
  default_channel_slug?: string;
  tags?: string[];
  quota_prompts_max: number;
  prompt_count: number;
  archived_at?: string;
};

export type NamespaceListResp = {
  namespaces: NamespaceDetail[];
};

export type CreateNamespaceReq = {
  ns_slug: string;
  display_name: string;
  description?: string;
  default_channel_slug?: string;
  tags?: string[];
  quota_prompts_max?: number;
};

export function listNamespaces() {
  return apiRequest<NamespaceListResp>("/api/v1/me/namespaces", { method: "GET" });
}

export function createNamespace(body: CreateNamespaceReq) {
  return apiRequest<NamespaceDetail>("/api/v1/me/namespaces", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function archiveNamespace(nsSlug: string) {
  return apiRequest<NamespaceDetail>(`/api/v1/me/namespaces/${encodeURIComponent(nsSlug)}/archive`, {
    method: "POST",
    body: "{}",
  });
}

export function restoreNamespace(nsSlug: string) {
  return apiRequest<NamespaceDetail>(`/api/v1/me/namespaces/${encodeURIComponent(nsSlug)}/restore`, {
    method: "POST",
    body: "{}",
  });
}
