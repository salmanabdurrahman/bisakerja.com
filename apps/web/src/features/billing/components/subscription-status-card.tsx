import type { BillingStatus, TransactionStatus } from "@/services/billing";
import type { SubscriptionState } from "@/services/auth";

interface SubscriptionStatusCardProps {
  subscriptionState: SubscriptionState | "status_unavailable";
  lastTransactionStatus?: TransactionStatus;
  premiumExpiredAt?: string | null;
  source?: "billing" | "profile_fallback" | "unavailable";
}

const subscriptionDescriptions: Record<
  SubscriptionStatusCardProps["subscriptionState"],
  string
> = {
  free: "Your account is currently on the free plan. Upgrade to enable premium notifications.",
  pending_payment:
    "Payment is still pending. Continue payment or wait for webhook synchronization.",
  premium_active: "Premium is active. You can use all premium features.",
  premium_expired:
    "Premium has expired. Upgrade again to reactivate premium features.",
  status_unavailable:
    "Subscription status cannot be verified from billing/status right now.",
};

export function SubscriptionStatusCard({
  subscriptionState,
  lastTransactionStatus,
  premiumExpiredAt,
  source = "billing",
}: SubscriptionStatusCardProps) {
  return (
    <section className="grid gap-2 rounded-lg border border-gray-200 p-4">
      <h3 className="text-lg font-semibold text-gray-900">
        Subscription overview
      </h3>
      <p className="text-sm text-gray-700">
        State: <span className="font-semibold">{subscriptionState}</span>
      </p>
      <p className="text-sm text-gray-700">
        {subscriptionDescriptions[subscriptionState]}
      </p>
      {lastTransactionStatus ? (
        <p className="text-sm text-gray-700">
          Last transaction status:{" "}
          <span className="font-medium">{lastTransactionStatus}</span>
        </p>
      ) : null}
      {premiumExpiredAt ? (
        <p className="text-sm text-gray-700">
          Premium expiry: {new Date(premiumExpiredAt).toLocaleString("en-US")}
        </p>
      ) : null}
      <p className="text-xs text-gray-500">Source: {source}</p>
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
