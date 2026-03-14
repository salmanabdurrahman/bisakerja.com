import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountNotificationsClient } from "@/features/growth/components/account-notifications-client";
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

describe("AccountNotificationsClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it("marks unread notification as read", async () => {
    const markReadMock = vi.fn().mockResolvedValue({
      meta: {
        code: 200,
        status: "success",
        message: "Notification marked as read",
      },
      data: {
        id: "notif_1",
        job_id: "job_1",
        channel: "email",
        status: "sent",
        sent_at: "2030-01-01T00:00:00Z",
        read_at: "2030-01-01T01:00:00Z",
        created_at: "2030-01-01T00:00:00Z",
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
      markNotificationAsRead: markReadMock,
      updateNotificationPreferences: vi.fn(),
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
    });

    render(
      <AccountNotificationsClient
        initialNotifications={[
          {
            id: "notif_1",
            job_id: "job_1",
            channel: "email",
            status: "sent",
            sent_at: "2030-01-01T00:00:00Z",
            read_at: null,
            created_at: "2030-01-01T00:00:00Z",
          },
        ]}
        initialPagination={{
          page: 1,
          limit: 20,
          total_pages: 1,
          total_records: 1,
        }}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Mark as read" }));

    await waitFor(() => {
      expect(markReadMock).toHaveBeenCalledWith("notif_1");
    });
    expect(screen.getByText(/Read at:/i)).toBeInTheDocument();
  });
});
