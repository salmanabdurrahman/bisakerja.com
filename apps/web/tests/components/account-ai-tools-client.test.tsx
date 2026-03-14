import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountAIToolsClient } from "@/features/ai/components/account-ai-tools-client";
import { APIRequestError } from "@/lib/utils/fetch-json";
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

describe("AccountAIToolsClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it("loads usage cards and generates search assistant suggestion", async () => {
    const generateAISearchAssistantMock = vi.fn().mockResolvedValue({
      meta: { code: 200, status: "success", message: "AI generated" },
      data: {
        feature: "search_assistant",
        prompt: "remote golang backend jobs",
        suggested_query: "golang backend remote",
        suggested_filters: {
          locations: ["Jakarta", "Remote"],
          job_types: ["fulltime"],
          salary_min: 15000000,
        },
        summary: "Start with backend remote roles and expand location scope.",
        tier: "free",
        provider: "openai_compatible",
        model: "gpt-4.1-mini",
        daily_quota: 5,
        used_today: 1,
        quota_remaining: 4,
        reset_at: "2030-01-01T00:00:00Z",
      },
    });

    const getAIUsageMock = vi
      .fn()
      .mockResolvedValueOnce({
        meta: { code: 200, status: "success", message: "Usage retrieved" },
        data: {
          feature: "search_assistant",
          tier: "free",
          daily_quota: 5,
          used: 0,
          remaining: 5,
          reset_at: "2030-01-01T00:00:00Z",
        },
      })
      .mockResolvedValueOnce({
        meta: { code: 200, status: "success", message: "Usage retrieved" },
        data: {
          feature: "job_fit_summary",
          tier: "free",
          daily_quota: 5,
          used: 0,
          remaining: 5,
          reset_at: "2030-01-01T00:00:00Z",
        },
      })
      .mockResolvedValueOnce({
        meta: { code: 200, status: "success", message: "Usage retrieved" },
        data: {
          feature: "cover_letter_draft",
          tier: "free",
          daily_quota: 5,
          used: 0,
          remaining: 5,
          reset_at: "2030-01-01T00:00:00Z",
        },
      });

    vi.mocked(createSessionAPIClient).mockReturnValue(
      createSessionClientMock({
        getAIUsage: getAIUsageMock,
        generateAISearchAssistant: generateAISearchAssistantMock,
      }),
    );

    render(
      <AccountAIToolsClient subscriptionState="free" infoMessage={null} />,
    );

    await waitFor(() => {
      expect(getAIUsageMock).toHaveBeenCalledTimes(3);
    });

    fireEvent.change(screen.getByLabelText("Prompt"), {
      target: { value: "remote golang backend jobs" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Generate suggestion" }),
    );

    await waitFor(() => {
      expect(generateAISearchAssistantMock).toHaveBeenCalledTimes(1);
    });
    expect(screen.getByText("golang backend remote")).toBeInTheDocument();
  });

  it("shows premium-only message when job fit is forbidden", async () => {
    const getAIUsageMock = vi
      .fn()
      .mockResolvedValueOnce({
        meta: { code: 200, status: "success", message: "Usage retrieved" },
        data: {
          feature: "search_assistant",
          tier: "free",
          daily_quota: 5,
          used: 0,
          remaining: 5,
          reset_at: "2030-01-01T00:00:00Z",
        },
      })
      .mockResolvedValueOnce({
        meta: { code: 200, status: "success", message: "Usage retrieved" },
        data: {
          feature: "job_fit_summary",
          tier: "free",
          daily_quota: 5,
          used: 0,
          remaining: 5,
          reset_at: "2030-01-01T00:00:00Z",
        },
      })
      .mockResolvedValueOnce({
        meta: { code: 200, status: "success", message: "Usage retrieved" },
        data: {
          feature: "cover_letter_draft",
          tier: "free",
          daily_quota: 5,
          used: 0,
          remaining: 5,
          reset_at: "2030-01-01T00:00:00Z",
        },
      });

    const generateAIJobFitSummaryMock = vi
      .fn()
      .mockRejectedValue(new APIRequestError("Forbidden", 403, "FORBIDDEN"));

    vi.mocked(createSessionAPIClient).mockReturnValue(
      createSessionClientMock({
        getAIUsage: getAIUsageMock,
        generateAIJobFitSummary: generateAIJobFitSummaryMock,
      }),
    );

    render(
      <AccountAIToolsClient subscriptionState="free" infoMessage={null} />,
    );

    await waitFor(() => {
      expect(getAIUsageMock).toHaveBeenCalledTimes(3);
    });

    fireEvent.change(screen.getAllByLabelText("Job ID")[0], {
      target: { value: "job_1" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Generate job fit" }));

    await waitFor(() => {
      expect(generateAIJobFitSummaryMock).toHaveBeenCalledTimes(1);
    });
    expect(
      screen.getByText("This AI action is available for premium users only."),
    ).toBeInTheDocument();
  });
});

function createSessionClientMock(
  overrides: Partial<ReturnType<typeof createSessionAPIClient>> = {},
): ReturnType<typeof createSessionAPIClient> {
  return {
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
    getAIUsage: vi.fn(),
    generateAISearchAssistant: vi.fn(),
    generateAIJobFitSummary: vi.fn(),
    generateAICoverLetterDraft: vi.fn(),
    ...overrides,
  };
}
