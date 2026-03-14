import { afterEach, describe, expect, it, vi } from "vitest";

import { getAPIBaseURL } from "@/lib/config/env";

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("getAPIBaseURL", () => {
  it("returns fallback URL when env is not set", () => {
    expect(getAPIBaseURL()).toBe("/api/v1");
  });

  it("normalizes absolute public API URL into same-origin path", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "https://api.example.com/api/v1/");
    expect(getAPIBaseURL()).toBe("/api/v1");
  });

  it("normalizes path-only API base with missing leading slash", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "api/v2/");
    expect(getAPIBaseURL()).toBe("/api/v2");
  });
});
