import { apiRequest } from "@/api/client";

export type RegisterReq = {
  phone: string;
  password: string;
};

export type RegisterResp = {
  message: string;
};

export type LoginReq = {
  phone: string;
  password: string;
};

export type LoginResp = {
  access_token: string;
  expires_in: number;
};

export function register(body: RegisterReq): Promise<RegisterResp> {
  return apiRequest<RegisterResp>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(body),
    skipAuthHeader: true,
    skipRefreshRetry: true,
  });
}

export function login(body: LoginReq): Promise<LoginResp> {
  return apiRequest<LoginResp>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(body),
    skipAuthHeader: true,
    skipRefreshRetry: true,
  });
}

/** Uses HttpOnly refresh cookie; updates stored access token from response. */
export function refresh(): Promise<LoginResp> {
  return apiRequest<LoginResp>("/api/v1/auth/refresh", {
    method: "POST",
    body: "{}",
    skipAuthHeader: true,
    skipRefreshRetry: true,
  });
}

export function logout(): Promise<Record<string, never>> {
  return apiRequest<Record<string, never>>("/api/v1/auth/logout", {
    method: "POST",
    body: "{}",
    skipAuthHeader: true,
    skipRefreshRetry: true,
  });
}
