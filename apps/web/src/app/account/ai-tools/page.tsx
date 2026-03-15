import { redirect } from "next/navigation";

import { AccountAIToolsClient } from "@/features/ai/components/account-ai-tools-client";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, type SubscriptionState } from "@/services/auth";
import { getBillingStatus } from "@/services/billing";

interface AIToolsViewState {
  subscriptionState: SubscriptionState | "status_unavailable";
  infoMessage: string | null;
}

export default async function AccountAIToolsPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/ai-tools"));
  }

  const viewState = await loadAIToolsViewState(accessToken);

  return (
    <AccountDashboardShell
      eyebrow="AI Copilot"
      title="AI tools"
      description="Use AI suggestions for search strategy, job-fit insights, and cover letter drafting."
    >
      <AccountAIToolsClient
        subscriptionState={viewState.subscriptionState}
        infoMessage={viewState.infoMessage}
      />
    </AccountDashboardShell>
  );
}

async function loadAIToolsViewState(
  accessToken: string,
): Promise<AIToolsViewState> {
  let profileSubscriptionState: SubscriptionState | null = null;
  let fallbackPremium = false;

  try {
    const profileResponse = await getMe(accessToken);
    profileSubscriptionState = profileResponse.data.subscription_state ?? null;
    fallbackPremium = profileResponse.data.is_premium;
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/ai-tools"));
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
      redirect(buildLoginHref("/account/ai-tools"));
    }
    infoMessage =
      "We could not refresh premium status right now. AI usage and capability may be stale.";
    if (!profileSubscriptionState && !fallbackPremium) {
      subscriptionState = "status_unavailable";
    }
  }

  return {
    subscriptionState,
    infoMessage,
  };
}
