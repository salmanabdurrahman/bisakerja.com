import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import JobDetailPage from "@/app/jobs/[id]/page";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getJobDetail } from "@/services/jobs";

vi.mock("@/services/jobs", () => ({
  listJobs: vi.fn(),
  getJobDetail: vi.fn(),
}));

describe("Job detail page", () => {
  it("renders not-found state when API returns 404", async () => {
    vi.mocked(getJobDetail).mockRejectedValueOnce(
      new APIRequestError("Job not found", 404, "NOT_FOUND"),
    );

    const page = await JobDetailPage({
      params: Promise.resolve({ id: "job_missing" }),
      searchParams: Promise.resolve({ back: "q=golang&page=2" }),
    });
    render(page);

    expect(
      screen.getByRole("heading", { name: "Job not found" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Back to jobs" })).toHaveAttribute(
      "href",
      "/jobs?q=golang&page=2",
    );
  });
});
