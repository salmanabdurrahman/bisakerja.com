import type { JobSort, JobSource, JobsQuery } from "@/services/jobs";

export interface JobsSearchParamsState {
  q: string;
  location: string;
  salaryMin?: number;
  source?: JobSource;
  sort: JobSort;
  page: number;
  limit: number;
}

const allowedSort: JobSort[] = [
  "-posted_at",
  "posted_at",
  "-created_at",
  "created_at",
];
const allowedSource: JobSource[] = ["glints", "kalibrr", "jobstreet"];

export function normalizeJobsSearchParams(
  searchParams: Record<string, string | string[] | undefined>,
): JobsSearchParamsState {
  const q = normalizeText(getFirst(searchParams.q));
  const location = normalizeText(getFirst(searchParams.location));
  const sortCandidate = normalizeText(getFirst(searchParams.sort));
  const sourceCandidate = normalizeText(getFirst(searchParams.source));
  const page = parsePositiveInt(getFirst(searchParams.page), 1);
  const limit = parseInRange(getFirst(searchParams.limit), 20, 1, 100);
  const salaryMin = parseOptionalNonNegativeInt(
    getFirst(searchParams.salary_min),
  );

  const sort = allowedSort.includes(sortCandidate as JobSort)
    ? (sortCandidate as JobSort)
    : "-posted_at";
  const source = allowedSource.includes(sourceCandidate as JobSource)
    ? (sourceCandidate as JobSource)
    : undefined;

  return {
    q,
    location,
    salaryMin,
    source,
    sort,
    page,
    limit,
  };
}

export function toJobsQuery(state: JobsSearchParamsState): JobsQuery {
  return {
    q: state.q || undefined,
    location: state.location || undefined,
    salaryMin: state.salaryMin,
    source: state.source,
    sort: state.sort,
    page: state.page,
    limit: state.limit,
  };
}

export function buildSearchSubmitHref(input: {
  q?: string;
  location?: string;
  salaryMin?: string;
  source?: string;
  sort?: string;
  limit?: string;
}): string {
  const nextState: JobsSearchParamsState = {
    q: normalizeText(input.q),
    location: normalizeText(input.location),
    salaryMin: parseOptionalNonNegativeInt(normalizeText(input.salaryMin)),
    source: normalizeSource(input.source),
    sort: normalizeSort(input.sort),
    page: 1,
    limit: parseInRange(normalizeText(input.limit), 20, 1, 100),
  };

  return `/jobs${serializeState(nextState)}`;
}

export function buildPaginationHref(
  state: JobsSearchParamsState,
  page: number,
): string {
  const nextPage = page < 1 ? 1 : page;
  return `/jobs${serializeState({ ...state, page: nextPage })}`;
}

function serializeState(state: JobsSearchParamsState): string {
  const params = new URLSearchParams();
  if (state.q) params.set("q", state.q);
  if (state.location) params.set("location", state.location);
  if (state.salaryMin !== undefined)
    params.set("salary_min", String(state.salaryMin));
  if (state.source) params.set("source", state.source);
  if (state.sort) params.set("sort", state.sort);
  if (state.page !== 1) params.set("page", String(state.page));
  if (state.limit !== 20) params.set("limit", String(state.limit));

  const queryString = params.toString();
  return queryString ? `?${queryString}` : "";
}

function normalizeSort(rawValue?: string): JobSort {
  const value = normalizeText(rawValue);
  return allowedSort.includes(value as JobSort)
    ? (value as JobSort)
    : "-posted_at";
}

function normalizeSource(rawValue?: string): JobSource | undefined {
  const value = normalizeText(rawValue);
  return allowedSource.includes(value as JobSource)
    ? (value as JobSource)
    : undefined;
}

function getFirst(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }
  return value;
}

function normalizeText(value?: string): string {
  return (value ?? "").trim();
}

function parsePositiveInt(value: string | undefined, fallback: number): number {
  const parsed = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(parsed) || parsed < 1) {
    return fallback;
  }
  return parsed;
}

function parseInRange(
  value: string | undefined,
  fallback: number,
  min: number,
  max: number,
): number {
  const parsed = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(parsed) || parsed < min || parsed > max) {
    return fallback;
  }
  return parsed;
}

function parseOptionalNonNegativeInt(
  value: string | undefined,
): number | undefined {
  if (!value) {
    return undefined;
  }
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || parsed < 0) {
    return undefined;
  }
  return parsed;
}
