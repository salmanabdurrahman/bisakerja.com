import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { AppShell } from "@/components/layout/app-shell";
import { ACCESS_TOKEN_COOKIE } from "@/lib/auth/session-constants";

describe("AppShell", () => {
  beforeEach(() => {
    document.cookie = `${ACCESS_TOKEN_COOKIE}=; Path=/; Max-Age=0`;
  });

  it("renders title and children", () => {
    render(
      <AppShell>
        <div>child-content</div>
      </AppShell>,
    );

    expect(
      screen.getByRole("heading", { name: "Bisakerja", level: 1 }),
    ).toBeInTheDocument();
    expect(screen.getByText("child-content")).toBeInTheDocument();
  });

  it("shows dashboard button for authenticated session", () => {
    document.cookie = `${ACCESS_TOKEN_COOKIE}=token; Path=/`;

    render(
      <AppShell>
        <div>authenticated-content</div>
      </AppShell>,
    );

    expect(screen.getByRole("link", { name: "Dashboard" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Log in" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Get Started" })).toBeNull();
  });
});
