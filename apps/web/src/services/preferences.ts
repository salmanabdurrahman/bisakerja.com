import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

export type PreferredJobType =
  | "fulltime"
  | "parttime"
  | "contract"
  | "internship";

export interface UserPreferences {
  user_id: string;
  keywords: string[];
  locations: string[];
  job_types: PreferredJobType[];
  salary_min: number;
  updated_at?: string | null;
}

export interface UpdatePreferencesInput {
  keywords: string[];
  locations: string[];
  job_types: PreferredJobType[];
  salary_min: number;
}

export async function getPreferences(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<UserPreferences>> {
  return fetchJSON<UserPreferences>(buildAPIURL("/preferences"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

export async function updatePreferences(
  accessToken: string,
  input: UpdatePreferencesInput,
  init?: RequestInit,
): Promise<APIResponse<UserPreferences>> {
  return fetchJSON<UserPreferences>(buildAPIURL("/preferences"), {
    method: "PUT",
    body: JSON.stringify(input),
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}
