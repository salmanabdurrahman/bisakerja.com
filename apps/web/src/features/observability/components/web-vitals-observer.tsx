"use client";

import { usePathname } from "next/navigation";
import { useReportWebVitals } from "next/web-vitals";

import {
  buildWebVitalPayload,
  sendWebVitalMetric,
} from "@/lib/observability/web-vitals";

/**
 * WebVitalsObserver streams critical web vitals to a same-origin collector endpoint.
 */
export function WebVitalsObserver() {
  const pathname = usePathname();

  useReportWebVitals((metric) => {
    sendWebVitalMetric(buildWebVitalPayload(metric, pathname));
  });

  return null;
}
