import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StateCard } from "@/components/ui/state-card";

describe("StateCard", () => {
  it("renders accessible section with title and description", () => {
    render(<StateCard title="Health" description="Everything is good" />);

    expect(screen.getByRole("region", { name: "Health" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Health" })).toBeInTheDocument();
    expect(screen.getByText("Everything is good")).toBeInTheDocument();
  });
});
