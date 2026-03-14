import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "@/app/page";

describe("Home page", () => {
  it("renders polished hero content", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", { name: "Bisakerja", level: 1 }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", {
        name: /the professional workspace for job discovery/i,
      }),
    ).toBeInTheDocument();
  });
});
