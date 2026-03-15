import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import JobDetailPage from "@/app/jobs/[id]/page";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getJobDetail } from "@/services/jobs";

vi.mock("@/services/jobs", () => ({
  listJobs: vi.fn(),
  getJobDetail: vi.fn(),
}));

vi.mock("@/lib/auth/server-session", () => ({
  resolveServerAccessToken: vi.fn().mockResolvedValue(null),
}));

vi.mock("@/services/tracker", () => ({
  listBookmarks: vi.fn().mockResolvedValue({ data: [] }),
}));

describe("Job detail page", () => {
  it("sanitizes rich HTML descriptions from API response", async () => {
    vi.mocked(getJobDetail).mockResolvedValueOnce({
      meta: {
        code: 200,
        status: "success",
        message: "Job detail retrieved",
      },
      data: {
        id: "job_123",
        title: "Backend Engineer",
        company: "Acme",
        location: "Jakarta",
        description:
          '<p>Build APIs safely.</p><script>alert("xss")</script><a href="javascript:alert(1)">Click me</a><ul><li>Write tests</li></ul>',
        salary_range: "10000000",
        source: "kalibrr",
        url: "https://example.com/jobs/job_123",
        posted_at: "2026-03-15T00:00:00Z",
        created_at: "2026-03-15T00:00:00Z",
      },
    });

    const page = await JobDetailPage({
      params: Promise.resolve({ id: "job_123" }),
      searchParams: Promise.resolve({}),
    });
    const { container } = render(page);

    expect(screen.getByText("Build APIs safely.")).toBeInTheDocument();
    expect(screen.getByText("Write tests")).toBeInTheDocument();
    expect(screen.getByText("Salary:")).toBeInTheDocument();
    expect(screen.getByText("Rp 10.000.000")).toBeInTheDocument();
    expect(container.querySelector("script")).not.toBeInTheDocument();
    expect(
      container.querySelector(".job-description-content")?.innerHTML,
    ).not.toContain("javascript:");
  });

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
