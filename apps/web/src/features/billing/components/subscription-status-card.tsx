import type { BillingStatus, TransactionStatus } from "@/services/billing";
import type { SubscriptionState } from "@/services/auth";

interface SubscriptionStatusCardProps {
  subscriptionState: SubscriptionState | "status_unavailable";
  lastTransactionStatus?: TransactionStatus;
  premiumExpiredAt?: string | null;
  source?: "billing" | "profile_fallback" | "unavailable";
}

const subscriptionLabels: Record<
  SubscriptionStatusCardProps["subscriptionState"],
  string
> = {
  free: "Free",
  pending_payment: "Pending payment",
  premium_active: "Premium active",
  premium_expired: "Premium expired",
  status_unavailable: "Status unavailable",
};

const subscriptionDescriptions: Record<
  SubscriptionStatusCardProps["subscriptionState"],
  string
> = {
  free: "Your account is currently on the free plan. Upgrade to enable premium notifications.",
  pending_payment:
    "Payment is still pending. Continue checkout or wait for status confirmation.",
  premium_active: "Premium is active. You can use all premium features.",
  premium_expired:
    "Premium has expired. Upgrade again to reactivate premium features.",
  status_unavailable:
    "We could not confirm your premium status right now. Please refresh shortly.",
};

const sourceLabel: Record<
  NonNullable<SubscriptionStatusCardProps["source"]>,
  string
> = {
  billing: "Live billing data",
  profile_fallback: "Profile fallback",
  unavailable: "Temporarily unavailable",
};

export function SubscriptionStatusCard({
  subscriptionState,
  lastTransactionStatus,
  premiumExpiredAt,
  source = "billing",
}: SubscriptionStatusCardProps) {
  return (
    <section className="bk-card grid gap-4 p-6 sm:p-8">
      <h3 className="bk-heading-card">Subscription overview</h3>
      <p className="bk-body">
        Current plan:{" "}
        <span className="rounded-full border border-[#E5E5E5] bg-[#F9F9F9] px-3 py-1 font-medium text-black">
          {subscriptionLabels[subscriptionState]}
        </span>
      </p>
      <p className="bk-body">{subscriptionDescriptions[subscriptionState]}</p>
      {lastTransactionStatus ? (
        <p className="bk-body">
          Last transaction status:{" "}
          <span className="font-medium text-black">
            {lastTransactionStatus}
          </span>
        </p>
      ) : null}
      {premiumExpiredAt ? (
        <p className="bk-body">
          Premium expiry:{" "}
          <span className="text-black">
            {new Date(premiumExpiredAt).toLocaleString("en-US")}
          </span>
        </p>
      ) : null}
      <p className="bk-body-sm mt-2 font-medium uppercase tracking-wider text-[#888888]">
        Status sync: {sourceLabel[source]}
      </p>
    </section>
  );
}

export function toSubscriptionCardProps(
  status: BillingStatus,
): SubscriptionStatusCardProps {
  return {
    subscriptionState: status.subscription_state,
    lastTransactionStatus: status.last_transaction_status,
    premiumExpiredAt: status.premium_expired_at,
    source: "billing",
  };
}
