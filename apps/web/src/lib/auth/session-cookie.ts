import {
  ACCESS_EXPIRES_AT_COOKIE,
  ACCESS_TOKEN_COOKIE,
  REFRESH_TOKEN_COOKIE,
  REFRESH_TOKEN_MAX_AGE_SECONDS,
} from "@/lib/auth/session-constants";

/**
 * BrowserSessionSnapshot defines the shape of browser session snapshot.
 */
export interface BrowserSessionSnapshot {
  accessToken: string | null;
  refreshToken: string | null;
  accessExpiresAt: number | null;
}

interface SessionWriteInput {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

interface AccessTokenWriteInput {
  accessToken: string;
  expiresIn: number;
}

const sessionChangedEventName = "bisakerja:session-changed";

/**
 * readBrowserSession reads browser session.
 */
export function readBrowserSession(): BrowserSessionSnapshot {
  return {
    accessToken: readCookie(ACCESS_TOKEN_COOKIE),
    refreshToken: readCookie(REFRESH_TOKEN_COOKIE),
    accessExpiresAt: parseOptionalNumber(readCookie(ACCESS_EXPIRES_AT_COOKIE)),
  };
}

/**
 * hasBrowserSession checks whether browser session.
 */
export function hasBrowserSession(): boolean {
  const session = readBrowserSession();
  return Boolean(session.accessToken || session.refreshToken);
}

/**
 * writeBrowserSession writes browser session.
 */
export function writeBrowserSession(input: SessionWriteInput): void {
  const expiresAt = Date.now() + Math.max(1, input.expiresIn) * 1000;

  writeCookie(ACCESS_TOKEN_COOKIE, input.accessToken, input.expiresIn);
  writeCookie(
    REFRESH_TOKEN_COOKIE,
    input.refreshToken,
    REFRESH_TOKEN_MAX_AGE_SECONDS,
  );
  writeCookie(
    ACCESS_EXPIRES_AT_COOKIE,
    String(expiresAt),
    REFRESH_TOKEN_MAX_AGE_SECONDS,
  );
  dispatchSessionChangedEvent();
}

/**
 * writeBrowserAccessToken writes browser access token.
 */
export function writeBrowserAccessToken(input: AccessTokenWriteInput): void {
  const expiresAt = Date.now() + Math.max(1, input.expiresIn) * 1000;
  writeCookie(ACCESS_TOKEN_COOKIE, input.accessToken, input.expiresIn);
  writeCookie(
    ACCESS_EXPIRES_AT_COOKIE,
    String(expiresAt),
    REFRESH_TOKEN_MAX_AGE_SECONDS,
  );
  dispatchSessionChangedEvent();
}

/**
 * clearBrowserSession handles clear browser session.
 */
export function clearBrowserSession(): void {
  clearCookie(ACCESS_TOKEN_COOKIE);
  clearCookie(REFRESH_TOKEN_COOKIE);
  clearCookie(ACCESS_EXPIRES_AT_COOKIE);
  dispatchSessionChangedEvent();
}

/**
 * subscribeSessionChanged handles subscribe session changed.
 */
export function subscribeSessionChanged(handler: () => void): () => void {
  if (typeof window === "undefined") {
    return () => {};
  }

  window.addEventListener(sessionChangedEventName, handler);
  return () => {
    window.removeEventListener(sessionChangedEventName, handler);
  };
}

function dispatchSessionChangedEvent(): void {
  if (typeof window === "undefined") {
    return;
  }
  window.dispatchEvent(new Event(sessionChangedEventName));
}

function parseOptionalNumber(value: string | null): number | null {
  if (!value) {
    return null;
  }

  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const encodedName = `${name}=`;
  const parts = document.cookie.split(";").map((value) => value.trim());
  for (const part of parts) {
    if (part.startsWith(encodedName)) {
      return decodeURIComponent(part.slice(encodedName.length));
    }
  }
  return null;
}

function writeCookie(name: string, value: string, maxAgeSeconds: number): void {
  if (typeof document === "undefined") {
    return;
  }

  const secure =
    typeof window !== "undefined" && window.location.protocol === "https:"
      ? "; Secure"
      : "";
  document.cookie = [
    `${name}=${encodeURIComponent(value)}`,
    "Path=/",
    `Max-Age=${Math.max(1, Math.floor(maxAgeSeconds))}`,
    "SameSite=Lax",
    secure,
  ].join("; ");
}

function clearCookie(name: string): void {
  if (typeof document === "undefined") {
    return;
  }
  document.cookie = `${name}=; Path=/; Max-Age=0; SameSite=Lax`;
}
