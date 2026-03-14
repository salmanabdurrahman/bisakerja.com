import { cookies } from "next/headers";

import {
  ACCESS_EXPIRES_AT_COOKIE,
  ACCESS_REFRESH_WINDOW_SECONDS,
  ACCESS_TOKEN_COOKIE,
  REFRESH_TOKEN_COOKIE,
} from "@/lib/auth/session-constants";
import { refreshAuthToken } from "@/services/auth";

/**
 * resolveServerAccessToken resolves server access token.
 */
export async function resolveServerAccessToken(): Promise<string | null> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(ACCESS_TOKEN_COOKIE)?.value ?? null;
  const refreshToken = cookieStore.get(REFRESH_TOKEN_COOKIE)?.value ?? null;
  const accessExpiresAtRaw = cookieStore.get(ACCESS_EXPIRES_AT_COOKIE)?.value;

  if (!accessToken && !refreshToken) {
    return null;
  }

  if (!accessToken && refreshToken) {
    return refreshFromServer(refreshToken);
  }

  if (accessToken && shouldRefreshSoon(accessExpiresAtRaw) && refreshToken) {
    const refreshed = await refreshFromServer(refreshToken);
    return refreshed ?? accessToken;
  }

  return accessToken;
}

async function refreshFromServer(refreshToken: string): Promise<string | null> {
  try {
    const response = await refreshAuthToken({
      refresh_token: refreshToken,
    });
    return response.data.access_token;
  } catch {
    return null;
  }
}

function shouldRefreshSoon(rawExpiresAt: string | undefined): boolean {
  if (!rawExpiresAt) {
    return false;
  }
  const parsed = Number(rawExpiresAt);
  if (!Number.isFinite(parsed)) {
    return false;
  }
  return Date.now() >= parsed - ACCESS_REFRESH_WINDOW_SECONDS * 1000;
}
