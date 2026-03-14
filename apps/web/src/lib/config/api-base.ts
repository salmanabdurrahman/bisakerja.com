const defaultPublicAPIBasePath = "/api/v1";
const defaultAPIOrigin = "http://localhost:8080";

function trimTrailingSlash(value: string): string {
  return value.length > 1 ? value.replace(/\/+$/, "") : value;
}

function normalizePath(path: string): string {
  const withLeadingSlash = path.startsWith("/") ? path : `/${path}`;
  const trimmed = trimTrailingSlash(withLeadingSlash);
  return trimmed.length > 0 ? trimmed : "/";
}

function parseURL(value: string): URL | null {
  try {
    return new URL(value);
  } catch {
    return null;
  }
}

/**
 * normalizeAPIBasePath normalizes api base path.
 */
export function normalizeAPIBasePath(value: string | undefined): string {
  const raw = value?.trim();
  if (!raw) {
    return defaultPublicAPIBasePath;
  }

  if (/^https?:\/\//i.test(raw)) {
    const parsed = parseURL(raw);
    if (!parsed) {
      return defaultPublicAPIBasePath;
    }

    const normalizedPath = normalizePath(parsed.pathname || "/");
    return normalizedPath === "/" ? defaultPublicAPIBasePath : normalizedPath;
  }

  const normalizedPath = normalizePath(raw);
  return normalizedPath === "/" ? defaultPublicAPIBasePath : normalizedPath;
}

/**
 * resolveAPIOrigin resolves api origin.
 */
export function resolveAPIOrigin(value: string | undefined): string {
  const raw = value?.trim();
  if (!raw) {
    return defaultAPIOrigin;
  }

  const parsed = /^https?:\/\//i.test(raw)
    ? parseURL(raw)
    : parseURL(`http://${raw}`);
  return parsed?.origin ?? defaultAPIOrigin;
}

/**
 * buildServerAPIBaseURL builds server api base url.
 */
export function buildServerAPIBaseURL(
  apiOrigin: string | undefined,
  publicBasePath: string | undefined,
): string {
  return `${resolveAPIOrigin(apiOrigin)}${normalizeAPIBasePath(publicBasePath)}`;
}
