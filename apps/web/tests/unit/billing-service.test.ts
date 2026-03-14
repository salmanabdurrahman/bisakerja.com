import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createCheckoutSession,
  getBillingTransactions,
} from "@/services/billing";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

describe("billing services", () => {
  it("sends idempotency key when creating checkout session", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: { code: 201, status: "success", message: "Checkout created" },
          data: {
            provider: "mayar",
            invoice_id: "inv_1",
            transaction_id: "trx_1",
            checkout_url: "https://checkout.example.com",
            expired_at: "2030-01-01T00:00:00Z",
            subscription_state: "pending_payment",
            transaction_status: "pending",
          },
        }),
        { status: 201 },
      ),
    );

    vi.stubGlobal("fetch", fetchMock);

    await createCheckoutSession("access-token", {
      plan_code: "pro_monthly",
      redirect_url: "https://app.bisakerja.com/billing/success",
      idempotency_key: "idem-123",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/billing/checkout-session",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
          "Idempotency-Key": "idem-123",
        }),
      }),
    );
  });

  it("builds query params for billing transactions endpoint", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 200,
            status: "success",
            message: "Transactions retrieved",
          },
          data: [],
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await getBillingTransactions("access-token", {
      page: 2,
      limit: 50,
      status: "failed",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/billing/transactions?page=2&limit=50&status=failed",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });
});
