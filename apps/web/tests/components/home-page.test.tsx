import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "@/app/page";

describe("Home page", () => {
  it("renders phase 0 sections", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", { name: "Bisakerja" }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Phase 0 foundation/i)).toBeInTheDocument();
  });
});
