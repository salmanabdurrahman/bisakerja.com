import type { APIPagination, APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";
import type { NotificationAlertMode } from "@/services/preferences";

/**
 * JobSource defines the shape of job source.
 */
export type JobSource = "glints" | "kalibrr" | "jobstreet";

/**
 * SavedSearch defines the shape of saved search.
 */
export interface SavedSearch {
  id: string;
  query: string;
  location?: string | null;
  source?: JobSource | null;
  salary_min?: number | null;
  frequency: NotificationAlertMode;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * CreateSavedSearchInput defines the shape of create saved search input.
 */
export interface CreateSavedSearchInput {
  query: string;
  location?: string;
  source?: JobSource;
  salary_min?: number;
  frequency?: NotificationAlertMode;
  is_active?: boolean;
}

/**
 * WatchlistCompany defines the shape of watchlist company.
 */
export interface WatchlistCompany {
  company_slug: string;
  [key: string]: unknown;
}

/**
 * NotificationRecord defines the shape of notification record.
 */
export interface NotificationRecord {
  id: string;
  job_id: string;
  channel: string;
  status: string;
  error_message?: string | null;
  sent_at: string;
  read_at?: string | null;
  created_at: string;
}

/**
 * NotificationsQuery defines the shape of notifications query.
 */
export interface NotificationsQuery {
  page?: number;
  limit?: number;
  unread_only?: boolean;
}

/**
 * NotificationListResult defines the shape of notification list result.
 */
export interface NotificationListResult {
  items: NotificationRecord[];
  pagination: APIPagination | null;
}

/**
 * createSavedSearch creates saved search.
 */
export async function createSavedSearch(
  accessToken: string,
  input: CreateSavedSearchInput,
  init?: RequestInit,
): Promise<APIResponse<SavedSearch>> {
  return fetchJSON<SavedSearch>(buildAPIURL("/saved-searches"), {
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

/**
 * listSavedSearches returns a list of saved searches.
 */
export async function listSavedSearches(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<SavedSearch[]>> {
  return fetchJSON<SavedSearch[]>(buildAPIURL("/saved-searches"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * deleteSavedSearch deletes saved search.
 */
export async function deleteSavedSearch(
  accessToken: string,
  id: string,
  init?: RequestInit,
): Promise<APIResponse<{ id: string }>> {
  return fetchJSON<{ id: string }>(buildAPIURL(`/saved-searches/${id}`), {
    method: "DELETE",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * addCompanyToWatchlist handles add company to watchlist.
 */
export async function addCompanyToWatchlist(
  accessToken: string,
  companySlug: string,
  init?: RequestInit,
): Promise<APIResponse<WatchlistCompany>> {
  return fetchJSON<WatchlistCompany>(buildAPIURL("/watchlist/companies"), {
    method: "POST",
    body: JSON.stringify({ company_slug: companySlug }),
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * listWatchlistCompanies returns a list of watchlist companies.
 */
export async function listWatchlistCompanies(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<WatchlistCompany[]>> {
  return fetchJSON<WatchlistCompany[]>(buildAPIURL("/watchlist/companies"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * removeCompanyFromWatchlist removes company from watchlist.
 */
export async function removeCompanyFromWatchlist(
  accessToken: string,
  companySlug: string,
  init?: RequestInit,
): Promise<APIResponse<{ company_slug: string }>> {
  return fetchJSON<{ company_slug: string }>(
    buildAPIURL(`/watchlist/companies/${companySlug}`),
    {
      method: "DELETE",
      cache: "no-store",
      ...init,
      headers: {
        Authorization: `Bearer ${accessToken}`,
        ...(init?.headers ?? {}),
      },
    },
  );
}

/**
 * listNotifications returns a list of notifications.
 */
export async function listNotifications(
  accessToken: string,
  query: NotificationsQuery = {},
  init?: RequestInit,
): Promise<APIResponse<NotificationRecord[]>> {
  const params = new URLSearchParams();
  if (query.page !== undefined) {
    params.set("page", String(query.page));
  }
  if (query.limit !== undefined) {
    params.set("limit", String(query.limit));
  }
  if (query.unread_only !== undefined) {
    params.set("unread_only", String(query.unread_only));
  }

  const suffix = params.toString();
  const endpoint = suffix ? `/notifications?${suffix}` : "/notifications";

  return fetchJSON<NotificationRecord[]>(buildAPIURL(endpoint), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * markNotificationAsRead marks notification as read.
 */
export async function markNotificationAsRead(
  accessToken: string,
  notificationID: string,
  init?: RequestInit,
): Promise<APIResponse<NotificationRecord>> {
  return fetchJSON<NotificationRecord>(
    buildAPIURL(`/notifications/${notificationID}/read`),
    {
      method: "PATCH",
      cache: "no-store",
      ...init,
      headers: {
        Authorization: `Bearer ${accessToken}`,
        ...(init?.headers ?? {}),
      },
    },
  );
}
