import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";

describe("JobsStatePanel", () => {
  it("renders title, description, and optional action", () => {
    render(
      <JobsStatePanel
        title="Rate limited"
        description="Please wait before retrying."
        actionHref="/jobs?q=golang"
        actionLabel="Retry search"
      />,
    );

    expect(
      screen.getByRole("heading", { name: "Rate limited" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Please wait before retrying."),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Retry search" })).toHaveAttribute(
      "href",
      "/jobs?q=golang",
    );
  });
});
