import type { SubscriptionState } from "@/services/auth";
import { ButtonLink } from "@/components/ui/button";

interface NotificationEntitlementBannerProps {
  subscriptionState: SubscriptionState | "status_unavailable";
}

export function NotificationEntitlementBanner({
  subscriptionState,
}: NotificationEntitlementBannerProps) {
  if (subscriptionState === "premium_active") {
    return (
      <section className="bk-card rounded-2xl border-emerald-200 bg-emerald-50 p-5">
        <h3 className="text-base font-semibold text-emerald-900">
          Premium notifications are active
        </h3>
        <p className="mt-1 text-sm text-emerald-800">
          Saved preferences are used for job matching automatically.
        </p>
      </section>
    );
  }

  if (subscriptionState === "pending_payment") {
    return (
      <section className="bk-card rounded-2xl border-amber-200 bg-amber-50 p-5">
        <h3 className="text-base font-semibold text-amber-900">
          Payment is being processed
        </h3>
        <p className="mt-1 text-sm text-amber-800">
          Notifications will be enabled after successful payment.
        </p>
        <ButtonLink
          href="/pricing"
          variant="outline"
          size="sm"
          className="mt-3"
        >
          Continue payment
        </ButtonLink>
      </section>
    );
  }

  if (subscriptionState === "premium_expired") {
    return (
      <section className="bk-card rounded-2xl border-orange-200 bg-orange-50 p-5">
        <h3 className="text-base font-semibold text-orange-900">
          Premium has expired
        </h3>
        <p className="mt-1 text-sm text-orange-800">
          Renew your subscription to reactivate matching notifications.
        </p>
        <ButtonLink
          href="/pricing"
          variant="outline"
          size="sm"
          className="mt-3"
        >
          Upgrade again
        </ButtonLink>
      </section>
    );
  }

  if (subscriptionState === "status_unavailable") {
    return (
      <section className="bk-card rounded-2xl border-red-200 bg-red-50 p-5">
        <h3 className="text-base font-semibold text-red-900">
          Premium status unavailable
        </h3>
        <p className="mt-1 text-sm text-red-800">
          We couldn&apos;t fetch the latest billing status. Please refresh the
          page.
        </p>
      </section>
    );
  }

  return (
    <section className="bk-card-muted p-5">
      <h3 className="text-base font-semibold text-gray-900">
        Premium notifications
      </h3>
      <p className="mt-1 text-sm text-gray-700">
        Free users can still save preferences, but matching notifications are
        only active for premium users.
      </p>
      <ButtonLink href="/pricing" variant="outline" size="sm" className="mt-3">
        View premium plans
      </ButtonLink>
    </section>
  );
}
