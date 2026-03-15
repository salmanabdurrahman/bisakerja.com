"use client";

import { useCallback, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import type { JobsSearchParamsState } from "@/features/jobs/search-params";
import { buildSearchSubmitHref } from "@/features/jobs/search-params";
import type { JobSort } from "@/services/jobs";
import { Button } from "@/components/ui/button";

interface JobsSearchFormProps {
  state: JobsSearchParamsState;
}

export function JobsSearchForm({ state }: JobsSearchFormProps) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();

  // Local form state initialized from URL params
  const [q, setQ] = useState(state.q);
  const [location, setLocation] = useState(state.location);
  const [salaryMin, setSalaryMin] = useState(
    state.salaryMin !== undefined ? String(state.salaryMin) : "",
  );
  const [source, setSource] = useState(state.source ?? "");
  const [sort, setSort] = useState(state.sort);
  const [limit, setLimit] = useState(String(state.limit));

  const handleSearch = useCallback(() => {
    const href = buildSearchSubmitHref({
      q,
      location,
      salaryMin,
      source,
      sort,
      limit,
    });

    startTransition(() => {
      router.push(href);
    });
  }, [q, location, salaryMin, source, sort, limit, router]);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    handleSearch();
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="bk-card grid gap-5 p-6 sm:p-8 md:grid-cols-2"
      aria-label="Jobs search form"
    >
      <label className="grid gap-2 bk-label">
        <span className="font-medium text-black">Keyword</span>
        <input
          type="text"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="golang, backend, intern..."
          className="bk-input"
        />
      </label>

      <label className="grid gap-2 bk-label">
        <span className="font-medium text-black">Location</span>
        <input
          type="text"
          value={location}
          onChange={(e) => setLocation(e.target.value)}
          placeholder="jakarta, remote..."
          className="bk-input"
        />
      </label>

      <label className="grid gap-2 bk-label">
        <span className="font-medium text-black">Salary minimum</span>
        <input
          type="number"
          min={0}
          value={salaryMin}
          onChange={(e) => setSalaryMin(e.target.value)}
          placeholder="10000000"
          className="bk-input"
        />
      </label>

      <label className="grid gap-2 bk-label">
        <span className="font-medium text-black">Source</span>
        <select
          value={source}
          onChange={(e) => setSource(e.target.value)}
          className="bk-select"
        >
          <option value="">All sources</option>
          <option value="glints">Glints</option>
          <option value="kalibrr">Kalibrr</option>
          <option value="jobstreet">JobStreet</option>
        </select>
      </label>

      <label className="grid gap-2 bk-label">
        <span className="font-medium text-black">Sort</span>
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as JobSort)}
          className="bk-select"
        >
          <option value="-posted_at">Newest posted</option>
          <option value="posted_at">Oldest posted</option>
          <option value="-created_at">Newest collected</option>
          <option value="created_at">Oldest collected</option>
        </select>
      </label>

      <label className="grid gap-2 bk-label">
        <span className="font-medium text-black">Limit</span>
        <select
          value={limit}
          onChange={(e) => setLimit(e.target.value)}
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
          disabled={isPending}
        >
          {isPending ? "Searching..." : "Search"}
        </Button>
      </div>
    </form>
  );
}
