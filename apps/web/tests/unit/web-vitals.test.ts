import { afterEach, describe, expect, it, vi } from "vitest";

import {
  buildWebVitalPayload,
  sendWebVitalMetric,
  type WebVitalMetricInput,
} from "@/lib/observability/web-vitals";

function createMetric(
  overrides: Partial<WebVitalMetricInput> = {},
): WebVitalMetricInput {
  return {
    id: "metric_1",
    name: "LCP",
    value: 1450,
    delta: 1450,
    rating: "good",
    ...overrides,
  };
}

describe("web vitals observability helpers", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("builds normalized telemetry payload", () => {
    const payload = buildWebVitalPayload(createMetric(), "/jobs");

    expect(payload.id).toBe("metric_1");
    expect(payload.name).toBe("LCP");
    expect(payload.rating).toBe("good");
    expect(payload.path).toBe("/jobs");
    expect(Date.parse(payload.timestamp)).not.toBeNaN();
  });

  it("uses sendBeacon when available", () => {
    const sendBeaconMock = vi.fn().mockReturnValue(true);
    Object.defineProperty(window.navigator, "sendBeacon", {
      configurable: true,
      value: sendBeaconMock,
    });

    sendWebVitalMetric(
      buildWebVitalPayload(
        createMetric({ name: "CLS", value: 0.01 }),
        "/pricing",
      ),
    );

    expect(sendBeaconMock).toHaveBeenCalledTimes(1);
    expect(sendBeaconMock).toHaveBeenCalledWith(
      "/api/web-vitals",
      expect.any(Blob),
    );
  });

  it("falls back to fetch when sendBeacon is unavailable", () => {
    Object.defineProperty(window.navigator, "sendBeacon", {
      configurable: true,
      value: undefined,
    });
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response(null, { status: 202 }));

    sendWebVitalMetric(
      buildWebVitalPayload(
        createMetric({ name: "INP", value: 180 }),
        "/account",
      ),
    );

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/web-vitals",
      expect.objectContaining({
        method: "POST",
        keepalive: true,
      }),
    );
  });
});
