import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

import type { SubscriptionState } from "@/services/auth";

/**
 * TransactionStatus defines the shape of transaction status.
 */
export type TransactionStatus = "pending" | "reminder" | "success" | "failed";

/**
 * BillingStatus defines the shape of billing status.
 */
export interface BillingStatus {
  plan_code?: string;
  subscription_state: SubscriptionState;
  is_premium: boolean;
  premium_expired_at?: string | null;
  last_transaction_status?: TransactionStatus;
}

/**
 * CreateCheckoutSessionInput defines the shape of create checkout session input.
 */
export interface CreateCheckoutSessionInput {
  plan_code: "pro_monthly";
  coupon_code?: string;
  customer_mobile: string;
  redirect_url: string;
  idempotency_key?: string;
}

/**
 * CheckoutSession defines the shape of checkout session.
 */
export interface CheckoutSession {
  provider: "midtrans";
  plan_code?: "pro_monthly";
  invoice_id: string;
  transaction_id: string;
  checkout_url: string;
  snap_token?: string;
  original_amount?: number;
  discount_amount?: number;
  final_amount?: number;
  coupon_code?: string;
  expired_at: string;
  subscription_state: SubscriptionState;
  transaction_status: TransactionStatus;
}

/**
 * BillingTransaction defines the shape of billing transaction.
 */
export interface BillingTransaction {
  id: string;
  provider: string;
  provider_transaction_id: string;
  amount: number;
  status: TransactionStatus;
  created_at: string;
}

/**
 * BillingTransactionsQuery defines the shape of billing transactions query.
 */
export interface BillingTransactionsQuery {
  page?: number;
  limit?: number;
  status?: TransactionStatus;
}

/**
 * getBillingStatus returns billing status.
 */
export async function getBillingStatus(
  accessToken: string,
  init?: RequestInit,
): Promise<APIResponse<BillingStatus>> {
  return fetchJSON<BillingStatus>(buildAPIURL("/billing/status"), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * createCheckoutSession creates checkout session.
 */
export async function createCheckoutSession(
  accessToken: string,
  input: CreateCheckoutSessionInput,
  init?: RequestInit,
): Promise<APIResponse<CheckoutSession>> {
  const { idempotency_key, ...payload } = input;
  return fetchJSON<CheckoutSession>(buildAPIURL("/billing/checkout-session"), {
    method: "POST",
    body: JSON.stringify(payload),
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(idempotency_key ? { "Idempotency-Key": idempotency_key } : {}),
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * getBillingTransactions returns billing transactions.
 */
export async function getBillingTransactions(
  accessToken: string,
  query: BillingTransactionsQuery = {},
  init?: RequestInit,
): Promise<APIResponse<BillingTransaction[]>> {
  const params = new URLSearchParams();
  if (query.page !== undefined) {
    params.set("page", String(query.page));
  }
  if (query.limit !== undefined) {
    params.set("limit", String(query.limit));
  }
  if (query.status) {
    params.set("status", query.status);
  }

  const querySuffix = params.toString();
  const endpoint = querySuffix
    ? `/billing/transactions?${querySuffix}`
    : "/billing/transactions";

  return fetchJSON<BillingTransaction[]>(buildAPIURL(endpoint), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}
