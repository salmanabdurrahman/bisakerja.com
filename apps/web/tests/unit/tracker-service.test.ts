import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createBookmark,
  deleteBookmark,
  listTrackedApplications,
} from "@/services/tracker";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

describe("tracker services", () => {
  it("creates bookmark with authenticated request", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 201,
            status: "success",
            message: "Bookmark created",
          },
          data: {
            id: "bm_1",
            job_id: "job_abc",
            created_at: "2030-01-01T00:00:00Z",
          },
        }),
        { status: 201 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await createBookmark("access-token", "job_abc");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/bookmarks",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });

  it("deletes bookmark", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 200,
            status: "success",
            message: "Bookmark deleted",
          },
          data: {
            job_id: "job_abc",
          },
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await deleteBookmark("access-token", "job_abc");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/bookmarks/job_abc",
      expect.objectContaining({
        method: "DELETE",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });

  it("lists tracked applications", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 200,
            status: "success",
            message: "Applications retrieved",
          },
          data: [],
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await listTrackedApplications("access-token");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/applications",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });
});
