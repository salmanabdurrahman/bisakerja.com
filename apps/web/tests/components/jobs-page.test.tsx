import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { APIRequestError } from "@/lib/utils/fetch-json";
import JobsPage from "@/app/jobs/page";
import { listJobs } from "@/services/jobs";

vi.mock("@/services/jobs", () => ({
  listJobs: vi.fn(),
  getJobDetail: vi.fn(),
}));

describe("Jobs page", () => {
  it("renders jobs list for successful response", async () => {
    vi.mocked(listJobs).mockResolvedValueOnce({
      meta: {
        code: 200,
        status: "success",
        message: "Jobs retrieved",
        pagination: {
          page: 1,
          limit: 20,
          total_pages: 1,
          total_records: 1,
        },
      },
      data: [
        {
          id: "job_1",
          title: "Backend Engineer",
          company: "Bisakerja",
          location: "Jakarta",
          salary_range: "10000000 - 15000000",
          source: "glints",
          posted_at: "2026-03-14T00:00:00Z",
        },
      ],
    });

    const page = await JobsPage({
      searchParams: Promise.resolve({ q: "backend" }),
    });
    render(page);

    expect(
      screen.getByRole("heading", { name: "Jobs discovery" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Backend Engineer")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "View detail" }),
    ).toBeInTheDocument();
  });

  it("renders rate limited state for 429 response", async () => {
    vi.mocked(listJobs).mockRejectedValueOnce(
      new APIRequestError("Too many requests", 429, "TOO_MANY_REQUESTS"),
    );

    const page = await JobsPage({
      searchParams: Promise.resolve({ q: "backend", page: "2" }),
    });
    render(page);

    expect(
      screen.getByRole("heading", {
        name: "Search is temporarily rate limited",
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Retry search" }),
    ).toBeInTheDocument();
  });
});
