import type { SubscriptionState } from "@/services/auth";

interface NotificationEntitlementBannerProps {
  subscriptionState: SubscriptionState | "status_unavailable";
}

export function NotificationEntitlementBanner({
  subscriptionState,
}: NotificationEntitlementBannerProps) {
  if (subscriptionState === "premium_active") {
    return (
      <section className="rounded-lg border border-emerald-200 bg-emerald-50 p-4">
        <h3 className="text-base font-semibold text-emerald-900">
          Premium notifications are active
        </h3>
        <p className="mt-1 text-sm text-emerald-800">
          Saved preferences are used for job matching
          automatically.
        </p>
      </section>
    );
  }

  if (subscriptionState === "pending_payment") {
    return (
      <section className="rounded-lg border border-amber-200 bg-amber-50 p-4">
        <h3 className="text-base font-semibold text-amber-900">
          Payment is being processed
        </h3>
        <p className="mt-1 text-sm text-amber-800">
          Notifications will be enabled after successful payment.
        </p>
        <a
          href="/pricing"
          className="mt-3 inline-flex text-sm text-blue-700 underline"
        >
          Continue payment
        </a>
      </section>
    );
  }

  if (subscriptionState === "premium_expired") {
    return (
      <section className="rounded-lg border border-orange-200 bg-orange-50 p-4">
        <h3 className="text-base font-semibold text-orange-900">
          Premium has expired
        </h3>
        <p className="mt-1 text-sm text-orange-800">
          Renew your subscription to reactivate matching notifications.
        </p>
        <a
          href="/pricing"
          className="mt-3 inline-flex text-sm text-blue-700 underline"
        >
          Upgrade again
        </a>
      </section>
    );
  }

  if (subscriptionState === "status_unavailable") {
    return (
      <section className="rounded-lg border border-red-200 bg-red-50 p-4">
        <h3 className="text-base font-semibold text-red-900">
          Premium status unavailable
        </h3>
        <p className="mt-1 text-sm text-red-800">
          We couldn&apos;t fetch the latest billing status. Please refresh
          the page.
        </p>
      </section>
    );
  }

  return (
    <section className="rounded-lg border border-gray-200 bg-gray-50 p-4">
      <h3 className="text-base font-semibold text-gray-900">
        Premium notifications
      </h3>
      <p className="mt-1 text-sm text-gray-700">
        Free users can still save preferences, but matching notifications
        are only active for premium users.
      </p>
      <a
        href="/pricing"
        className="mt-3 inline-flex text-sm text-blue-700 underline"
      >
        View premium plans
      </a>
    </section>
  );
}
