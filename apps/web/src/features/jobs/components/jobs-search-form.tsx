import type { JobsSearchParamsState } from "@/features/jobs/search-params";
import { Button } from "@/components/ui/button";

interface JobsSearchFormProps {
  state: JobsSearchParamsState;
}

export function JobsSearchForm({ state }: JobsSearchFormProps) {
  return (
    <form
      action="/jobs"
      method="GET"
      className="bk-card grid gap-5 p-6 sm:p-8 md:grid-cols-2"
      aria-label="Jobs search form"
    >
      <input type="hidden" name="page" value="1" />
      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Keyword</span>
        <input
          type="text"
          name="q"
          defaultValue={state.q}
          placeholder="golang, backend, intern..."
          className="bk-input"
        />
      </label>

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Location</span>
        <input
          type="text"
          name="location"
          defaultValue={state.location}
          placeholder="jakarta, remote..."
          className="bk-input"
        />
      </label>

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Salary minimum</span>
        <input
          type="number"
          min={0}
          name="salary_min"
          defaultValue={state.salaryMin}
          placeholder="10000000"
          className="bk-input"
        />
      </label>

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Source</span>
        <select
          name="source"
          defaultValue={state.source ?? ""}
          className="bk-select"
        >
          <option value="">All sources</option>
          <option value="glints">Glints</option>
          <option value="kalibrr">Kalibrr</option>
          <option value="jobstreet">JobStreet</option>
        </select>
      </label>

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Sort</span>
        <select name="sort" defaultValue={state.sort} className="bk-select">
          <option value="-posted_at">Newest posted</option>
          <option value="posted_at">Oldest posted</option>
          <option value="-created_at">Newest collected</option>
          <option value="created_at">Oldest collected</option>
        </select>
      </label>

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Limit</span>
        <select
          name="limit"
          defaultValue={String(state.limit)}
          className="bk-select"
        >
          <option value="10">10</option>
          <option value="20">20</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </select>
      </label>

      <div className="flex items-end md:col-span-2 mt-4">
        <Button
          type="submit"
          variant="secondary"
          size="lg"
          className="w-full md:w-auto min-w-40"
        >
          Search
        </Button>
      </div>
    </form>
  );
}
