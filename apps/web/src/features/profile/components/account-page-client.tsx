"use client";

import { useRouter } from "next/navigation";

import { Button, ButtonLink } from "@/components/ui/button";
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
      <div className="bk-card grid gap-4 p-6 sm:p-8">
        <h3 className="text-lg font-semibold text-slate-900">
          Subscription status
        </h3>
        <SubscriptionBadge state={badgeState} source={badgeSource} />
        {warningMessage ? (
          <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
            {warningMessage}
          </p>
        ) : null}
      </div>

      <ProfileSummary profile={profile} />

      <div className="flex flex-wrap gap-3">
        <ButtonLink href="/account/preferences" variant="outline">
          Manage preferences
        </ButtonLink>
        <ButtonLink href="/account/saved-searches" variant="outline">
          Saved searches
        </ButtonLink>
        <ButtonLink href="/account/notifications" variant="outline">
          Notification center
        </ButtonLink>
        <ButtonLink href="/account/ai-tools" variant="outline">
          AI tools
        </ButtonLink>
        <Button
          type="button"
          onClick={() => void handleLogout()}
          variant="primary"
        >
          Logout
        </Button>
      </div>
    </section>
  );
}
