import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { APIRequestError } from "@/lib/utils/fetch-json";
import { AccountPreferencesClient } from "@/features/preferences/components/account-preferences-client";
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

describe("AccountPreferencesClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it("stores draft and redirects to login when submit gets 401", async () => {
    const updatePreferencesMock = vi
      .fn()
      .mockRejectedValue(
        new APIRequestError("Unauthorized", 401, "UNAUTHORIZED"),
      );

    vi.mocked(createSessionAPIClient).mockReturnValue({
      getPreferences: vi.fn(),
      getBillingStatus: vi.fn(),
      getMe: vi.fn(),
      updatePreferences: updatePreferencesMock,
    });

    render(
      <AccountPreferencesClient
        initialPreferences={{
          keywords: ["golang"],
          locations: ["jakarta"],
          job_types: ["fulltime"],
          salary_min: 1_000_000,
        }}
        initialUpdatedAt="2026-03-14T00:00:00Z"
        subscriptionState="free"
        infoMessage={null}
      />,
    );

    const keywordsField = screen.getByLabelText("Keywords");
    fireEvent.change(keywordsField, {
      target: { value: "golang, backend" },
    });

    const saveButton = screen.getByRole("button", { name: "Save preferences" });
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(updatePreferencesMock).toHaveBeenCalledTimes(1);
    });

    await waitFor(() => {
      expect(replaceMock).toHaveBeenCalledWith(
        "/auth/login?redirect=%2Faccount%2Fpreferences",
      );
    });

    expect(markAnonymousMock).toHaveBeenCalledTimes(1);
    const storedDraft = window.localStorage.getItem(
      "bisakerja:preferences-draft",
    );
    expect(storedDraft).toContain('"keywords":["golang","backend"]');
  });
});
