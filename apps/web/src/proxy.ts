import { NextResponse, type NextRequest } from "next/server";

import {
  ACCESS_TOKEN_COOKIE,
  REFRESH_TOKEN_COOKIE,
} from "@/lib/auth/session-constants";
import { normalizeRedirectPath } from "@/lib/auth/redirect-path";

export function proxy(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  const hasSession = Boolean(
    request.cookies.get(ACCESS_TOKEN_COOKIE)?.value ||
    request.cookies.get(REFRESH_TOKEN_COOKIE)?.value,
  );
  const isProtectedPath =
    pathname.startsWith("/account") ||
    pathname === "/pricing" ||
    pathname.startsWith("/billing/success");

  if (isProtectedPath && !hasSession) {
    const redirectPath = normalizeRedirectPath(
      `${pathname}${request.nextUrl.search}`,
    );
    const target = new URL(
      `/auth/login?redirect=${encodeURIComponent(redirectPath)}`,
      request.url,
    );
    return NextResponse.redirect(target);
  }

  if (
    (pathname === "/auth/login" || pathname === "/auth/register") &&
    hasSession
  ) {
    return NextResponse.redirect(new URL("/account", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/account/:path*",
    "/pricing",
    "/billing/success",
    "/auth/login",
    "/auth/register",
  ],
};
