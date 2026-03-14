import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import BillingSuccessPage from "@/app/billing/success/page";
import { getBillingStatus } from "@/services/billing";
import { resolveServerAccessToken } from "@/lib/auth/server-session";

vi.mock("@/services/billing", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/billing")>(
      "@/services/billing",
    );
  return {
    ...actual,
    getBillingStatus: vi.fn(),
  };
});

vi.mock("@/lib/auth/server-session", () => ({
  resolveServerAccessToken: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn((href: string) => {
    throw new Error(`REDIRECT:${href}`);
  }),
}));

describe("Billing success page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(resolveServerAccessToken).mockResolvedValue("access-token");
  });

  it("renders upgrade success state when subscription is premium_active", async () => {
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

    const page = await BillingSuccessPage();
    render(page);

    expect(
      screen.getByText("Pembayaran terverifikasi. Premium kamu sudah aktif."),
    ).toBeInTheDocument();
  });

  it("renders pending state when subscription remains pending_payment", async () => {
    vi.mocked(getBillingStatus).mockResolvedValueOnce({
      meta: {
        code: 200,
        status: "success",
        message: "Billing status retrieved",
      },
      data: {
        subscription_state: "pending_payment",
        is_premium: false,
        last_transaction_status: "pending",
      },
    });

    const page = await BillingSuccessPage();
    render(page);

    expect(
      screen.getByText(
        "Pembayaran masih diproses. Cek ulang status dalam beberapa saat.",
      ),
    ).toBeInTheDocument();
  });

  it("renders re-offer state for failed transaction", async () => {
    vi.mocked(getBillingStatus).mockResolvedValueOnce({
      meta: {
        code: 200,
        status: "success",
        message: "Billing status retrieved",
      },
      data: {
        subscription_state: "free",
        is_premium: false,
        last_transaction_status: "failed",
      },
    });

    const page = await BillingSuccessPage();
    render(page);

    expect(
      screen.getByText(
        "Pembayaran belum berhasil. Kamu bisa memulai checkout baru.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Kembali ke pricing" }),
    ).toHaveAttribute("href", "/pricing");
  });
});
