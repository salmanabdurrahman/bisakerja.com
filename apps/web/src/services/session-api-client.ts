import { ACCESS_REFRESH_WINDOW_SECONDS } from "@/lib/auth/session-constants";
import {
  clearBrowserSession,
  readBrowserSession,
  writeBrowserAccessToken,
} from "@/lib/auth/session-cookie";
import type { APIResponse } from "@/lib/types/api";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, refreshAuthToken, type AuthMe } from "@/services/auth";
import {
  createCheckoutSession,
  getBillingStatus,
  getBillingTransactions,
  type BillingStatus,
  type BillingTransaction,
  type BillingTransactionsQuery,
  type CheckoutSession,
  type CreateCheckoutSessionInput,
} from "@/services/billing";
import {
  getPreferences,
  updateNotificationPreferences,
  updatePreferences,
  type NotificationPreferences,
  type UpdateNotificationPreferencesInput,
  type UpdatePreferencesInput,
  type UserPreferences,
} from "@/services/preferences";
import {
  createSavedSearch,
  deleteSavedSearch,
  listNotifications,
  listSavedSearches,
  markNotificationAsRead,
  type CreateSavedSearchInput,
  type NotificationRecord,
  type NotificationsQuery,
  type SavedSearch,
} from "@/services/growth";
import {
  generateAICoverLetterDraft,
  generateAIJobFitSummary,
  generateAISearchAssistant,
  getAIUsage,
  type AICoverLetterDraftResult,
  type AIFeature,
  type AIJobFitSummaryResult,
  type AISearchAssistantResult,
  type AIUsage,
  type GenerateAICoverLetterDraftInput,
  type GenerateAIJobFitSummaryInput,
  type GenerateAISearchAssistantInput,
} from "@/services/ai";

interface SessionClientDependencies {
  getSession: typeof readBrowserSession;
  updateAccessToken: typeof writeBrowserAccessToken;
  clearSession: typeof clearBrowserSession;
  refresh: typeof refreshAuthToken;
}

const defaultDependencies: SessionClientDependencies = {
  getSession: readBrowserSession,
  updateAccessToken: writeBrowserAccessToken,
  clearSession: clearBrowserSession,
  refresh: refreshAuthToken,
};

/**
 * SessionAPIClient defines the shape of session api client.
 */
export interface SessionAPIClient {
  getMe: () => Promise<APIResponse<AuthMe>>;
  getBillingStatus: () => Promise<APIResponse<BillingStatus>>;
  createCheckoutSession: (
    input: CreateCheckoutSessionInput,
  ) => Promise<APIResponse<CheckoutSession>>;
  getBillingTransactions: (
    query?: BillingTransactionsQuery,
  ) => Promise<APIResponse<BillingTransaction[]>>;
  getPreferences: () => Promise<APIResponse<UserPreferences>>;
  updatePreferences: (
    input: UpdatePreferencesInput,
  ) => Promise<APIResponse<UserPreferences>>;
  listSavedSearches: () => Promise<APIResponse<SavedSearch[]>>;
  createSavedSearch: (
    input: CreateSavedSearchInput,
  ) => Promise<APIResponse<SavedSearch>>;
  deleteSavedSearch: (id: string) => Promise<APIResponse<{ id: string }>>;
  listNotifications: (
    query?: NotificationsQuery,
  ) => Promise<APIResponse<NotificationRecord[]>>;
  markNotificationAsRead: (
    notificationID: string,
  ) => Promise<APIResponse<NotificationRecord>>;
  updateNotificationPreferences: (
    input: UpdateNotificationPreferencesInput,
  ) => Promise<APIResponse<NotificationPreferences>>;
  getAIUsage: (feature: AIFeature) => Promise<APIResponse<AIUsage>>;
  generateAISearchAssistant: (
    input: GenerateAISearchAssistantInput,
  ) => Promise<APIResponse<AISearchAssistantResult>>;
  generateAIJobFitSummary: (
    input: GenerateAIJobFitSummaryInput,
  ) => Promise<APIResponse<AIJobFitSummaryResult>>;
  generateAICoverLetterDraft: (
    input: GenerateAICoverLetterDraftInput,
  ) => Promise<APIResponse<AICoverLetterDraftResult>>;
}

let refreshInFlight: Promise<string | null> | null = null;

/**
 * createSessionAPIClient creates session api client.
 */
