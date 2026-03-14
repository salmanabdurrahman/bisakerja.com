import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

export type UserRole = "user" | "admin";
export type SubscriptionState =
  | "free"
  | "pending_payment"
  | "premium_active"
  | "premium_expired";

export interface RegisterInput {
  email: string;
  password: string;
  name: string;
}

export interface RegisterResult {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  created_at: string;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface LoginResult {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_in: number;
}

export interface RefreshTokenInput {
  refresh_token: string;
}

export interface RefreshTokenResult {
  access_token: string;
  token_type: "Bearer";
  expires_in: number;
}

export interface AuthMe {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  is_premium: boolean;
  premium_expired_at?: string | null;
  subscription_state?: SubscriptionState;
}

export async function registerUser(
  input: RegisterInput,
  init?: RequestInit,
): Promise<APIResponse<RegisterResult>> {
  return fetchJSON<RegisterResult>(buildAPIURL("/auth/register"), {
    method: "POST",
    body: JSON.stringify(input),
    cache: "no-store",
    ...init,
  });
}

export async function loginUser(
  input: LoginInput,
  init?: RequestInit,
): Promise<APIResponse<LoginResult>> {
  return fetchJSON<LoginResult>(buildAPIURL("/auth/login"), {
    method: "POST",
    body: JSON.stringify(input),
    cache: "no-store",
    ...init,
  });
}

export async function refreshAuthToken(
  input: RefreshTokenInput,
  init?: RequestInit,
): Promise<APIResponse<RefreshTokenResult>> {
  return fetchJSON<RefreshTokenResult>(buildAPIURL("/auth/refresh"), {
    method: "POST",
    body: JSON.stringify(input),
    cache: "no-store",
    ...init,
  });
}

export async function getMe(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<AuthMe>> {
  return fetchJSON<AuthMe>(buildAPIURL("/auth/me"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}
