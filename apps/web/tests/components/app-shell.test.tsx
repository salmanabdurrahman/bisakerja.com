import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AppShell } from "@/components/layout/app-shell";

describe("AppShell", () => {
  it("renders title and children", () => {
    render(
      <AppShell>
        <div>child-content</div>
      </AppShell>,
    );

    expect(screen.getByRole("heading", { name: "Bisakerja" })).toBeInTheDocument();
    expect(screen.getByText("child-content")).toBeInTheDocument();
  });
});
