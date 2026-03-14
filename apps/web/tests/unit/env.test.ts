import { describe, expect, it } from "vitest";

import { getAPIBaseURL } from "@/lib/config/env";

describe("getAPIBaseURL", () => {
  it("returns fallback URL when env is not set", () => {
    expect(getAPIBaseURL()).toBe("http://localhost:8080/api/v1");
  });
});
