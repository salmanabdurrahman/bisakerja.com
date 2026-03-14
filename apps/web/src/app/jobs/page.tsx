import { AppShell } from "@/components/layout/app-shell";
import { JobsList } from "@/features/jobs/components/jobs-list";
import { JobsPagination } from "@/features/jobs/components/jobs-pagination";
import { JobsSearchForm } from "@/features/jobs/components/jobs-search-form";
import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";
import {
  buildPaginationHref,
  normalizeJobsSearchParams,
  toJobsQuery,
  type JobsSearchParamsState,
} from "@/features/jobs/search-params";
import type { APIPagination, APIResponse } from "@/lib/types/api";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { listJobs, type JobListItem } from "@/services/jobs";

interface JobsPageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

type JobsViewState =
  | {
      kind: "ready";
      response: APIResponse<JobListItem[]>;
      pagination: APIPagination;
      detailBaseHref: string;
    }
  | { kind: "validation_error" }
  | { kind: "rate_limited" }
  | { kind: "error" };

export default async function JobsPage({ searchParams }: JobsPageProps) {
  const state = normalizeJobsSearchParams(await searchParams);
  const currentHref = buildPaginationHref(state, state.page);
  const viewState = await loadJobsViewState(state);

  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <h2 className="text-xl font-semibold">Jobs discovery</h2>
        <JobsSearchForm state={state} />
        {renderJobsView(viewState, state, currentHref)}
      </main>
    </AppShell>
  );
}

function renderJobsView(
  viewState: JobsViewState,
  state: JobsSearchParamsState,
  currentHref: string,
) {
  if (viewState.kind === "validation_error") {
    return (
      <JobsStatePanel
        title="Invalid search filter"
        description="Some filter values are invalid. Adjust your input and try again."
        actionHref="/jobs"
        actionLabel="Reset filters"
      />
    );
  }

  if (viewState.kind === "rate_limited") {
    return (
      <JobsStatePanel
        title="Search is temporarily rate limited"
        description="Please wait a moment before retrying this search query."
        actionHref={currentHref}
        actionLabel="Retry search"
      />
    );
  }

  if (viewState.kind === "error") {
    return (
      <JobsStatePanel
        title="Unable to load jobs"
        description="The service is currently unavailable. Please retry shortly."
        actionHref={currentHref}
        actionLabel="Try again"
      />
    );
  }

  if (viewState.response.data.length === 0) {
    return (
      <JobsStatePanel
        title="No jobs found"
        description="Try changing keyword, location, source, or reset your filters."
        actionHref="/jobs"
        actionLabel="Reset filters"
      />
    );
  }

  return (
    <>
      <JobsPagination
        state={state}
        totalPages={viewState.pagination.total_pages}
        totalRecords={viewState.pagination.total_records}
      />
      <JobsList
        jobs={viewState.response.data}
        detailBaseHref={viewState.detailBaseHref}
      />
    </>
  );
}

async function loadJobsViewState(
  state: JobsSearchParamsState,
): Promise<JobsViewState> {
  try {
    const response = await listJobs(toJobsQuery(state));
    const pagination = response.meta.pagination ?? {
      page: state.page,
      limit: state.limit,
      total_pages: state.page,
      total_records: response.data.length,
    };
    const backQuery = buildPaginationHref(state, state.page).replace(
      /^\/jobs\??/,
      "",
    );
    const detailBaseHref = backQuery
      ? `?back=${encodeURIComponent(backQuery)}`
      : "";

    return {
      kind: "ready",
      response,
      pagination,
      detailBaseHref,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 400) {
      return { kind: "validation_error" };
    }
    if (error instanceof APIRequestError && error.status === 429) {
      return { kind: "rate_limited" };
    }
    return { kind: "error" };
  }
}
