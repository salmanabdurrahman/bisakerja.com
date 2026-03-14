import { AppShell } from "@/components/layout/app-shell";
import { ButtonLink } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";
import { SubscriptionStatusCard } from "@/features/billing/components/subscription-status-card";
import { UpgradeCTA } from "@/features/billing/components/upgrade-cta";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, type SubscriptionState } from "@/services/auth";
import { getBillingStatus, type TransactionStatus } from "@/services/billing";
import { redirect } from "next/navigation";

type PricingViewState =
  | {
      kind: "ready";
      subscriptionState: SubscriptionState | "status_unavailable";
      lastTransactionStatus?: TransactionStatus;
      premiumExpiredAt?: string | null;
      source: "billing" | "profile_fallback" | "unavailable";
      warningMessage: string | null;
    }
  | { kind: "error" };

export default async function PricingPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/pricing"));
  }

  const viewState = await loadPricingViewState(accessToken);

  return (
    <AppShell>
      <main className="grid gap-6 mx-auto w-full" role="main">
        <PageHeader
          eyebrow="Monetization"
          title="Pricing"
          description="Start checkout, monitor payment progress, and keep your premium plan in sync."
        />
        {renderPricingView(viewState)}
      </main>
    </AppShell>
  );
}

async function loadPricingViewState(
  accessToken: string,
): Promise<PricingViewState> {
  let profileFallbackState: SubscriptionState = "free";

  try {
    const meResponse = await getMe(accessToken);
    if (meResponse.data.subscription_state) {
      profileFallbackState = meResponse.data.subscription_state;
    } else {
      profileFallbackState = meResponse.data.is_premium
        ? "premium_active"
        : "free";
    }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/pricing"));
    }
  }

  try {
    const billingStatusResponse = await getBillingStatus(accessToken);
    return {
      kind: "ready",
      subscriptionState: billingStatusResponse.data.subscription_state,
      lastTransactionStatus: billingStatusResponse.data.last_transaction_status,
      premiumExpiredAt: billingStatusResponse.data.premium_expired_at,
      source: "billing",
      warningMessage: null,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/pricing"));
    }

    return {
      kind: "ready",
      subscriptionState: profileFallbackState,
      source: "profile_fallback",
      warningMessage:
        "Live billing status is temporarily unavailable. Showing your last known account status.",
    };
  }
}

function renderPricingView(viewState: PricingViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="bk-card grid gap-4 border-red-200 bg-red-50 p-6 sm:p-8">
        <h3 className="text-[24px] font-normal text-red-900">
          Pricing unavailable
        </h3>
        <p className="text-[14px] text-red-800">
          Pricing data is currently unavailable. Please try again shortly.
        </p>
        <ButtonLink href="/pricing" variant="danger">
          Retry
        </ButtonLink>
      </section>
    );
  }

  return (
    <>
      <SubscriptionStatusCard
        subscriptionState={viewState.subscriptionState}
        lastTransactionStatus={viewState.lastTransactionStatus}
        premiumExpiredAt={viewState.premiumExpiredAt}
        source={viewState.source}
      />
      {viewState.warningMessage ? (
        <p className="text-sm text-amber-700">{viewState.warningMessage}</p>
      ) : null}
      <UpgradeCTA
        subscriptionState={viewState.subscriptionState}
        lastTransactionStatus={viewState.lastTransactionStatus}
      />
      <ButtonLink href="/account/subscription" variant="outline" size="sm">
        View detailed subscription status
      </ButtonLink>
    </>
  );
}
