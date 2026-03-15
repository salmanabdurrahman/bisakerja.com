import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { APIRequestError } from "@/lib/utils/fetch-json";
import { UpgradeCTA } from "@/features/billing/components/upgrade-cta";

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

describe("UpgradeCTA", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  function fillRequiredPhoneNumber(): void {
    fireEvent.change(screen.getByLabelText("Phone number (required)"), {
      target: { value: "08123456789" },
    });
  }

  it("creates checkout and opens snap payment popup", async () => {
    const snapPayMock = vi.fn();
    window.snap = { pay: snapPayMock };

    const createCheckoutSessionMock = vi.fn().mockResolvedValue({
      meta: {
        code: 201,
        status: "success",
        message: "Checkout session created",
      },
      data: {
        provider: "midtrans",
        plan_code: "pro_monthly",
        invoice_id: "inv_1",
        transaction_id: "pay-abc123",
        checkout_url:
          "https://app.sandbox.midtrans.com/snap/v4/redirection/tok_1",
        snap_token: "tok_1",
        original_amount: 49000,
        discount_amount: 0,
        final_amount: 49000,
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fillRequiredPhoneNumber();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(createCheckoutSessionMock).toHaveBeenCalledTimes(1);
    });
    expect(createCheckoutSessionMock).toHaveBeenCalledWith(
      expect.objectContaining({
        customer_mobile: "08123456789",
        plan_code: "pro_monthly",
      }),
    );
    await waitFor(() => {
      expect(snapPayMock).toHaveBeenCalledWith(
        "tok_1",
        expect.objectContaining({
          onSuccess: expect.any(Function),
          onPending: expect.any(Function),
          onError: expect.any(Function),
          onClose: expect.any(Function),
        }),
      );
    });
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fillRequiredPhoneNumber();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "The payment provider is currently unavailable. Please try again shortly.",
        ),
      ).toBeInTheDocument();
    });
  });

  it("renders retry guidance for checkout rate limit errors", async () => {
    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError("Too many requests", 429, "TOO_MANY_REQUESTS"),
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fillRequiredPhoneNumber();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Too many checkout requests. Please wait around 10 seconds, then retry or continue your latest pending checkout.",
        ),
      ).toBeInTheDocument();
    });
  });

  it("renders generic validation error for unrecognized validation errors", async () => {
    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(
            "Validation error",
            400,
            "UNKNOWN_VALIDATION_CODE",
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fillRequiredPhoneNumber();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Invalid checkout request. Please ensure the plan and redirect URL are correct.",
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fillRequiredPhoneNumber();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Redirect URL is invalid. Use an allowlisted host and https (http is only allowed for localhost in local development).",
        ),
      ).toBeInTheDocument();
    });
  });

  it("renders customer-mobile validation message from API", async () => {
    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi
        .fn()
        .mockRejectedValue(
          new APIRequestError(
            "Validation error",
            400,
            "INVALID_CUSTOMER_MOBILE",
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fillRequiredPhoneNumber();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Phone number is invalid. Use 9-15 digits and try again.",
        ),
      ).toBeInTheDocument();
    });
  });

  it("blocks checkout when phone number is missing", async () => {
    const createCheckoutSessionMock = vi.fn();
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
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: vi.fn(),
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: vi.fn(),
      listTrackedApplications: vi.fn(),
    });

    render(<UpgradeCTA subscriptionState="free" />);
    fireEvent.click(screen.getByRole("button", { name: "Upgrade to Pro" }));

    await waitFor(() => {
      expect(
        screen.getByText(
          "Phone number is required and must contain 9-15 digits (numbers only).",
        ),
      ).toBeInTheDocument();
    });
    expect(createCheckoutSessionMock).not.toHaveBeenCalled();
  });
});
