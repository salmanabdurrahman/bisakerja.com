import { ButtonLink } from "@/components/ui/button";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
import { AccountPreferencesClient } from "@/features/preferences/components/account-preferences-client";
import {
  loadPreferencesDraftFromCookie,
  preferencesDraftCookie,
} from "@/features/preferences/draft";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, type SubscriptionState } from "@/services/auth";
import { getBillingStatus } from "@/services/billing";
import type { UpdatePreferencesInput } from "@/services/preferences";
import { getPreferences } from "@/services/preferences";
import type { NotificationAlertMode } from "@/services/preferences";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

export default async function AccountPreferencesPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/preferences"));
  }

  const cookieStore = await cookies();
  const cookieDraft = loadPreferencesDraftFromCookie(
    cookieStore.get(preferencesDraftCookie)?.value,
  );

  let profileSubscriptionState: SubscriptionState | null = null;
  let fallbackPremium = false;
  try {
    const profileResponse = await getMe(accessToken);
    profileSubscriptionState = profileResponse.data.subscription_state ?? null;
    fallbackPremium = profileResponse.data.is_premium;
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/preferences"));
    }
  }

  let subscriptionState: SubscriptionState | "status_unavailable" =
    profileSubscriptionState ?? (fallbackPremium ? "premium_active" : "free");
  let infoMessage: string | null = null;

  try {
    const billingResponse = await getBillingStatus(accessToken);
    subscriptionState = billingResponse.data.subscription_state;
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/preferences"));
    }
    infoMessage =
      "We could not refresh premium status right now. Using your last known account status.";
    if (!profileSubscriptionState && !fallbackPremium) {
      subscriptionState = "status_unavailable";
    }
  }

  let initialPreferences: UpdatePreferencesInput | null = cookieDraft;
  let initialUpdatedAt: string | null = null;
  let initialNotificationSettings: {
    alert_mode: NotificationAlertMode;
    digest_hour?: number | null;
    updated_at?: string | null;
  } = {
    alert_mode: "instant",
    digest_hour: null,
    updated_at: null,
  };

  try {
    const preferencesResponse = await getPreferences(accessToken);
    initialNotificationSettings = {
      alert_mode: preferencesResponse.data.alert_mode,
      digest_hour: preferencesResponse.data.digest_hour ?? null,
      updated_at: preferencesResponse.data.updated_at ?? null,
    };

    if (!initialPreferences) {
      initialPreferences = {
        keywords: preferencesResponse.data.keywords,
        locations: preferencesResponse.data.locations,
        job_types: preferencesResponse.data.job_types,
        salary_min: preferencesResponse.data.salary_min,
      };
      initialUpdatedAt = preferencesResponse.data.updated_at ?? null;
    }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/preferences"));
    }
  }

  if (cookieDraft) {
    infoMessage =
      "Local draft restored successfully. Save preferences again to sync with the server.";
  }

  if (!initialPreferences) {
    return (
      <AccountDashboardShell
        eyebrow="Preferences"
        title="Preferences"
        description="Configure keywords, locations, and job types to power personalized matching."
      >
        <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
          <h3 className="bk-heading-card text-red-900">
            Failed to load preferences
          </h3>
          <p className="bk-body text-red-800">
            Preferences data is currently unavailable. Please refresh the page.
          </p>
          <ButtonLink href="/account/preferences" variant="danger">
            Try again
          </ButtonLink>
        </section>
      </AccountDashboardShell>
    );
  }

  return (
    <AccountDashboardShell
      eyebrow="Preferences"
      title="Preferences"
      description="Configure keywords, locations, and job types to power personalized matching."
    >
      <AccountPreferencesClient
        initialPreferences={initialPreferences}
        initialUpdatedAt={initialUpdatedAt}
        initialNotificationSettings={initialNotificationSettings}
        subscriptionState={subscriptionState}
        infoMessage={infoMessage}
      />
    </AccountDashboardShell>
  );
}
