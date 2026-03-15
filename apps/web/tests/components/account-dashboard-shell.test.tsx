import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";

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

describe("AccountDashboardShell", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders shared dashboard navigation and handles logout", () => {
    render(
      <AccountDashboardShell
        eyebrow="Account"
        title="Preferences"
        description="Account page description"
      >
        <div>account-page-content</div>
      </AccountDashboardShell>,
    );

    const nav = screen.getByRole("navigation", {
      name: "Account dashboard navigation",
    });

    expect(
      within(nav).getByRole("link", { name: "Account overview" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Manage preferences" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Saved searches" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Notification center" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Subscription" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "AI tools" }),
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Logout" }));

    expect(markAnonymousMock).toHaveBeenCalledTimes(1);
    expect(replaceMock).toHaveBeenCalledWith("/auth/login");
  });
});
