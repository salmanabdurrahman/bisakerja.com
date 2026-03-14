import type { JobsSearchParamsState } from "@/features/jobs/search-params";

interface JobsSearchFormProps {
  state: JobsSearchParamsState;
}

export function JobsSearchForm({ state }: JobsSearchFormProps) {
  return (
    <form
      action="/jobs"
      method="GET"
      className="grid gap-3 rounded-lg border border-gray-200 p-4 md:grid-cols-2"
      aria-label="Jobs search form"
    >
      <input type="hidden" name="page" value="1" />
      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Keyword</span>
        <input
          type="text"
          name="q"
          defaultValue={state.q}
          placeholder="golang, backend, intern..."
          className="rounded-md border border-gray-300 px-3 py-2"
        />
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Location</span>
        <input
          type="text"
          name="location"
          defaultValue={state.location}
          placeholder="jakarta, remote..."
          className="rounded-md border border-gray-300 px-3 py-2"
        />
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Salary minimum</span>
        <input
          type="number"
          min={0}
          name="salary_min"
          defaultValue={state.salaryMin}
          placeholder="10000000"
          className="rounded-md border border-gray-300 px-3 py-2"
        />
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Source</span>
        <select
          name="source"
          defaultValue={state.source ?? ""}
          className="rounded-md border border-gray-300 px-3 py-2"
        >
          <option value="">All sources</option>
          <option value="glints">Glints</option>
          <option value="kalibrr">Kalibrr</option>
          <option value="jobstreet">JobStreet</option>
        </select>
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Sort</span>
        <select
          name="sort"
          defaultValue={state.sort}
          className="rounded-md border border-gray-300 px-3 py-2"
        >
          <option value="-posted_at">Newest posted</option>
          <option value="posted_at">Oldest posted</option>
          <option value="-created_at">Newest collected</option>
          <option value="created_at">Oldest collected</option>
        </select>
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Limit</span>
        <select
          name="limit"
          defaultValue={String(state.limit)}
          className="rounded-md border border-gray-300 px-3 py-2"
        >
          <option value="10">10</option>
          <option value="20">20</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </select>
      </label>

      <div className="flex items-end">
        <button
          type="submit"
          className="rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Search
        </button>
      </div>
    </form>
  );
}
