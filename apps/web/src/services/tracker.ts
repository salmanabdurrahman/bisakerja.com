import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

export type ApplicationStatus =
  | "applied"
  | "interview"
  | "offer"
  | "rejected"
  | "withdrawn";

export interface Bookmark {
  id: string;
  job_id: string;
  created_at: string;
}

export interface TrackedApplication {
  id: string;
  job_id: string;
  status: ApplicationStatus;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTrackedApplicationInput {
  job_id: string;
  notes?: string;
}

export interface UpdateApplicationStatusInput {
  status: ApplicationStatus;
}

export async function createBookmark(
  accessToken: string,
  jobID: string,
  init?: RequestInit,
): Promise<APIResponse<Bookmark>> {
  return fetchJSON<Bookmark>(buildAPIURL("/bookmarks"), {
    method: "POST",
    body: JSON.stringify({ job_id: jobID }),
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

export async function deleteBookmark(
  accessToken: string,
  jobID: string,
  init?: RequestInit,
): Promise<APIResponse<{ job_id: string }>> {
  return fetchJSON<{ job_id: string }>(buildAPIURL(`/bookmarks/${jobID}`), {
    method: "DELETE",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

export async function listBookmarks(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<Bookmark[]>> {
  return fetchJSON<Bookmark[]>(buildAPIURL("/bookmarks"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

export async function createTrackedApplication(
  accessToken: string,
  input: CreateTrackedApplicationInput,
  init?: RequestInit,
): Promise<APIResponse<TrackedApplication>> {
  return fetchJSON<TrackedApplication>(buildAPIURL("/applications"), {
    method: "POST",
    body: JSON.stringify(input),
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

export async function updateApplicationStatus(
  accessToken: string,
  id: string,
  input: UpdateApplicationStatusInput,
  init?: RequestInit,
): Promise<APIResponse<TrackedApplication>> {
  return fetchJSON<TrackedApplication>(
    buildAPIURL(`/applications/${id}/status`),
    {
      method: "PATCH",
      body: JSON.stringify(input),
      cache: "no-store",
      ...init,
      headers: {
        Authorization: `Bearer ${accessToken}`,
        ...(init?.headers ?? {}),
      },
    },
  );
}

export async function deleteTrackedApplication(
  accessToken: string,
  id: string,
  init?: RequestInit,
): Promise<APIResponse<{ id: string }>> {
  return fetchJSON<{ id: string }>(buildAPIURL(`/applications/${id}`), {
    method: "DELETE",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

export async function listTrackedApplications(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<TrackedApplication[]>> {
  return fetchJSON<TrackedApplication[]>(buildAPIURL("/applications"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}
