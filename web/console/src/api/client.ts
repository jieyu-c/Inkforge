import {
  readStoredAccessToken,
  writeAccessToken,
} from "@/lib/auth-token";
import { notifyAccessTokenChanged } from "@/lib/token-sync";

const BASE = import.meta.env.VITE_API_BASE || "";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function parseError(res: Response): Promise<ApiError> {
  try {
    const j = (await res.json()) as { code?: string; message?: string };
    return new ApiError(
      res.status,
      j.code ?? "HTTP_ERROR",
      j.message ?? res.statusText,
    );
  } catch {
    return new ApiError(res.status, "HTTP_ERROR", res.statusText);
  }
}

let refreshFlight: Promise<boolean> | null = null;

async function refreshSession(): Promise<boolean> {
  const res = await fetch(`${BASE}/api/v1/auth/refresh`, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  if (!res.ok) {
    writeAccessToken(null);
    notifyAccessTokenChanged();
    return false;
  }
  const data = (await res.json()) as {
    access_token: string;
    expires_in: number;
  };
  writeAccessToken(data.access_token);
  notifyAccessTokenChanged();
  return true;
}

async function sharedRefresh(): Promise<boolean> {
  if (refreshFlight) return refreshFlight;
  refreshFlight = refreshSession().finally(() => {
    refreshFlight = null;
  });
  return refreshFlight;
}

function skipRefreshForPath(path: string): boolean {
  return (
    path.includes("/auth/login") ||
    path.includes("/auth/register") ||
    path.includes("/auth/refresh")
  );
}

export type ApiRequestOptions = RequestInit & {
  /** Do not send Authorization (login, register, refresh). */
  skipAuthHeader?: boolean;
  /** Do not run 401 → refresh → retry (refresh endpoint itself). */
  skipRefreshRetry?: boolean;
};

export async function apiRequest<T>(
  path: string,
  init: ApiRequestOptions = {},
): Promise<T> {
  const {
    skipAuthHeader,
    skipRefreshRetry,
    headers: initHeaders,
    ...restInit
  } = init;
  const url = path.startsWith("http") ? path : `${BASE}${path}`;

  const headers = new Headers(initHeaders);
  if (!skipAuthHeader) {
    const t = readStoredAccessToken();
    if (t) headers.set("Authorization", `Bearer ${t}`);
  }
  if (!headers.has("Content-Type") && restInit.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  const exec = () =>
    fetch(url, {
      ...restInit,
      headers,
      credentials: "same-origin",
    });

  let res = await exec();

  if (
    res.status === 401 &&
    !skipRefreshRetry &&
    !skipRefreshForPath(path)
  ) {
    const ok = await sharedRefresh();
    if (ok) {
      const h2 = new Headers(initHeaders);
      if (!skipAuthHeader) {
        const t = readStoredAccessToken();
        if (t) h2.set("Authorization", `Bearer ${t}`);
      }
      if (!h2.has("Content-Type") && restInit.body !== undefined) {
        h2.set("Content-Type", "application/json");
      }
      res = await fetch(url, {
        ...restInit,
        headers: h2,
        credentials: "same-origin",
      });
    }
  }

  if (!res.ok) {
    throw await parseError(res);
  }

  const text = await res.text();
  if (!text) return undefined as T;
  return JSON.parse(text) as T;
}
