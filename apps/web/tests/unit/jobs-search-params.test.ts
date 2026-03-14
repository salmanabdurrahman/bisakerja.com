import { describe, expect, it } from "vitest";

import {
  buildPaginationHref,
  buildSearchSubmitHref,
  normalizeJobsSearchParams,
} from "@/features/jobs/search-params";

describe("jobs search params", () => {
  it("normalizes search params with defaults", () => {
    const normalized = normalizeJobsSearchParams({
      q: " golang ",
      location: " jakarta ",
      source: "glints",
      sort: "-posted_at",
      salary_min: "10000000",
      page: "2",
      limit: "20",
    });

    expect(normalized.q).toBe("golang");
    expect(normalized.location).toBe("jakarta");
    expect(normalized.source).toBe("glints");
    expect(normalized.salaryMin).toBe(10000000);
    expect(normalized.page).toBe(2);
    expect(normalized.limit).toBe(20);
    expect(normalized.sort).toBe("-posted_at");
  });

  it("builds submit href and resets page to 1", () => {
    const href = buildSearchSubmitHref({
      q: "backend",
      location: "remote",
      source: "kalibrr",
      sort: "posted_at",
      limit: "50",
    });

    expect(href).toContain("/jobs?");
    expect(href).toContain("q=backend");
    expect(href).toContain("location=remote");
    expect(href).not.toContain("page=");
  });

  it("builds pagination href preserving current filters", () => {
    const current = normalizeJobsSearchParams({
      q: "golang",
      location: "jakarta",
      page: "1",
      sort: "-posted_at",
    });
    const href = buildPaginationHref(current, 3);

    expect(href).toContain("/jobs?");
    expect(href).toContain("q=golang");
    expect(href).toContain("location=jakarta");
    expect(href).toContain("page=3");
  });
});
