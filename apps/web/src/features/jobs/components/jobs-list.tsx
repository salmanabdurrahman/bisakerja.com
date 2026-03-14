import type { JobListItem } from "@/services/jobs";

interface JobsListProps {
  jobs: JobListItem[];
  detailBaseHref: string;
}

export function JobsList({ jobs, detailBaseHref }: JobsListProps) {
  return (
    <ul className="grid gap-3" aria-label="Jobs list">
      {jobs.map((job) => (
        <li key={job.id} className="rounded-lg border border-gray-200 p-4">
          <h2 className="text-base font-semibold text-gray-900">{job.title}</h2>
          <p className="mt-1 text-sm text-gray-700">{job.company}</p>
          <p className="mt-1 text-sm text-gray-600">{job.location}</p>
          <p className="mt-1 text-xs uppercase tracking-wide text-gray-500">
            {job.source}
          </p>
          <p className="mt-2 text-sm text-gray-700">
            Salary: {job.salary_range || "Not specified"}
          </p>
          <a
            href={`/jobs/${job.id}${detailBaseHref}`}
            className="mt-3 inline-flex text-sm font-medium text-blue-700 underline"
          >
            View detail
          </a>
        </li>
      ))}
    </ul>
  );
}
