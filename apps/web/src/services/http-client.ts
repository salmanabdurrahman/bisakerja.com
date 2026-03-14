import { getAPIBaseURL } from "@/lib/config/env";

/**
 * buildAPIURL builds apiurl.
 */
export function buildAPIURL(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${getAPIBaseURL()}${normalizedPath}`;
}
