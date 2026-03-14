import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountSavedSearchesClient } from "@/features/growth/components/account-saved-searches-client";
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

describe("AccountSavedSearchesClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it("creates and deletes saved search", async () => {
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
      markNotificationAsRead: vi.fn(),
      updateNotificationPreferences: vi.fn(),
    });

    render(<AccountSavedSearchesClient initialSavedSearches={[]} />);

    fireEvent.change(screen.getByLabelText("Query"), {
      target: { value: "golang backend" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Tambah saved search" }),
    );

    await waitFor(() => {
      expect(createSavedSearchMock).toHaveBeenCalledTimes(1);
    });
    expect(screen.getByText("golang backend")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Hapus" }));
    await waitFor(() => {
      expect(deleteSavedSearchMock).toHaveBeenCalledWith("ss_1");
    });
  });
});
