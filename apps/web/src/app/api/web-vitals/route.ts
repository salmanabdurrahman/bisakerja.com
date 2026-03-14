import { NextResponse } from "next/server";

import type { WebVitalPayload } from "@/lib/observability/web-vitals";

function isWebVitalPayload(value: unknown): value is WebVitalPayload {
  if (!value || typeof value !== "object") {
    return false;
  }

  const payload = value as Partial<WebVitalPayload>;
  return (
    typeof payload.id === "string" &&
    typeof payload.name === "string" &&
    typeof payload.rating === "string" &&
    typeof payload.value === "number" &&
    typeof payload.delta === "number" &&
    typeof payload.path === "string" &&
    typeof payload.timestamp === "string"
  );
}

/**
 * POST collects browser web vitals for critical page observability.
 */
export async function POST(request: Request) {
  let payload: unknown;
  try {
    payload = await request.json();
  } catch {
    return NextResponse.json(
      {
        ok: false,
        message: "Invalid JSON payload.",
      },
      { status: 400 },
    );
  }

  if (!isWebVitalPayload(payload)) {
    return NextResponse.json(
      {
        ok: false,
        message: "Invalid web vitals payload.",
      },
      { status: 400 },
    );
  }

  console.info("[web-vitals]", JSON.stringify(payload));

  return NextResponse.json({ ok: true }, { status: 202 });
}
