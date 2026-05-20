import { apiRequest } from "@/api/client";

export type PromptSummary = {
  prompt_key: string;
  title?: string;
  tags?: string[];
  owner_user_id?: string;
  updated_at: string;
};

export type PromptListResp = {
  items: PromptSummary[];
  total: number;
};

export type PromptDetail = {
  prompt_key: string;
  title?: string;
  tags?: string[];
  owner_user_id?: string;
  draft_body: string;
  draft_schema?: string;
  updated_at: string;
};

export type CreatePromptReq = {
  title: string;
  tags?: string[];
  owner_user_id?: string;
};

export type PatchPromptReq = {
  title?: string;
  tags?: string[];
  owner_user_id?: string;
};

export type PutDraftReq = {
  body: string;
  schema?: string;
};

export type DraftResp = {
  body: string;
  schema?: string;
  warnings?: string[];
  updated_at: string;
};

export type CreateVersionReq = {
  change_note?: string;
  /** Omit for auto (1.0.0, patch+1, …); set for semver e.g. 1.2.3 */
  version?: string;
};

export type PromptVersionItem = {
  id: string;
  version: string;
  change_note?: string;
  created_by_user_id: string;
  created_at: string;
};

export type PromptVersionListResp = {
  items: PromptVersionItem[];
  total: number;
};

export type VersionDiffResp = {
  body_diff: string;
  schema_diff: string;
};

export type ChannelPointerResp = {
  channel: string;
  version_id: string;
  version: string;
  updated_at: string;
};

export type PatchChannelPointerReq = {
  version_id: string;
};

function enc(s: string) {
  return encodeURIComponent(s);
}

export function listPrompts(
  nsSlug: string,
  params?: { page?: number; page_size?: number; q?: string },
) {
  const q = new URLSearchParams();
  if (params?.page != null) q.set("page", String(params.page));
  if (params?.page_size != null) q.set("page_size", String(params.page_size));
  if (params?.q) q.set("q", params.q);
  const qs = q.toString();
  const path = `/api/v1/me/namespaces/${enc(nsSlug)}/prompts${qs ? `?${qs}` : ""}`;
  return apiRequest<PromptListResp>(path, { method: "GET" });
}

export function createPrompt(nsSlug: string, body: CreatePromptReq) {
  return apiRequest<PromptDetail>(`/api/v1/me/namespaces/${enc(nsSlug)}/prompts`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function getPrompt(nsSlug: string, promptKey: string) {
  return apiRequest<PromptDetail>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}`,
    { method: "GET" },
  );
}

export function patchPrompt(nsSlug: string, promptKey: string, body: PatchPromptReq) {
  return apiRequest<PromptDetail>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}`,
    { method: "PATCH", body: JSON.stringify(body) },
  );
}

export function deletePrompt(nsSlug: string, promptKey: string) {
  return apiRequest<unknown>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}`,
    { method: "DELETE", body: "{}" },
  );
}

export function getDraft(nsSlug: string, promptKey: string) {
  return apiRequest<DraftResp>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/draft`,
    { method: "GET" },
  );
}

export function putDraft(nsSlug: string, promptKey: string, body: PutDraftReq) {
  return apiRequest<DraftResp>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/draft`,
    { method: "PUT", body: JSON.stringify(body) },
  );
}

export function createVersion(nsSlug: string, promptKey: string, body: CreateVersionReq) {
  return apiRequest<PromptVersionItem>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/versions`,
    { method: "POST", body: JSON.stringify(body) },
  );
}

export function listVersions(
  nsSlug: string,
  promptKey: string,
  params?: { page?: number; page_size?: number; q?: string },
) {
  const q = new URLSearchParams();
  if (params?.page != null) q.set("page", String(params.page));
  if (params?.page_size != null) q.set("page_size", String(params.page_size));
  if (params?.q?.trim()) q.set("q", params.q.trim());
  const qs = q.toString();
  return apiRequest<PromptVersionListResp>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/versions${qs ? `?${qs}` : ""}`,
    { method: "GET" },
  );
}

export function diffVersions(nsSlug: string, promptKey: string, versionA: string, versionB: string) {
  const q = new URLSearchParams();
  q.set("version_a", versionA.trim());
  q.set("version_b", versionB.trim());
  return apiRequest<VersionDiffResp>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/versions/diff?${q}`,
    { method: "GET" },
  );
}

export function getChannelPointer(nsSlug: string, promptKey: string, channel: string) {
  return apiRequest<ChannelPointerResp>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/channels/${enc(channel)}`,
    { method: "GET" },
  );
}

export function patchChannelPointer(
  nsSlug: string,
  promptKey: string,
  channel: string,
  body: PatchChannelPointerReq,
) {
  return apiRequest<ChannelPointerResp>(
    `/api/v1/me/namespaces/${enc(nsSlug)}/prompts/${enc(promptKey)}/channels/${enc(channel)}`,
    { method: "PATCH", body: JSON.stringify(body) },
  );
}
