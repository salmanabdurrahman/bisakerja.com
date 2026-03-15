import { SubscriptionBadge } from "@/features/billing/components/subscription-badge";
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
  return (
    <section className="grid gap-4">
      <div className="bk-card grid gap-4 p-6 sm:p-8">
        <h3 className="bk-heading-card">Subscription status</h3>
        <SubscriptionBadge state={badgeState} source={badgeSource} />
        {warningMessage ? (
          <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 bk-body text-amber-800">
            {warningMessage}
          </p>
        ) : null}
      </div>

      <ProfileSummary profile={profile} />
    </section>
  );
}
