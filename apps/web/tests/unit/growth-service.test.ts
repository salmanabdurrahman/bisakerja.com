import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createSavedSearch,
  listNotifications,
  markNotificationAsRead,
} from "@/services/growth";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

describe("growth services", () => {
  it("creates saved search with authenticated request", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 201,
            status: "success",
            message: "Saved search created",
          },
          data: {
            id: "ss_1",
            query: "golang backend",
            location: "jakarta",
            source: "glints",
            salary_min: 12000000,
            frequency: "daily_digest",
            is_active: true,
            created_at: "2030-01-01T00:00:00Z",
            updated_at: "2030-01-01T00:00:00Z",
          },
        }),
        { status: 201 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await createSavedSearch("access-token", {
      query: "golang backend",
      location: "jakarta",
      source: "glints",
      salary_min: 12000000,
      frequency: "daily_digest",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/saved-searches",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });

  it("builds query params for notifications endpoint", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 200,
            status: "success",
            message: "Notifications retrieved",
            pagination: {
              page: 2,
              limit: 20,
              total_pages: 3,
              total_records: 55,
            },
          },
          data: [],
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await listNotifications("access-token", {
      page: 2,
      limit: 20,
      unread_only: true,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/notifications?page=2&limit=20&unread_only=true",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });

  it("marks notification as read", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: {
            code: 200,
            status: "success",
            message: "Notification marked as read",
          },
          data: {
            id: "notif_1",
            job_id: "job_1",
            channel: "email",
            status: "sent",
            sent_at: "2030-01-01T00:00:00Z",
            read_at: "2030-01-01T01:00:00Z",
            created_at: "2030-01-01T00:00:00Z",
          },
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await markNotificationAsRead("access-token", "notif_1");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/notifications/notif_1/read",
      expect.objectContaining({
        method: "PATCH",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });
});
