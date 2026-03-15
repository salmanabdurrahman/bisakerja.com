import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import AccountAIToolsPage from "@/app/account/ai-tools/page";
import { useAuthSession } from "@/features/auth/session-provider";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { getMe } from "@/services/auth";
import { getBillingStatus } from "@/services/billing";
import { createSessionAPIClient } from "@/services/session-api-client";

vi.mock("@/services/auth", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/auth")>("@/services/auth");
  return {
    ...actual,
    getMe: vi.fn(),
  };
});

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

vi.mock("@/services/session-api-client", async () => {
  const actual = await vi.importActual<
    typeof import("@/services/session-api-client")
  >("@/services/session-api-client");
  return {
    ...actual,
    createSessionAPIClient: vi.fn(),
  };
});

vi.mock("@/lib/auth/server-session", () => ({
  resolveServerAccessToken: vi.fn(),
}));

vi.mock("@/features/auth/session-provider", () => ({
  useAuthSession: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn((href: string) => {
    throw new Error(`REDIRECT:${href}`);
  }),
  useRouter: () => ({
    replace: vi.fn(),
  }),
}));

describe("Account AI tools page", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(resolveServerAccessToken).mockResolvedValue("access-token");
    vi.mocked(useAuthSession).mockReturnValue({
      state: "authenticated",
      markAuthenticated: vi.fn(),
      markAnonymous: vi.fn(),
    });
  });

  it("renders AI tools page and loads usage cards", async () => {
    vi.mocked(getMe).mockResolvedValueOnce({
      meta: { code: 200, status: "success", message: "Profile retrieved" },
      data: {
        id: "user_1",
        email: "user@example.com",
        name: "User",
        role: "user",
        is_premium: true,
        subscription_state: "premium_active",
      },
    });
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

    const getAIUsageMock = vi.fn().mockResolvedValue({
      meta: { code: 200, status: "success", message: "AI usage retrieved" },
      data: {
        feature: "search_assistant",
        tier: "premium",
        daily_quota: 30,
        used: 2,
        remaining: 28,
        reset_at: "2030-01-01T00:00:00Z",
      },
    });

    vi.mocked(createSessionAPIClient).mockReturnValue({
      getMe: vi.fn(),
      getBillingStatus: vi.fn(),
      createCheckoutSession: vi.fn(),
      getBillingTransactions: vi.fn(),
      getPreferences: vi.fn(),
      updatePreferences: vi.fn(),
      listSavedSearches: vi.fn(),
      createSavedSearch: vi.fn(),
      deleteSavedSearch: vi.fn(),
      listNotifications: vi.fn(),
      markNotificationAsRead: vi.fn(),
      updateNotificationPreferences: vi.fn(),
      getAIUsage: getAIUsageMock,
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

    const page = await AccountAIToolsPage();
    render(page);

    expect(
      screen.getByRole("heading", { name: "AI tools", level: 2 }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "AI usage today", level: 3 }),
    ).toBeInTheDocument();

    await waitFor(() => {
      expect(getAIUsageMock).toHaveBeenCalledTimes(3);
    });
  });
});
