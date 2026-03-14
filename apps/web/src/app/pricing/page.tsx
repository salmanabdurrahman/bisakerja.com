import { AppShell } from "@/components/layout/app-shell";
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
      <main className="grid gap-4" role="main">
        <h2 className="text-xl font-semibold">Pricing</h2>
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
        "Status billing belum tersedia. Menampilkan fallback sementara dari profile.",
    };
  }
}

function renderPricingView(viewState: PricingViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="grid gap-3 rounded-lg border border-red-200 bg-red-50 p-4">
        <h3 className="text-lg font-semibold text-red-900">
          Pricing unavailable
        </h3>
        <p className="text-sm text-red-800">
          Data pricing belum bisa dimuat. Coba lagi beberapa saat.
        </p>
        <a
          href="/pricing"
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
      <a
        href="/account/subscription"
        className="text-sm text-blue-700 underline"
      >
        Lihat status subscription detail
      </a>
    </>
  );
}
