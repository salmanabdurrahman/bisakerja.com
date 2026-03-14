import { describe, expect, it } from "vitest";

import {
  buildServerAPIBaseURL,
  normalizeAPIBasePath,
  resolveAPIOrigin,
} from "@/lib/config/api-base";

describe("api base config", () => {
  it("uses default public base path when value is empty", () => {
    expect(normalizeAPIBasePath(undefined)).toBe("/api/v1");
  });

  it("normalizes relative path base", () => {
    expect(normalizeAPIBasePath("api/v2/")).toBe("/api/v2");
  });

  it("extracts path from absolute public API URL", () => {
    expect(normalizeAPIBasePath("https://api.example.com/api/v1/")).toBe(
      "/api/v1",
    );
  });

  it("falls back to default path for absolute URL without pathname", () => {
    expect(normalizeAPIBasePath("https://api.example.com")).toBe("/api/v1");
  });

  it("resolves origin from full API origin URL", () => {
    expect(resolveAPIOrigin("https://api.example.com/api/v1")).toBe(
      "https://api.example.com",
    );
  });

  it("supports host:port API origin input", () => {
    expect(resolveAPIOrigin("localhost:9000")).toBe("http://localhost:9000");
  });

  it("builds server API base URL from origin and path", () => {
    expect(
      buildServerAPIBaseURL(
        "https://backend.example.com/api/v1",
        "https://public.example.com/custom/v2",
      ),
    ).toBe("https://backend.example.com/custom/v2");
  });
});
