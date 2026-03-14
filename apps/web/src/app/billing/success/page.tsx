import { AppShell } from "@/components/layout/app-shell";
import { ButtonLink } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";
import { SubscriptionStatusCard } from "@/features/billing/components/subscription-status-card";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getBillingStatus } from "@/services/billing";
import { redirect } from "next/navigation";

type VerifyState =
  | {
      kind: "upgrade_success";
      status: Awaited<ReturnType<typeof getBillingStatus>>["data"];
    }
  | {
      kind: "upgrade_pending";
      status: Awaited<ReturnType<typeof getBillingStatus>>["data"];
    }
  | {
      kind: "upgrade_reoffer";
      status: Awaited<ReturnType<typeof getBillingStatus>>["data"];
    }
  | { kind: "verify_error" };

export default async function BillingSuccessPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/billing/success"));
  }

  const verifyState = await loadVerifyState(accessToken);

  return (
    <AppShell>
      <main className="grid gap-5" role="main">
        <PageHeader
          eyebrow="Billing Callback"
          title="Payment verification"
          description="We verify payment and refresh your premium status automatically."
        />
        {renderVerifyState(verifyState)}
      </main>
    </AppShell>
  );
}

async function loadVerifyState(accessToken: string): Promise<VerifyState> {
  try {
    const statusResponse = await getBillingStatus(accessToken);
    const status = statusResponse.data;

    if (status.subscription_state === "premium_active") {
      return { kind: "upgrade_success", status };
    }
    if (status.subscription_state === "pending_payment") {
      return { kind: "upgrade_pending", status };
    }
    if (
      (status.subscription_state === "free" ||
        status.subscription_state === "premium_expired") &&
      status.last_transaction_status === "failed"
    ) {
      return { kind: "upgrade_reoffer", status };
    }
    return { kind: "upgrade_pending", status };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/billing/success"));
    }
    return { kind: "verify_error" };
  }
}

function renderVerifyState(state: VerifyState) {
  if (state.kind === "verify_error") {
    return (
      <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
        <h3 className="text-lg font-semibold text-red-900">
          Verification unavailable
        </h3>
        <p className="text-sm text-red-800">
          We are unable to verify your payment status right now.
        </p>
        <ButtonLink href="/billing/success" variant="danger">
          Try again
        </ButtonLink>
      </section>
    );
  }

  if (state.kind === "upgrade_success") {
    return (
      <section className="grid gap-3">
        <SubscriptionStatusCard
          subscriptionState={state.status.subscription_state}
          lastTransactionStatus={state.status.last_transaction_status}
          premiumExpiredAt={state.status.premium_expired_at}
        />
        <p className="text-sm text-emerald-700">
          Payment verified. Your premium subscription is now active.
        </p>
        <ButtonLink href="/account/subscription" variant="secondary" size="sm">
          View subscription details
        </ButtonLink>
      </section>
    );
  }

  if (state.kind === "upgrade_reoffer") {
    return (
      <section className="grid gap-3">
        <SubscriptionStatusCard
          subscriptionState={state.status.subscription_state}
          lastTransactionStatus={state.status.last_transaction_status}
          premiumExpiredAt={state.status.premium_expired_at}
        />
        <p className="text-sm text-amber-700">
          Payment has not completed yet. You can start a new checkout.
        </p>
        <ButtonLink href="/pricing" variant="outline" size="sm">
          Back to pricing
        </ButtonLink>
      </section>
    );
  }

  return (
    <section className="grid gap-3">
      <SubscriptionStatusCard
        subscriptionState={state.status.subscription_state}
        lastTransactionStatus={state.status.last_transaction_status}
        premiumExpiredAt={state.status.premium_expired_at}
      />
      <p className="text-sm text-amber-700">
        Payment is still being processed. Check the status again shortly.
      </p>
      <ButtonLink href="/billing/success" variant="outline" size="sm">
        Refresh status
      </ButtonLink>
    </section>
  );
}
