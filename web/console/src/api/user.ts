import { apiRequest } from "@/api/client";

export type MeResp = {
  user_id: string;
  phone: string;
};

export function me(): Promise<MeResp> {
  return apiRequest<MeResp>("/api/v1/me");
}
