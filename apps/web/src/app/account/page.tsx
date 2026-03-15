import { ButtonLink } from "@/components/ui/button";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
import { AccountPageClient } from "@/features/profile/components/account-page-client";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, type SubscriptionState } from "@/services/auth";
import { getBillingStatus } from "@/services/billing";
import { redirect } from "next/navigation";

type BadgeState = SubscriptionState | "status_unavailable";
type BadgeSource = "billing" | "profile_fallback" | "unavailable";
type AccountPageViewState =
  | {
      kind: "ready";
      profile: Awaited<ReturnType<typeof getMe>>["data"];
      badgeState: BadgeState;
      badgeSource: BadgeSource;
      warningMessage: string | null;
    }
  | { kind: "error" };

export default async function AccountPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account"));
  }

  const viewState = await loadAccountViewState(accessToken);

  return (
    <AccountDashboardShell
      eyebrow="Account"
      title="Account"
      description="Review profile identity, premium entitlement, and shortcuts to growth settings."
    >
      {renderAccountView(viewState)}
    </AccountDashboardShell>
  );
}

function deriveFallbackBadgeState(profile: {
  subscription_state?: SubscriptionState;
  is_premium: boolean;
}): BadgeState {
  if (profile.subscription_state) {
    return profile.subscription_state;
  }
  return profile.is_premium ? "premium_active" : "free";
}

async function loadAccountViewState(
  accessToken: string,
): Promise<AccountPageViewState> {
  try {
    const profileResponse = await getMe(accessToken);
    let badgeState: BadgeState = deriveFallbackBadgeState(profileResponse.data);
    let badgeSource: BadgeSource = "profile_fallback";
    let warningMessage: string | null = null;

    try {
      const billingResponse = await getBillingStatus(accessToken);
      badgeState = billingResponse.data.subscription_state;
      badgeSource = "billing";
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        redirect(buildLoginHref("/account"));
      }
      warningMessage =
        "We could not refresh your latest subscription status. Showing your last known account data.";
      badgeSource = "profile_fallback";
    }

    return {
      kind: "ready",
      profile: profileResponse.data,
      badgeState,
      badgeSource,
      warningMessage,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account"));
    }
    return { kind: "error" };
  }
}

function renderAccountView(viewState: AccountPageViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
        <h3 className="bk-heading-card text-red-900">Failed to load account</h3>
        <p className="bk-body text-red-800">
          Account data is currently unavailable. Please refresh the page.
        </p>
        <ButtonLink href="/account" variant="danger">
          Try again
        </ButtonLink>
      </section>
    );
  }

  return (
    <AccountPageClient
      profile={viewState.profile}
      badgeState={viewState.badgeState}
      badgeSource={viewState.badgeSource}
      warningMessage={viewState.warningMessage}
    />
  );
}
