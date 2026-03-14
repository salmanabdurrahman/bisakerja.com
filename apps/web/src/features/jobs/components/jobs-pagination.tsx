import {
  buildPaginationHref,
  type JobsSearchParamsState,
} from "@/features/jobs/search-params";

interface JobsPaginationProps {
  state: JobsSearchParamsState;
  totalPages: number;
  totalRecords: number;
}

export function JobsPagination({
  state,
  totalPages,
  totalRecords,
}: JobsPaginationProps) {
  if (totalPages <= 1) {
    return (
      <p className="text-sm text-gray-600">
        {totalRecords} jobs found. Showing page {state.page}.
      </p>
    );
  }

  return (
    <nav
      aria-label="Jobs pagination"
      className="flex items-center gap-2 text-sm"
    >
      <a
        href={buildPaginationHref(state, state.page - 1)}
        aria-disabled={state.page <= 1}
        className={`rounded-md border px-3 py-2 ${
          state.page <= 1
            ? "pointer-events-none border-gray-200 text-gray-400"
            : "border-gray-300 text-gray-700"
        }`}
      >
        Previous
      </a>
      <span className="text-gray-600">
        Page {state.page} of {totalPages} ({totalRecords} jobs)
      </span>
      <a
        href={buildPaginationHref(state, state.page + 1)}
        aria-disabled={state.page >= totalPages}
        className={`rounded-md border px-3 py-2 ${
          state.page >= totalPages
            ? "pointer-events-none border-gray-200 text-gray-400"
            : "border-gray-300 text-gray-700"
        }`}
      >
        Next
      </a>
    </nav>
  );
}
