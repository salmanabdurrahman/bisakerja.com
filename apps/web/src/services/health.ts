import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

/**
 * HealthData defines the shape of health data.
 */
export interface HealthData {
  status: string;
  timestamp: string;
}

/**
 * getHealthStatus returns health status.
 */
export async function getHealthStatus() {
  return fetchJSON<HealthData>(buildAPIURL("/healthz"), {
    method: "GET",
    cache: "no-store",
  });
}
