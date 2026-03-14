"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { useAuthSession } from "@/features/auth/session-provider";
import {
  clearCheckoutSession,
  loadCheckoutSession,
  saveCheckoutSession,
} from "@/features/billing/checkout-session-cache";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { redirectToExternalURL } from "@/lib/utils/browser-navigation";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";
import type { SubscriptionState } from "@/services/auth";
import type { TransactionStatus } from "@/services/billing";

interface UpgradeCTAProps {
  subscriptionState: SubscriptionState | "status_unavailable";
  lastTransactionStatus?: TransactionStatus;
}

export function UpgradeCTA({
  subscriptionState,
  lastTransactionStatus,
}: UpgradeCTAProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);
  const [cachedCheckout, setCachedCheckout] = useState(() =>
    loadCheckoutSession(),
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const isPremiumActive = subscriptionState === "premium_active";
  const hasPendingContinueLink =
    subscriptionState === "pending_payment" &&
    Boolean(cachedCheckout?.checkout_url);

  const buttonLabel = getButtonLabel(
    subscriptionState,
    hasPendingContinueLink,
    lastTransactionStatus,
  );

  async function handleUpgrade() {
    setMessage(null);

    if (hasPendingContinueLink && cachedCheckout) {
      redirectToExternalURL(cachedCheckout.checkout_url);
      return;
    }

    setIsSubmitting(true);
    try {
      const redirectURL = `${window.location.origin}/billing/success`;
      const response = await sessionClient.createCheckoutSession({
        plan_code: "pro_monthly",
        redirect_url: redirectURL,
        idempotency_key: createIdempotencyKey(),
      });

      saveCheckoutSession(response.data);
      setCachedCheckout(loadCheckoutSession());
      setMessage("Checkout created. Redirecting to the payment page...");
      redirectToExternalURL(response.data.checkout_url);
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          clearBrowserSession();
          markAnonymous();
          router.replace(buildLoginHref("/pricing"));
          return;
        }
        if (error.status === 409) {
          setMessage(
            "Your account is already premium active. Please check the latest status.",
          );
          clearCheckoutSession();
          setCachedCheckout(null);
          return;
        }
        if (error.status === 429) {
          setMessage("Too many checkout requests. Please try again shortly.");
          return;
        }
        if (error.status === 400) {
          setMessage(
            "Invalid checkout request. Please ensure the plan and redirect URL are correct.",
          );
          return;
        }
        if (error.status === 502 || error.status === 503) {
          setMessage(
            "The payment provider is currently unavailable. Please try again shortly.",
          );
          return;
        }
        setMessage(error.message);
        return;
      }

      setMessage("Failed to create checkout. Please try again shortly.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="bk-card grid gap-4 p-6 sm:p-8">
      <h3 className="text-[24px] font-normal text-black">Upgrade premium</h3>
      <p className="text-[14px] text-[#666666]">
        Start or continue checkout to unlock premium features and faster
        notifications.
      </p>
      <Button
        type="button"
        disabled={isSubmitting || isPremiumActive}
        onClick={() => void handleUpgrade()}
        variant="secondary"
      >
        {isSubmitting ? "Processing..." : buttonLabel}
      </Button>
      {message ? (
        <p
          className="rounded-2xl border border-[#E5E5E5] bg-[#F4F4F4] px-4 py-3 text-[14px] text-[#666666]"
          role="status"
          aria-live="polite"
        >
          {message}
        </p>
      ) : null}
    </section>
  );
}

function getButtonLabel(
  state: SubscriptionState | "status_unavailable",
  hasPendingContinueLink: boolean,
  lastTransactionStatus?: TransactionStatus,
): string {
  if (state === "premium_active") {
    return "Premium active";
  }
  if (state === "pending_payment") {
    return hasPendingContinueLink
      ? "Continue payment"
      : "Create a new checkout";
  }
  if (state === "premium_expired") {
    return "Upgrade again";
  }
  if (lastTransactionStatus === "failed") {
    return "Retry checkout";
  }
  if (state === "status_unavailable") {
    return "Try premium checkout";
  }
  return "Upgrade to Pro";
}

function createIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `checkout-${Date.now()}-${Math.round(Math.random() * 1_000_000)}`;
}
