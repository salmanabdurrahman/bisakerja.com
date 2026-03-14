import { afterEach, describe, expect, it, vi } from "vitest";

import { buildAPIURL } from "@/services/http-client";

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("buildAPIURL", () => {
  it("builds URL with default base and normalized path", () => {
    const url = buildAPIURL("healthz");
    expect(url).toBe("http://localhost:8080/api/v1/healthz");
  });

  it("uses custom env base URL", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "https://api.example.com/api/v1");

    const url = buildAPIURL("/jobs");
    expect(url).toBe("https://api.example.com/api/v1/jobs");
  });
});
