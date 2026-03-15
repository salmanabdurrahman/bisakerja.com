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
        plan_code: "pro_monthly",
        invoice_id: "inv_1",
        transaction_id: "trx_1",
        checkout_url: "https://checkout.example.com/inv_1",
        original_amount: 49000,
        discount_amount: 10000,
        final_amount: 39000,
        coupon_code: "SAVE10",
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
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fireEvent.change(screen.getByLabelText("Coupon code (optional)"), {
      target: { value: "save10" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(createCheckoutSessionMock).toHaveBeenCalledTimes(1);
    });
    expect(createCheckoutSessionMock).toHaveBeenCalledWith(
      expect.objectContaining({
        coupon_code: "SAVE10",
      }),
    );
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
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
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

  it("renders invalid coupon message for coupon validation errors", async () => {
    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError("Validation error", 400, "INVALID_COUPON_CODE"),
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
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fireEvent.change(screen.getByLabelText("Coupon code (optional)"), {
      target: { value: "BADCODE" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Coupon code is invalid or unavailable. Please try another code.",
        ),
      ).toBeInTheDocument();
    });
  });

  it("renders actionable redirect-url message for redirect validation errors", async () => {
    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError("Validation error", 400, "INVALID_REDIRECT_URL"),
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
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Redirect URL is invalid. Use an allowlisted host and https (http is only allowed for localhost in local development).",
        ),
      ).toBeInTheDocument();
    });
  });
});
