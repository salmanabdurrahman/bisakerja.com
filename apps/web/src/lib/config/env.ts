import {
  buildServerAPIBaseURL,
  normalizeAPIBasePath,
} from "@/lib/config/api-base";

export function getAPIBaseURL(): string {
  if (typeof window !== "undefined") {
    return normalizeAPIBasePath(process.env.NEXT_PUBLIC_API_BASE_URL);
  }

  return buildServerAPIBaseURL(
    process.env.API_ORIGIN,
    process.env.NEXT_PUBLIC_API_BASE_URL,
  );
}
