import { AppShell } from "@/components/layout/app-shell";
import { BillingHistoryList } from "@/features/billing/components/billing-history-list";
import { SubscriptionStatusCard } from "@/features/billing/components/subscription-status-card";
import { UpgradeCTA } from "@/features/billing/components/upgrade-cta";
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
    <AppShell>
      <main className="grid gap-4" role="main">
        <h2 className="text-xl font-semibold">Subscription</h2>
        {renderSubscriptionView(viewState)}
      </main>
    </AppShell>
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
      <section className="grid gap-3 rounded-lg border border-red-200 bg-red-50 p-4">
        <h3 className="text-lg font-semibold text-red-900">
          Subscription data unavailable
        </h3>
        <p className="text-sm text-red-800">
          Subscription status could not be loaded. Please try again shortly.
        </p>
        <a
          href="/account/subscription"
          className="w-fit rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Retry
        </a>
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
