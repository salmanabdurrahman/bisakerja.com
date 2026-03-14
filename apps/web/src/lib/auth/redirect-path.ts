const defaultRedirectPath = "/account";

/**
 * normalizeRedirectPath normalizes redirect path.
 */
export function normalizeRedirectPath(
  rawValue: string | null | undefined,
): string {
  if (!rawValue) {
    return defaultRedirectPath;
  }

  const value = rawValue.trim();
  if (!value.startsWith("/")) {
    return defaultRedirectPath;
  }
  if (value.startsWith("//")) {
    return defaultRedirectPath;
  }
  if (value.startsWith("/auth/login") || value.startsWith("/auth/register")) {
    return defaultRedirectPath;
  }

  return value;
}

/**
 * buildLoginHref builds login href.
 */
export function buildLoginHref(redirectPath: string): string {
  const normalized = normalizeRedirectPath(redirectPath);
  return `/auth/login?redirect=${encodeURIComponent(normalized)}`;
}
