/**
 * WebVitalPayload defines the shape of normalized web vital telemetry payload.
 */
export interface WebVitalPayload {
  id: string;
  name: string;
  rating: string;
  value: number;
  delta: number;
  path: string;
  timestamp: string;
}

/**
 * WebVitalMetricInput defines minimal metric fields required for transport.
 */
export interface WebVitalMetricInput {
  id: string;
  name: string;
  rating: string;
  value: number;
  delta: number;
}

/**
 * buildWebVitalPayload normalizes browser web vital metrics before transport.
 */
export function buildWebVitalPayload(
  metric: WebVitalMetricInput,
  path: string,
): WebVitalPayload {
  return {
    id: metric.id,
    name: metric.name,
    rating: metric.rating,
    value: metric.value,
    delta: metric.delta,
    path,
    timestamp: new Date().toISOString(),
  };
}

/**
 * sendWebVitalMetric sends normalized web vital telemetry to same-origin endpoint.
 */
export function sendWebVitalMetric(
  payload: WebVitalPayload,
  endpoint = "/api/web-vitals",
): void {
  if (typeof window === "undefined") {
    return;
  }

  const body = JSON.stringify(payload);
  if (typeof navigator.sendBeacon === "function") {
    const blob = new Blob([body], { type: "application/json" });
    navigator.sendBeacon(endpoint, blob);
    return;
  }

  void fetch(endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body,
    keepalive: true,
  });
}
