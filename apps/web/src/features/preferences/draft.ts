import type { UpdatePreferencesInput } from "@/services/preferences";

const draftStorageKey = "bisakerja:preferences-draft";
export const preferencesDraftCookie = "bisakerja_preferences_draft";

export interface PreferencesDraftPayload extends UpdatePreferencesInput {
  saved_at: string;
}

export function loadPreferencesDraft(): PreferencesDraftPayload | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(draftStorageKey);
  if (!raw) {
    return loadPreferencesDraftFromCookie(readCookie(preferencesDraftCookie));
  }

  try {
    return validateDraftPayload(JSON.parse(raw));
  } catch {
    return loadPreferencesDraftFromCookie(readCookie(preferencesDraftCookie));
  }
}

export function savePreferencesDraft(input: UpdatePreferencesInput): void {
  if (typeof window === "undefined") {
    return;
  }

  const payload: PreferencesDraftPayload = {
    ...input,
    saved_at: new Date().toISOString(),
  };
  const serialized = JSON.stringify(payload);
  window.localStorage.setItem(draftStorageKey, serialized);
  writeCookie(preferencesDraftCookie, serialized, 60 * 60 * 24);
}

export function clearPreferencesDraft(): void {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.removeItem(draftStorageKey);
  clearCookie(preferencesDraftCookie);
}

export function loadPreferencesDraftFromCookie(
  rawCookieValue: string | undefined | null,
): PreferencesDraftPayload | null {
  if (!rawCookieValue) {
    return null;
  }
  try {
    return validateDraftPayload(JSON.parse(decodeURIComponent(rawCookieValue)));
  } catch {
    return null;
  }
}

function validateDraftPayload(
  payload: unknown,
): PreferencesDraftPayload | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }

  const data = payload as Partial<PreferencesDraftPayload>;
  if (
    !Array.isArray(data.keywords) ||
    !Array.isArray(data.locations) ||
    !Array.isArray(data.job_types) ||
    typeof data.salary_min !== "number"
  ) {
    return null;
  }

  return {
    keywords: data.keywords,
    locations: data.locations,
    job_types: data.job_types,
    salary_min: data.salary_min,
    saved_at:
      typeof data.saved_at === "string"
        ? data.saved_at
        : new Date().toISOString(),
  };
}

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }
  const encodedName = `${name}=`;
  const parts = document.cookie.split(";").map((value) => value.trim());
  for (const part of parts) {
    if (part.startsWith(encodedName)) {
      return part.slice(encodedName.length);
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
