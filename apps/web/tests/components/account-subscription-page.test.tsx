import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import AccountSubscriptionPage from "@/app/account/subscription/page";
import { getBillingStatus, getBillingTransactions } from "@/services/billing";
import { resolveServerAccessToken } from "@/lib/auth/server-session";

vi.mock("@/services/billing", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/billing")>(
      "@/services/billing",
    );
  return {
    ...actual,
    getBillingStatus: vi.fn(),
    getBillingTransactions: vi.fn(),
  };
});

vi.mock("@/lib/auth/server-session", () => ({
  resolveServerAccessToken: vi.fn(),
}));

vi.mock("@/features/auth/session-provider", () => ({
  useAuthSession: () => ({
    state: "authenticated",
    markAuthenticated: vi.fn(),
    markAnonymous: vi.fn(),
  }),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn((href: string) => {
    throw new Error(`REDIRECT:${href}`);
  }),
  useRouter: () => ({
    replace: vi.fn(),
  }),
}));

describe("Account subscription page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(resolveServerAccessToken).mockResolvedValue("access-token");
  });

  it("renders status and billing transactions", async () => {
    vi.mocked(getBillingStatus).mockResolvedValueOnce({
      meta: {
        code: 200,
        status: "success",
        message: "Billing status retrieved",
      },
      data: {
        subscription_state: "premium_active",
        is_premium: true,
        last_transaction_status: "success",
        premium_expired_at: "2030-01-01T00:00:00Z",
      },
    });

    vi.mocked(getBillingTransactions).mockResolvedValueOnce({
      meta: { code: 200, status: "success", message: "Transactions retrieved" },
      data: [
        {
          id: "trx_1",
          provider: "mayar",
          provider_transaction_id: "mayar_trx_1",
          amount: 49000,
          status: "success",
          created_at: "2026-03-14T00:00:00Z",
        },
      ],
    });

    const page = await AccountSubscriptionPage();
    render(page);

    const nav = screen.getByRole("navigation", {
      name: "Account dashboard navigation",
    });
    expect(
      within(nav).getByRole("link", { name: "Manage preferences" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Subscription" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Current plan:")).toBeInTheDocument();
    expect(screen.getAllByText(/premium active/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/success - IDR 49,000/i)).toBeInTheDocument();
  });
});
