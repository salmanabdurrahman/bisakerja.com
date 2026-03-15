import { AppShell } from "@/components/layout/app-shell";
import { ButtonLink } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";
import {
  isHTMLDescription,
  sanitizeJobDescription,
} from "@/lib/utils/sanitize-job-description";
import { formatSalaryDisplay } from "@/lib/utils/format-salary-display";
import { APIRequestError } from "@/lib/utils/fetch-json";

import { getJobDetail, type JobDetail } from "@/services/jobs";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { listBookmarks } from "@/services/tracker";
import { BookmarkButton } from "@/features/tracker/components/bookmark-button";

interface JobDetailPageProps {
  params: Promise<{ id: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

type JobDetailViewState =
  | { kind: "ready"; job: JobDetail }
  | { kind: "not_found" }
  | { kind: "error" };

export default async function JobDetailPage({
  params,
  searchParams,
}: JobDetailPageProps) {
  const { id } = await params;
  const resolvedSearchParams = await searchParams;
  const backHref = buildBackHref(resolvedSearchParams.back);
  const viewState = await loadDetailViewState(id);

  const accessToken = await resolveServerAccessToken();
  let isBookmarked = false;
  if (accessToken) {
    try {
      const bookmarksResponse = await listBookmarks(accessToken);
      isBookmarked = bookmarksResponse.data.some((b) => b.job_id === id);
    } catch {
      // silently skip — bookmark status is non-critical
    }
  }

  return (
    <AppShell>
      <main className="grid gap-5" role="main">
        <PageHeader
          eyebrow="Job Detail"
          title="Opportunity overview"
          description="Review role details, compensation context, and apply directly on the source site."
        />
        {renderDetailView(viewState, backHref, id, isBookmarked)}
      </main>
    </AppShell>
  );
}

function renderDetailView(
  viewState: JobDetailViewState,
  backHref: string,
  jobID: string,
  isBookmarked: boolean,
) {
  if (viewState.kind === "not_found") {
    return (
      <JobsStatePanel
        title="Job not found"
        description="This job may have been removed or is no longer available."
        actionHref={backHref}
        actionLabel="Back to jobs"
      />
    );
  }

  if (viewState.kind === "error") {
    return (
      <JobsStatePanel
        title="Unable to load job detail"
        description="Please try again in a few moments."
        actionHref={backHref}
        actionLabel="Back to jobs"
      />
    );
  }

  const job = viewState.job;
  const normalizedDescription = job.description.trim();
  const richDescription = isHTMLDescription(normalizedDescription);
  const sanitizedDescription = sanitizeJobDescription(normalizedDescription);
  const formattedSalary = formatSalaryDisplay(job.salary_range);

  return (
    <>
      <ButtonLink href={backHref} size="sm" variant="ghost">
        Back to search results
      </ButtonLink>

      <Card>
        <CardHeader className="gap-2">
          <CardTitle>{job.title}</CardTitle>
          <CardDescription className="bk-body">
            {job.company} · {job.location}
          </CardDescription>
          <div className="flex items-center gap-2">
            <BookmarkButton jobID={jobID} initialIsBookmarked={isBookmarked} />
          </div>
          <p className="bk-body">
            Salary:{" "}
            <span className="font-medium text-black">
              {formattedSalary || "Not specified"}
            </span>
          </p>
        </CardHeader>
        <CardContent className="grid gap-4">
          {sanitizedDescription ? (
            richDescription ? (
              <div
                className="job-description-content bk-body [&_a]:underline [&_a]:underline-offset-2 [&_li]:mb-1 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:mb-3 [&_ul]:list-disc [&_ul]:pl-5"
                dangerouslySetInnerHTML={{ __html: sanitizedDescription }}
              />
            ) : (
              <p className="job-description-content bk-body whitespace-pre-wrap">
                {sanitizedDescription}
              </p>
            )
          ) : (
            <p className="job-description-content bk-body-sm text-[#888888]">
              Description is not available for this listing.
            </p>
          )}
          <div>
            <ButtonLink
              href={job.url}
              target="_blank"
              rel="noreferrer"
              variant="secondary"
            >
              Apply on source site
            </ButtonLink>
          </div>
        </CardContent>
      </Card>
    </>
  );
}

async function loadDetailViewState(jobID: string): Promise<JobDetailViewState> {
  try {
    const response = await getJobDetail(jobID);
    return { kind: "ready", job: response.data };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 404) {
      return { kind: "not_found" };
    }
    return { kind: "error" };
  }
}

function buildBackHref(rawBackValue: string | string[] | undefined): string {
  const raw = Array.isArray(rawBackValue) ? rawBackValue[0] : rawBackValue;
  const normalized = (raw ?? "").trim();
  if (!normalized) {
    return "/jobs";
  }

  const withoutPrefix = normalized
    .replace(/^https?:\/\/[^/]+/i, "")
    .replace(/^\/jobs\??/i, "")
    .replace(/^\?/, "");
  return withoutPrefix ? `/jobs?${withoutPrefix}` : "/jobs";
}
