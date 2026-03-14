import type { CheckoutSession } from "@/services/billing";

const checkoutSessionStorageKey = "bisakerja:checkout-session";

export interface CachedCheckoutSession {
  checkout_url: string;
  expired_at: string;
  transaction_id: string;
}

export function saveCheckoutSession(session: CheckoutSession): void {
  if (typeof window === "undefined") {
    return;
  }
  const payload: CachedCheckoutSession = {
    checkout_url: session.checkout_url,
    expired_at: session.expired_at,
    transaction_id: session.transaction_id,
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
