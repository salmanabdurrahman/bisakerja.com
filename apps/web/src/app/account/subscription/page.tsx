import { ButtonLink } from "@/components/ui/button";
import { BillingHistoryList } from "@/features/billing/components/billing-history-list";
import { SubscriptionStatusCard } from "@/features/billing/components/subscription-status-card";
import { UpgradeCTA } from "@/features/billing/components/upgrade-cta";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getBillingStatus, getBillingTransactions } from "@/services/billing";
import { redirect } from "next/navigation";

type SubscriptionViewState =
  | {
      kind: "ready";
      status: Awaited<ReturnType<typeof getBillingStatus>>["data"];
      transactions: Awaited<ReturnType<typeof getBillingTransactions>>["data"];
      warningMessage: string | null;
    }
  | { kind: "error" };

export default async function AccountSubscriptionPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/subscription"));
  }

  const viewState = await loadSubscriptionViewState(accessToken);

  return (
    <AccountDashboardShell
      eyebrow="Monetization"
      title="Subscription"
      description="Track canonical subscription state, latest transactions, and available upgrade actions."
    >
      {renderSubscriptionView(viewState)}
    </AccountDashboardShell>
  );
}

async function loadSubscriptionViewState(
  accessToken: string,
): Promise<SubscriptionViewState> {
  try {
    const statusResponse = await getBillingStatus(accessToken);
    let transactions: Awaited<
      ReturnType<typeof getBillingTransactions>
    >["data"] = [];
    let warningMessage: string | null = null;

    try {
      const transactionsResponse = await getBillingTransactions(accessToken, {
        page: 1,
        limit: 20,
      });
      transactions = transactionsResponse.data;
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        redirect(buildLoginHref("/account/subscription"));
      }
      warningMessage =
        "Transaction history is currently unavailable. Please refresh shortly.";
    }

    return {
      kind: "ready",
      status: statusResponse.data,
      transactions,
      warningMessage,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/subscription"));
    }
    return { kind: "error" };
  }
}

function renderSubscriptionView(viewState: SubscriptionViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
        <h3 className="text-lg font-semibold text-red-900">
          Subscription data unavailable
        </h3>
        <p className="text-sm text-red-800">
          Subscription status could not be loaded. Please try again shortly.
        </p>
        <ButtonLink href="/account/subscription" variant="danger">
          Retry
        </ButtonLink>
      </section>
    );
  }

  return (
    <>
      <SubscriptionStatusCard
        subscriptionState={viewState.status.subscription_state}
        lastTransactionStatus={viewState.status.last_transaction_status}
        premiumExpiredAt={viewState.status.premium_expired_at}
      />
      <UpgradeCTA
        subscriptionState={viewState.status.subscription_state}
        lastTransactionStatus={viewState.status.last_transaction_status}
      />
      <BillingHistoryList
        transactions={viewState.transactions}
        warningMessage={viewState.warningMessage}
      />
    </>
  );
}
