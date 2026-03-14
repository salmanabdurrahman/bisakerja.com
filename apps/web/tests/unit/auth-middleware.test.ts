import { describe, expect, it } from "vitest";
import { NextRequest } from "next/server";

import { proxy } from "@/proxy";
import {
  ACCESS_TOKEN_COOKIE,
  REFRESH_TOKEN_COOKIE,
} from "@/lib/auth/session-constants";

describe("auth proxy", () => {
  it("redirects anonymous user from protected route", () => {
    const request = new NextRequest("http://localhost/account/preferences");
    const response = proxy(request);

    expect(response.status).toBe(307);
    expect(response.headers.get("location")).toBe(
      "http://localhost/auth/login?redirect=%2Faccount%2Fpreferences",
    );
  });

  it("allows protected route when refresh token exists", () => {
    const request = new NextRequest("http://localhost/account", {
      headers: {
        cookie: `${REFRESH_TOKEN_COOKIE}=refresh-token`,
      },
    });
    const response = proxy(request);

    expect(response.status).toBe(200);
    expect(response.headers.get("x-middleware-next")).toBe("1");
  });

  it("redirects authenticated user away from login page", () => {
    const request = new NextRequest("http://localhost/auth/login", {
      headers: {
        cookie: `${ACCESS_TOKEN_COOKIE}=access-token`,
      },
    });
    const response = proxy(request);

    expect(response.status).toBe(307);
    expect(response.headers.get("location")).toBe("http://localhost/account");
  });

  it("protects pricing route for anonymous users", () => {
    const request = new NextRequest("http://localhost/pricing");
    const response = proxy(request);

    expect(response.status).toBe(307);
    expect(response.headers.get("location")).toBe(
      "http://localhost/auth/login?redirect=%2Fpricing",
    );
  });
});
