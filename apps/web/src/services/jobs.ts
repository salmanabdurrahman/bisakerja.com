import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

export type JobSource = "glints" | "kalibrr" | "jobstreet";
export type JobSort = "-posted_at" | "posted_at" | "-created_at" | "created_at";

export interface JobsQuery {
  q?: string;
  location?: string;
  salaryMin?: number;
  page?: number;
  limit?: number;
  sort?: JobSort;
  source?: JobSource;
}

export interface JobListItem {
  id: string;
  title: string;
  company: string;
  location: string;
  salary_range: string;
  source: JobSource;
  posted_at?: string | null;
}

export interface JobDetail {
  id: string;
  title: string;
  company: string;
  location: string;
  description: string;
  salary_range: string;
  source: JobSource;
  url: string;
  posted_at?: string | null;
  created_at?: string;
}

export async function listJobs(
  query: JobsQuery,
  init?: RequestInit,
): Promise<APIResponse<JobListItem[]>> {
  const url = buildAPIURL(`/jobs${buildJobsQuery(query)}`);
  return fetchJSON<JobListItem[]>(url, {
    method: "GET",
    cache: "no-store",
    ...init,
  });
}

export async function getJobDetail(
  jobID: string,
  init?: RequestInit,
): Promise<APIResponse<JobDetail>> {
  return fetchJSON<JobDetail>(buildAPIURL(`/jobs/${jobID}`), {
    method: "GET",
    cache: "no-store",
    ...init,
  });
}

function buildJobsQuery(query: JobsQuery): string {
  const params = new URLSearchParams();
  const q = query.q?.trim();
  const location = query.location?.trim();
  const source = query.source?.trim();
  const sort = query.sort?.trim();

  if (q) params.set("q", q);
  if (location) params.set("location", location);
  if (source) params.set("source", source);
  if (sort) params.set("sort", sort);
  if (query.salaryMin !== undefined)
    params.set("salary_min", String(query.salaryMin));
  if (query.page !== undefined) params.set("page", String(query.page));
  if (query.limit !== undefined) params.set("limit", String(query.limit));

  const serialized = params.toString();
  return serialized ? `?${serialized}` : "";
}
