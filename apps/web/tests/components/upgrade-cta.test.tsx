import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { APIRequestError } from "@/lib/utils/fetch-json";
import { UpgradeCTA } from "@/features/billing/components/upgrade-cta";
import { redirectToExternalURL } from "@/lib/utils/browser-navigation";
import { createSessionAPIClient } from "@/services/session-api-client";

const replaceMock = vi.fn();
const markAnonymousMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    replace: replaceMock,
  }),
}));

vi.mock("@/features/auth/session-provider", () => ({
  useAuthSession: () => ({
    state: "authenticated",
    markAuthenticated: vi.fn(),
    markAnonymous: markAnonymousMock,
  }),
}));

vi.mock("@/services/session-api-client", async () => {
  const actual = await vi.importActual<
    typeof import("@/services/session-api-client")
  >("@/services/session-api-client");
  return {
    ...actual,
    createSessionAPIClient: vi.fn(),
  };
});

vi.mock("@/lib/utils/browser-navigation", () => ({
  redirectToExternalURL: vi.fn(),
}));

describe("UpgradeCTA", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it("creates checkout and redirects to provider URL", async () => {
    const createCheckoutSessionMock = vi.fn().mockResolvedValue({
      meta: {
        code: 201,
        status: "success",
        message: "Checkout session created",
      },
      data: {
        provider: "mayar",
        invoice_id: "inv_1",
        transaction_id: "trx_1",
        checkout_url: "https://checkout.example.com/inv_1",
        expired_at: "2030-01-01T00:00:00Z",
        subscription_state: "pending_payment",
        transaction_status: "pending",
      },
    });

    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: createCheckoutSessionMock,
      getBillingTransactions: vi.fn(),
      getPreferences: vi.fn(),
      updatePreferences: vi.fn(),
      listSavedSearches: vi.fn(),
      createSavedSearch: vi.fn(),
      deleteSavedSearch: vi.fn(),
      listNotifications: vi.fn(),
      markNotificationAsRead: vi.fn(),
      updateNotificationPreferences: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(createCheckoutSessionMock).toHaveBeenCalledTimes(1);
    });
    expect(redirectToExternalURL).toHaveBeenCalledWith(
      "https://checkout.example.com/inv_1",
    );
  });

  it("renders retry message for provider errors", async () => {
    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(
            "Service unavailable",
            503,
            "SERVICE_UNAVAILABLE",
          ),
        ),
      getBillingTransactions: vi.fn(),
      getPreferences: vi.fn(),
      updatePreferences: vi.fn(),
      listSavedSearches: vi.fn(),
      createSavedSearch: vi.fn(),
      deleteSavedSearch: vi.fn(),
      listNotifications: vi.fn(),
      markNotificationAsRead: vi.fn(),
      updateNotificationPreferences: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "The payment provider is currently unavailable. Please try again shortly.",
        ),
      ).toBeInTheDocument();
    });
  });
});
