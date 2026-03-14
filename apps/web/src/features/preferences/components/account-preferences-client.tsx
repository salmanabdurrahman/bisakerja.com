"use client";

import { useMemo } from "react";
import { useRouter } from "next/navigation";

import { useAuthSession } from "@/features/auth/session-provider";
import { NotificationEntitlementBanner } from "@/features/preferences/components/notification-entitlement-banner";
import { PreferencesForm } from "@/features/preferences/components/preferences-form";
import {
  clearPreferencesDraft,
  savePreferencesDraft,
} from "@/features/preferences/draft";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { createSessionAPIClient } from "@/services/session-api-client";
import type { SubscriptionState } from "@/services/auth";
import type { UpdatePreferencesInput } from "@/services/preferences";

interface AccountPreferencesClientProps {
  initialPreferences: UpdatePreferencesInput;
  initialUpdatedAt?: string | null;
  subscriptionState: SubscriptionState | "status_unavailable";
  infoMessage: string | null;
}

export function AccountPreferencesClient({
  initialPreferences,
  initialUpdatedAt = null,
  subscriptionState,
  infoMessage,
}: AccountPreferencesClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);

  return (
    <section className="grid gap-4">
      <NotificationEntitlementBanner subscriptionState={subscriptionState} />
      {infoMessage ? (
        <p className="text-sm text-amber-700">{infoMessage}</p>
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
          clearBrowserSession();
          markAnonymous();
          router.replace(buildLoginHref("/account/preferences"));
        }}
      />
    </section>
  );
}
