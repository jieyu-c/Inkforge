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

function tryRestEnvelope(
  parsed: unknown,
): { code: number; msg: string; data?: unknown } | null {
  if (parsed === null || typeof parsed !== "object") return null;
  const o = parsed as Record<string, unknown>;
  if (typeof o.code !== "number" || typeof o.msg !== "string") return null;
  return { code: o.code, msg: o.msg, data: o.data };
}

/** Server uses jieyuc-common RestResponse: HTTP often 200 with code/msg/data. */
function unwrapDataOrThrow<T>(res: Response, parsed: unknown): T {
  const env = tryRestEnvelope(parsed);
  if (!env) return parsed as T;
  if (env.code !== 0) {
    throw new ApiError(res.status, String(env.code), env.msg);
  }
  return env.data as T;
}

async function parseError(res: Response): Promise<ApiError> {
  try {
    const j = (await res.json()) as {
      code?: string | number;
      message?: string;
      msg?: string;
    };
    const message = j.message ?? j.msg ?? res.statusText;
    const codeRaw = j.code;
    const code =
      typeof codeRaw === "number"
        ? String(codeRaw)
        : (codeRaw ?? "HTTP_ERROR");
    return new ApiError(res.status, code, message);
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
  const text = await res.text();
  if (!text) {
    writeAccessToken(null);
    notifyAccessTokenChanged();
    return false;
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    writeAccessToken(null);
    notifyAccessTokenChanged();
    return false;
  }
  try {
    const data = unwrapDataOrThrow<{
      access_token: string;
      expires_in: number;
    }>(res, parsed);
    writeAccessToken(data.access_token);
  } catch {
    writeAccessToken(null);
    notifyAccessTokenChanged();
    return false;
  }
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
  const parsed: unknown = JSON.parse(text);
  return unwrapDataOrThrow<T>(res, parsed);
}
