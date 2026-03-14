"use client";

import { useMemo } from "react";
import { useRouter } from "next/navigation";

import { useAuthSession } from "@/features/auth/session-provider";
import { NotificationEntitlementBanner } from "@/features/preferences/components/notification-entitlement-banner";
import { NotificationDigestControl } from "@/features/preferences/components/notification-digest-control";
import { PreferencesForm } from "@/features/preferences/components/preferences-form";
import {
  clearPreferencesDraft,
  savePreferencesDraft,
} from "@/features/preferences/draft";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { createSessionAPIClient } from "@/services/session-api-client";
import type { SubscriptionState } from "@/services/auth";
import type {
  NotificationAlertMode,
  UpdatePreferencesInput,
} from "@/services/preferences";

interface AccountPreferencesClientProps {
  initialPreferences: UpdatePreferencesInput;
  initialUpdatedAt?: string | null;
  initialNotificationSettings: {
    alert_mode: NotificationAlertMode;
    digest_hour?: number | null;
    updated_at?: string | null;
  };
  subscriptionState: SubscriptionState | "status_unavailable";
  infoMessage: string | null;
}

export function AccountPreferencesClient({
  initialPreferences,
  initialUpdatedAt = null,
  initialNotificationSettings,
  subscriptionState,
  infoMessage,
}: AccountPreferencesClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);

  function handleUnauthorizedRedirect() {
    clearBrowserSession();
    markAnonymous();
    router.replace(buildLoginHref("/account/preferences"));
  }

  return (
    <section className="grid gap-4">
      <NotificationEntitlementBanner subscriptionState={subscriptionState} />
      {infoMessage ? (
        <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          {infoMessage}
        </p>
      ) : null}
      <PreferencesForm
        initial={initialPreferences}
        initialUpdatedAt={initialUpdatedAt}
        onSubmit={async (payload) => {
          const response = await sessionClient.updatePreferences(payload);
          clearPreferencesDraft();
          return { updated_at: response.data.updated_at };
        }}
        onUnauthorized={(draft) => {
          savePreferencesDraft(draft);
          handleUnauthorizedRedirect();
        }}
      />
      <NotificationDigestControl
        initialSettings={initialNotificationSettings}
        onSubmit={(payload) =>
          sessionClient
            .updateNotificationPreferences(payload)
            .then((response) => response.data)
        }
        onUnauthorized={handleUnauthorizedRedirect}
      />
    </section>
  );
}
