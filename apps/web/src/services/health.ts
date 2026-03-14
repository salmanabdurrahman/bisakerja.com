import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

export interface HealthData {
  status: string;
  timestamp: string;
}

export async function getHealthStatus() {
  return fetchJSON<HealthData>(buildAPIURL("/healthz"), {
    method: "GET",
    cache: "no-store",
  });
}
