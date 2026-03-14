import { ButtonLink } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { JobListItem } from "@/services/jobs";

interface JobsListProps {
  jobs: JobListItem[];
  detailBaseHref: string;
}

export function JobsList({ jobs, detailBaseHref }: JobsListProps) {
  return (
    <ul className="grid gap-6" aria-label="Jobs list">
      {jobs.map((job) => (
        <li key={job.id}>
          <Card>
            <CardHeader className="gap-2">
              <p className="inline-flex w-fit rounded-full border border-[#E5E5E5] bg-[#F9F9F9] px-3 py-1 text-[12px] font-medium text-[#666666] uppercase tracking-wider">
                {job.source}
              </p>
              <CardTitle className="text-[20px] sm:text-[24px] mt-2">
                {job.title}
              </CardTitle>
              <p className="text-[16px] text-black">{job.company}</p>
              <p className="text-[14px] text-[#666666]">{job.location}</p>
            </CardHeader>
            <CardContent className="flex flex-wrap items-center justify-between gap-4">
              <p className="text-[14px] text-[#666666]">
                Salary:{" "}
                <span className="text-black">
                  {job.salary_range || "Not specified"}
                </span>
              </p>
              <ButtonLink
                href={`/jobs/${job.id}${detailBaseHref}`}
                variant="outline"
                size="sm"
              >
                View detail
              </ButtonLink>
            </CardContent>
          </Card>
        </li>
      ))}
    </ul>
  );
}
