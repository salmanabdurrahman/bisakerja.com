import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

import type { SubscriptionState } from "@/services/auth";

export type TransactionStatus = "pending" | "reminder" | "success" | "failed";

export interface BillingStatus {
  plan_code?: string;
  subscription_state: SubscriptionState;
  is_premium: boolean;
  premium_expired_at?: string | null;
  last_transaction_status?: TransactionStatus;
}

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
