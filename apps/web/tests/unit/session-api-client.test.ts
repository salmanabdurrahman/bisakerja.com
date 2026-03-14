import { beforeEach, describe, expect, it, vi } from "vitest";

import { APIRequestError } from "@/lib/utils/fetch-json";
import {
  createSessionAPIClient,
  resetSessionRefreshStateForTests,
} from "@/services/session-api-client";
import { getMe } from "@/services/auth";

vi.mock("@/services/auth", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/auth")>("@/services/auth");
  return {
    ...actual,
    getMe: vi.fn(),
  };
});

vi.mock("@/services/billing", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/billing")>(
      "@/services/billing",
    );
  return {
    ...actual,
    getBillingStatus: vi.fn(),
  };
});

vi.mock("@/services/preferences", async () => {
  const actual = await vi.importActual<typeof import("@/services/preferences")>(
    "@/services/preferences",
  );
  return {
    ...actual,
    getPreferences: vi.fn(),
    updatePreferences: vi.fn(),
  };
});

describe("session api client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetSessionRefreshStateForTests();
  });

  it("runs single refresh for concurrent 401 responses", async () => {
    const getSession = vi.fn().mockReturnValue({
      accessToken: "expired-access-token",
      refreshToken: "refresh-token",
      accessExpiresAt: Date.now() + 60_000,
    });
    const updateAccessToken = vi.fn();
    const clearSession = vi.fn();
    const refresh = vi.fn().mockResolvedValue({
      meta: { code: 200, status: "success", message: "Token refreshed" },
      data: {
        access_token: "new-access-token",
        token_type: "Bearer",
        expires_in: 900,
      },
    });

    let requestCount = 0;
    vi.mocked(getMe).mockImplementation(async () => {
      requestCount += 1;
      if (requestCount <= 2) {
        throw new APIRequestError("Unauthorized", 401, "UNAUTHORIZED");
      }
      return {
        meta: { code: 200, status: "success", message: "Profile retrieved" },
        data: {
          id: "user_1",
          email: "user@example.com",
          name: "User",
          role: "user",
          is_premium: false,
          subscription_state: "free",
        },
      };
    });

    const client = createSessionAPIClient({
      getSession,
      updateAccessToken,
      clearSession,
      refresh,
    });

    const [first, second] = await Promise.all([client.getMe(), client.getMe()]);

    expect(first.data.id).toBe("user_1");
    expect(second.data.id).toBe("user_1");
    expect(refresh).toHaveBeenCalledTimes(1);
    expect(updateAccessToken).toHaveBeenCalledTimes(1);
    expect(clearSession).not.toHaveBeenCalled();
  });

  it("clears session when refresh token is missing", async () => {
    const clearSession = vi.fn();
    const client = createSessionAPIClient({
      getSession: vi.fn().mockReturnValue({
        accessToken: null,
        refreshToken: null,
        accessExpiresAt: null,
      }),
      updateAccessToken: vi.fn(),
      clearSession,
      refresh: vi.fn(),
    });

    await expect(client.getMe()).rejects.toBeInstanceOf(APIRequestError);
    expect(clearSession).toHaveBeenCalledTimes(1);
  });
});
