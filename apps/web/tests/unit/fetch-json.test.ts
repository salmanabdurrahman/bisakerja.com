import { describe, expect, it, vi } from "vitest";

import { fetchJSON } from "@/lib/utils/fetch-json";

describe("fetchJSON", () => {
  it("returns parsed payload for successful response", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: { code: 200, status: "success", message: "OK" },
          data: { value: 1 },
        }),
        { status: 200 },
      ),
    );

    vi.stubGlobal("fetch", fetchMock);

    const response = await fetchJSON<{ value: number }>("/api/v1/healthz", {
      method: "GET",
      headers: { "X-Test": "1" },
    });

    expect(response.data.value).toBe(1);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/healthz",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          "Content-Type": "application/json",
          "X-Test": "1",
        }),
      }),
    );
  });

  it("throws for non-ok response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response("error", { status: 503 })),
    );

    await expect(fetchJSON("/api/v1/fail")).rejects.toThrow(
      "Request failed with status 503",
    );
  });
});
