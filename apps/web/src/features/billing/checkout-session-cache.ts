import type { CheckoutSession } from "@/services/billing";

const checkoutSessionStorageKey = "bisakerja:checkout-session";

export interface CachedCheckoutSession {
  checkout_url: string;
  expired_at: string;
  transaction_id: string;
  plan_code?: "pro_monthly";
  original_amount?: number;
  discount_amount?: number;
  final_amount?: number;
  coupon_code?: string;
}

export function saveCheckoutSession(session: CheckoutSession): void {
  if (typeof window === "undefined") {
    return;
  }
  const payload: CachedCheckoutSession = {
    checkout_url: session.checkout_url,
    expired_at: session.expired_at,
    transaction_id: session.transaction_id,
    plan_code: session.plan_code,
    original_amount: session.original_amount,
    discount_amount: session.discount_amount,
    final_amount: session.final_amount,
    coupon_code: session.coupon_code,
  };
  window.localStorage.setItem(
    checkoutSessionStorageKey,
    JSON.stringify(payload),
  );
}

export function loadCheckoutSession(): CachedCheckoutSession | null {
  if (typeof window === "undefined") {
    return null;
  }
  const raw = window.localStorage.getItem(checkoutSessionStorageKey);
  if (!raw) {
    return null;
  }
  try {
    const parsed = JSON.parse(raw) as CachedCheckoutSession;
    if (
      !parsed ||
      typeof parsed.checkout_url !== "string" ||
      typeof parsed.expired_at !== "string" ||
      typeof parsed.transaction_id !== "string"
    ) {
      return null;
    }
    if (isCheckoutExpired(parsed.expired_at)) {
      clearCheckoutSession();
      return null;
    }

    if (parsed.plan_code !== undefined && parsed.plan_code !== "pro_monthly") {
      return null;
    }
    if (!isOptionalAmount(parsed.original_amount)) {
      return null;
    }
    if (!isOptionalAmount(parsed.discount_amount)) {
      return null;
    }
    if (!isOptionalAmount(parsed.final_amount)) {
      return null;
    }
    if (
      parsed.coupon_code !== undefined &&
      typeof parsed.coupon_code !== "string"
    ) {
      return null;
    }

    return parsed;
  } catch {
    return null;
  }
}

export function clearCheckoutSession(): void {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.removeItem(checkoutSessionStorageKey);
}

export function isCheckoutExpired(expiredAt: string): boolean {
  const value = Date.parse(expiredAt);
  if (Number.isNaN(value)) {
    return true;
  }
  return Date.now() >= value;
}

function isOptionalAmount(value: unknown): value is number | undefined {
  if (value === undefined) {
    return true;
  }
  return typeof value === "number" && Number.isFinite(value) && value >= 0;
}
