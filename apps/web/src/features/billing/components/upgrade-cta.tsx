"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { useAuthSession } from "@/features/auth/session-provider";
import {
  type CachedCheckoutSession,
  clearCheckoutSession,
  loadCheckoutSession,
  saveCheckoutSession,
} from "@/features/billing/checkout-session-cache";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";
import type { SubscriptionState } from "@/services/auth";
import type { TransactionStatus } from "@/services/billing";

declare global {
  interface Window {
    snap: {
      pay: (
        token: string,
        options: {
          onSuccess?: () => void;
          onPending?: () => void;
          onError?: () => void;
          onClose?: () => void;
        },
      ) => void;
    };
  }
}

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
  const [customerMobileInput, setCustomerMobileInput] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const isPremiumActive = subscriptionState === "premium_active";
  const hasPendingContinueLink =
    subscriptionState === "pending_payment" &&
    Boolean(cachedCheckout?.snap_token);
  const checkoutAmountSummary = cachedCheckout
    ? getCheckoutAmountSummary(cachedCheckout)
    : null;

  const buttonLabel = getButtonLabel(
    subscriptionState,
    hasPendingContinueLink,
    lastTransactionStatus,
  );

  function openSnapPay(snapToken: string) {
    if (!window.snap) {
      setMessage(
        "Payment popup is not ready. Please refresh the page and try again.",
      );
      return;
    }
    window.snap.pay(snapToken, {
      onSuccess: () => {
        clearCheckoutSession();
        router.push("/billing/success");
      },
      onPending: () => {
        setCachedCheckout(loadCheckoutSession());
        setMessage("Payment pending. You can continue payment later.");
      },
      onError: () => {
        setMessage("Payment failed. Please try again.");
      },
      onClose: () => {
        setMessage("Payment popup closed. You can continue payment later.");
      },
    });
  }

  async function handleUpgrade() {
    setMessage(null);

    if (hasPendingContinueLink && cachedCheckout?.snap_token) {
      openSnapPay(cachedCheckout.snap_token);
      return;
    }

    const normalizedCustomerMobile =
      normalizeCustomerMobile(customerMobileInput);
    if (!isCustomerMobileValid(normalizedCustomerMobile)) {
      setMessage(
        "Phone number is required and must contain 9-15 digits (numbers only).",
      );
      return;
    }

    setIsSubmitting(true);
    try {
      const redirectURL = `${window.location.origin}/billing/success`;
      const response = await sessionClient.createCheckoutSession({
        plan_code: "pro_monthly",
        customer_mobile: normalizedCustomerMobile,
        redirect_url: redirectURL,
        idempotency_key: createIdempotencyKey(),
      });

      saveCheckoutSession(response.data);
      setCachedCheckout(loadCheckoutSession());
      setMessage("Checkout created. Opening payment popup...");

      if (response.data.snap_token) {
        openSnapPay(response.data.snap_token);
      }
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
          setMessage(
            "Too many checkout requests. Please wait around 10 seconds, then retry or continue your latest pending checkout.",
          );
          return;
        }
        if (error.status === 400) {
          setMessage(getCheckoutValidationErrorMessage(error.code));
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
      <h3 className="bk-heading-card">Upgrade premium</h3>
      <p className="bk-body">
        Start or continue checkout to unlock premium features and faster
        notifications.
      </p>
      {!isPremiumActive ? (
        <div className="grid gap-2">
          <label
            htmlFor="checkout-customer-mobile"
            className="bk-body-sm font-medium uppercase tracking-wider text-[#666666]"
          >
            Phone number (required)
          </label>
          <input
            id="checkout-customer-mobile"
            name="customer_mobile"
            value={customerMobileInput}
            onChange={(event) => setCustomerMobileInput(event.target.value)}
            placeholder="e.g. 08123456789"
            autoComplete="tel-national"
            className="h-11 rounded-2xl border border-[#E5E5E5] bg-white px-4 text-[14px] text-black placeholder:text-[#999999] focus:border-black focus:outline-none"
            disabled={isSubmitting || hasPendingContinueLink}
          />
          <p className="bk-body-sm text-[#888888]">
            Used for Midtrans customer detail during checkout.
          </p>
        </div>
      ) : null}
      {checkoutAmountSummary ? (
        <div className="grid gap-1 rounded-2xl border border-[#E5E5E5] bg-[#F9F9F9] px-4 py-3 bk-body-sm text-[#555555]">
          <p className="bk-body-sm font-medium uppercase tracking-wider text-[#777777]">
            Latest checkout summary
          </p>
          <p>
            Original amount: {formatIDRCurrency(checkoutAmountSummary.original)}
          </p>
          {checkoutAmountSummary.discount > 0 ? (
            <p>
              Discount: -{formatIDRCurrency(checkoutAmountSummary.discount)}
            </p>
          ) : null}
          <p className="font-medium text-black">
            Final amount: {formatIDRCurrency(checkoutAmountSummary.final)}
          </p>
        </div>
      ) : null}
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
          className="rounded-2xl border border-[#E5E5E5] bg-[#F4F4F4] px-4 py-3 bk-body"
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

function normalizeCustomerMobile(raw: string): string {
  return raw.replace(/[^\d]/g, "");
}

function isCustomerMobileValid(raw: string): boolean {
  return /^\d{9,15}$/.test(raw);
}

interface CheckoutAmountSummary {
  original: number;
  discount: number;
  final: number;
}

function getCheckoutAmountSummary(
  checkout: CachedCheckoutSession,
): CheckoutAmountSummary | null {
  const finalAmount = parseAmount(checkout.final_amount);
  if (finalAmount === null || finalAmount <= 0) {
    return null;
  }

  const originalAmount = parseAmount(checkout.original_amount) ?? finalAmount;
  const calculatedDiscount = Math.max(0, originalAmount - finalAmount);
  const discountAmount =
    parseAmount(checkout.discount_amount) ?? calculatedDiscount;

  return {
    original: originalAmount,
    discount: Math.max(0, discountAmount),
    final: finalAmount,
  };
}

function parseAmount(value: number | undefined): number | null {
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0) {
    return null;
  }
  return value;
}

function formatIDRCurrency(value: number): string {
  return `IDR ${value.toLocaleString("en-US")}`;
}

function getCheckoutValidationErrorMessage(
  code: string | null | undefined,
): string {
  switch (code) {
    case "INVALID_REDIRECT_URL":
      return "Redirect URL is invalid. Use an allowlisted host and https (http is only allowed for localhost in local development).";
    case "INVALID_PLAN_CODE":
      return "Selected plan is unavailable. Please refresh and try again.";
    case "INVALID_CUSTOMER_MOBILE":
      return "Phone number is invalid. Use 9-15 digits and try again.";
    default:
      return "Invalid checkout request. Please ensure the plan and redirect URL are correct.";
  }
}
