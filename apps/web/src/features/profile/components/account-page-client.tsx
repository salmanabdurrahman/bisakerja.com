"use client";

import { useRouter } from "next/navigation";

import { SubscriptionBadge } from "@/features/billing/components/subscription-badge";
import { useAuthSession } from "@/features/auth/session-provider";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { ProfileSummary } from "@/features/profile/components/profile-summary";
import type { AuthMe, SubscriptionState } from "@/services/auth";

type BadgeState = SubscriptionState | "status_unavailable";
type BadgeSource = "billing" | "profile_fallback" | "unavailable";

interface AccountPageClientProps {
  profile: AuthMe;
  badgeState: BadgeState;
  badgeSource: BadgeSource;
  warningMessage: string | null;
}

export function AccountPageClient({
  profile,
  badgeState,
  badgeSource,
  warningMessage,
}: AccountPageClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();

  async function handleLogout() {
    clearBrowserSession();
    markAnonymous();
    router.replace("/auth/login");
  }

  return (
    <section className="grid gap-4">
      <div className="grid gap-2 rounded-lg border border-gray-200 p-4">
        <h3 className="text-lg font-semibold text-gray-900">
          Subscription status
        </h3>
        <SubscriptionBadge state={badgeState} source={badgeSource} />
        {warningMessage ? (
          <p className="text-sm text-amber-700">{warningMessage}</p>
        ) : null}
      </div>

      <ProfileSummary profile={profile} />

      <div className="flex flex-wrap gap-3">
        <a
          href="/account/preferences"
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-800 hover:bg-gray-50"
        >
          Kelola preferences
        </a>
        <a
          href="/account/saved-searches"
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-800 hover:bg-gray-50"
        >
          Saved searches
        </a>
        <a
          href="/account/notifications"
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-800 hover:bg-gray-50"
        >
          Notification center
        </a>
        <button
          type="button"
          onClick={() => void handleLogout()}
          className="rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Logout
        </button>
      </div>
    </section>
  );
}
