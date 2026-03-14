import Link from "next/link";

import {
  buildPaginationHref,
  type JobsSearchParamsState,
} from "@/features/jobs/search-params";
import { buttonVariants } from "@/components/ui/button";

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
      <p className="text-sm text-slate-600">
        {totalRecords} jobs found. Showing page {state.page}.
      </p>
    );
  }

  return (
    <nav
      aria-label="Jobs pagination"
      className="bk-card flex flex-wrap items-center justify-between gap-3 p-4 text-sm"
    >
      <Link
        href={buildPaginationHref(state, state.page - 1)}
        aria-disabled={state.page <= 1}
        className={buttonVariants({
          variant: "outline",
          size: "sm",
          className: state.page <= 1 ? "pointer-events-none opacity-50" : "",
        })}
      >
        Previous
      </Link>
      <span className="text-slate-600">
        Page {state.page} of {totalPages} ({totalRecords} jobs)
      </span>
      <Link
        href={buildPaginationHref(state, state.page + 1)}
        aria-disabled={state.page >= totalPages}
        className={buttonVariants({
          variant: "outline",
          size: "sm",
          className:
            state.page >= totalPages ? "pointer-events-none opacity-50" : "",
        })}
      >
        Next
      </Link>
    </nav>
  );
}
