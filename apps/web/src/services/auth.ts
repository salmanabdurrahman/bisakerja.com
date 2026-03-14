import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

/**
 * UserRole defines the shape of user role.
 */
export type UserRole = "user" | "admin";
/**
 * SubscriptionState defines the shape of subscription state.
 */
export type SubscriptionState =
  | "free"
  | "pending_payment"
  | "premium_active"
  | "premium_expired";

/**
 * RegisterInput defines the shape of register input.
 */
export interface RegisterInput {
  email: string;
  password: string;
  name: string;
}

/**
 * RegisterResult defines the shape of register result.
 */
export interface RegisterResult {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  created_at: string;
}

/**
 * LoginInput defines the shape of login input.
 */
export interface LoginInput {
  email: string;
  password: string;
}

/**
 * LoginResult defines the shape of login result.
 */
export interface LoginResult {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_in: number;
}

/**
 * RefreshTokenInput defines the shape of refresh token input.
 */
export interface RefreshTokenInput {
  refresh_token: string;
}

/**
 * RefreshTokenResult defines the shape of refresh token result.
 */
export interface RefreshTokenResult {
  access_token: string;
  token_type: "Bearer";
  expires_in: number;
}

/**
 * AuthMe defines the shape of auth me.
 */
export interface AuthMe {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  is_premium: boolean;
  premium_expired_at?: string | null;
  subscription_state?: SubscriptionState;
}

/**
 * registerUser handles register user.
 */
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

/**
 * loginUser handles login user.
 */
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

/**
 * refreshAuthToken handles refresh auth token.
 */
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

/**
 * getMe returns me.
 */
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
