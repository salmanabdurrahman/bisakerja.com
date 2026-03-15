import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountTrackerClient } from "@/features/tracker/components/account-tracker-client";
import { createSessionAPIClient } from "@/services/session-api-client";
import * as jobsService from "@/services/jobs";

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

vi.mock("@/services/jobs", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/jobs")>("@/services/jobs");
  return {
    ...actual,
    listJobs: vi.fn(),
    getJobDetail: vi.fn(),
  };
});

describe("AccountTrackerClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it("tracks and deletes application", async () => {
    const createTrackedApplicationMock = vi.fn().mockResolvedValue({
      meta: { code: 201, status: "success", message: "Application tracked" },
      data: {
        id: "app_1",
        job_id: "job_abc",
        status: "applied",
        notes: "",
        created_at: "2030-01-01T00:00:00Z",
        updated_at: "2030-01-01T00:00:00Z",
      },
    });

    const deleteTrackedApplicationMock = vi.fn().mockResolvedValue({
      meta: { code: 200, status: "success", message: "Application deleted" },
      data: { id: "app_1" },
    });

    vi.mocked(jobsService.listJobs).mockResolvedValue({
      meta: { code: 200, status: "success", message: "OK" },
      data: [
        {
          id: "job_abc",
          title: "Software Engineer",
          company: "Acme Corp",
          location: "Remote",
          salary_range: "100000-120000",
          source: "jobstreet",
        },
      ],
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
      getAIUsage: vi.fn(),
      generateAISearchAssistant: vi.fn(),
      generateAIJobFitSummary: vi.fn(),
      generateAICoverLetterDraft: vi.fn(),
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
      listBookmarks: vi.fn(),
      createTrackedApplication: createTrackedApplicationMock,
      updateApplicationStatus: vi.fn(),
      deleteTrackedApplication: deleteTrackedApplicationMock,
      listTrackedApplications: vi.fn(),
    });

    render(
      <AccountTrackerClient
        initialBookmarks={[]}
        initialApplications={[]}
        subscriptionState="free"
      />,
    );

    // Type into the autocomplete input to trigger job search
    fireEvent.change(screen.getByPlaceholderText("Search job title..."), {
      target: { value: "Software" },
    });

    // Wait for dropdown result and click to select
    await waitFor(() => {
      expect(screen.getByText("Software Engineer")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("Software Engineer"));

    fireEvent.click(screen.getByRole("button", { name: "Track application" }));

    await waitFor(() => {
      expect(createTrackedApplicationMock).toHaveBeenCalledTimes(1);
    });
    expect(createTrackedApplicationMock).toHaveBeenCalledWith({
      job_id: "job_abc",
      notes: undefined,
    });
    // The new application should show job title, not raw UUID
    expect(screen.getByText("Software Engineer")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Delete" }));
    await waitFor(() => {
      expect(deleteTrackedApplicationMock).toHaveBeenCalledWith("app_1");
    });
  });
});