export function createSessionAPIClient(
  dependencies?: Partial<SessionClientDependencies>,
): SessionAPIClient {
  const deps = { ...defaultDependencies, ...(dependencies ?? {}) };

  async function withAuthorizedRequest<T>(
    request: (accessToken: string) => Promise<APIResponse<T>>,
  ): Promise<APIResponse<T>> {
    const accessToken = await ensureAccessToken(deps);

    try {
      return await request(accessToken);
    } catch (error) {
      if (!(error instanceof APIRequestError) || error.status !== 401) {
        throw error;
      }

      const refreshedToken = await refreshAccessTokenSingleFlight(deps);
      if (!refreshedToken) {
        throw new APIRequestError("Session expired", 401, "UNAUTHORIZED");
      }
      return request(refreshedToken);
    }
  }

  return {
    getMe: () => withAuthorizedRequest((token) => getMe(token)),
    getBillingStatus: () =>
      withAuthorizedRequest((token) => getBillingStatus(token)),
    createCheckoutSession: (input) =>
      withAuthorizedRequest((token) => createCheckoutSession(token, input)),
    getBillingTransactions: (query) =>
      withAuthorizedRequest((token) => getBillingTransactions(token, query)),
    getPreferences: () =>
      withAuthorizedRequest((token) => getPreferences(token)),
    updatePreferences: (input) =>
      withAuthorizedRequest((token) => updatePreferences(token, input)),
    listSavedSearches: () =>
      withAuthorizedRequest((token) => listSavedSearches(token)),
    createSavedSearch: (input) =>
      withAuthorizedRequest((token) => createSavedSearch(token, input)),
    deleteSavedSearch: (id) =>
      withAuthorizedRequest((token) => deleteSavedSearch(token, id)),
    listNotifications: (query) =>
      withAuthorizedRequest((token) => listNotifications(token, query)),
    markNotificationAsRead: (notificationID) =>
      withAuthorizedRequest((token) =>
        markNotificationAsRead(token, notificationID),
      ),
    updateNotificationPreferences: (input) =>
      withAuthorizedRequest((token) =>
        updateNotificationPreferences(token, input),
      ),
    getAIUsage: (feature) =>
      withAuthorizedRequest((token) => getAIUsage(token, feature)),
    generateAISearchAssistant: (input) =>
      withAuthorizedRequest((token) => generateAISearchAssistant(token, input)),
    generateAIJobFitSummary: (input) =>
      withAuthorizedRequest((token) => generateAIJobFitSummary(token, input)),
    generateAICoverLetterDraft: (input) =>
      withAuthorizedRequest((token) =>
        generateAICoverLetterDraft(token, input),
      ),
  };
}

/**
 * resetSessionRefreshStateForTests resets session refresh state for tests.
 */
export function resetSessionRefreshStateForTests(): void {
  refreshInFlight = null;
}

async function ensureAccessToken(
  deps: SessionClientDependencies,
): Promise<string> {
  const snapshot = deps.getSession();

  if (!snapshot.accessToken) {
    const refreshedToken = await refreshAccessTokenSingleFlight(deps);
    if (!refreshedToken) {
      throw new APIRequestError("Session expired", 401, "UNAUTHORIZED");
    }
    return refreshedToken;
  }

  if (shouldRefreshSoon(snapshot.accessExpiresAt)) {
    const refreshedToken = await refreshAccessTokenSingleFlight(deps);
    return refreshedToken ?? snapshot.accessToken;
  }

  return snapshot.accessToken;
}

function shouldRefreshSoon(accessExpiresAt: number | null): boolean {
  if (!accessExpiresAt) {
    return false;
  }
  const refreshThresholdMillis = ACCESS_REFRESH_WINDOW_SECONDS * 1000;
  return Date.now() >= accessExpiresAt - refreshThresholdMillis;
}

async function refreshAccessTokenSingleFlight(
  deps: SessionClientDependencies,
): Promise<string | null> {
  if (refreshInFlight) {
    return refreshInFlight;
  }

  const session = deps.getSession();
  const refreshToken = session.refreshToken;
  if (!refreshToken) {
    deps.clearSession();
    return null;
  }

  refreshInFlight = (async () => {
    try {
      const response = await deps.refresh({
        refresh_token: refreshToken,
      });

      deps.updateAccessToken({
        accessToken: response.data.access_token,
        expiresIn: response.data.expires_in,
      });
      return response.data.access_token;
    } catch {
      deps.clearSession();
      return null;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}
