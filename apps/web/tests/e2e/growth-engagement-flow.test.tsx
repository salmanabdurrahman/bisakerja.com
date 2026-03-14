import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountNotificationsClient } from "@/features/growth/components/account-notifications-client";
import { AccountSavedSearchesClient } from "@/features/growth/components/account-saved-searches-client";
import { NotificationDigestControl } from "@/features/preferences/components/notification-digest-control";
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

describe("Growth engagement flow (e2e-like)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
  });

  it("completes saved-search, notifications, and digest updates in one journey", async () => {
    const createSavedSearchMock = vi.fn().mockResolvedValue({
      meta: { code: 201, status: "success", message: "Saved search created" },
      data: {
        id: "ss_1",
        query: "golang backend",
        location: "jakarta",
        source: "glints",
        salary_min: 12000000,
        frequency: "daily_digest",
        is_active: true,
        created_at: "2030-01-01T00:00:00Z",
        updated_at: "2030-01-01T00:00:00Z",
      },
    });

    const deleteSavedSearchMock = vi.fn().mockResolvedValue({
      meta: { code: 200, status: "success", message: "Saved search deleted" },
      data: { id: "ss_1" },
    });

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
      createSavedSearch: createSavedSearchMock,
      deleteSavedSearch: deleteSavedSearchMock,
      listNotifications: vi.fn(),
      markNotificationAsRead: markReadMock,
      updateNotificationPreferences: vi.fn(),
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
    });

    const savedSearchesView = render(
      <AccountSavedSearchesClient initialSavedSearches={[]} />,
    );

    fireEvent.change(screen.getByLabelText("Query"), {
      target: { value: "golang backend" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add saved search" }));

    await waitFor(() => {
      expect(createSavedSearchMock).toHaveBeenCalledTimes(1);
    });
    expect(screen.getByText("golang backend")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Delete" }));
    await waitFor(() => {
      expect(deleteSavedSearchMock).toHaveBeenCalledWith("ss_1");
    });

    savedSearchesView.unmount();

    const notificationsView = render(
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
    await waitFor(() => {
      expect(screen.getByText(/Read at:/i)).toBeInTheDocument();
    });

    notificationsView.unmount();

    const submitDigestMock = vi.fn().mockResolvedValue({
      user_id: "user_1",
      alert_mode: "daily_digest",
      digest_hour: 8,
      updated_at: "2030-01-01T00:00:00Z",
    });

    render(
      <NotificationDigestControl
        initialSettings={{
          alert_mode: "instant",
          digest_hour: null,
          updated_at: null,
        }}
        onSubmit={submitDigestMock}
        onUnauthorized={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("Alert mode"), {
      target: { value: "daily_digest" },
    });
    fireEvent.change(screen.getByRole("spinbutton"), {
      target: { value: "8" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Save notification settings" }),
    );

    await waitFor(() => {
      expect(submitDigestMock).toHaveBeenCalledWith({
        alert_mode: "daily_digest",
        digest_hour: 8,
      });
    });
  });
});
