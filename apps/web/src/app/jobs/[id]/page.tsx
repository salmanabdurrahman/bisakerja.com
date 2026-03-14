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
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getJobDetail, type JobDetail } from "@/services/jobs";

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

  return (
    <AppShell>
      <main className="grid gap-5" role="main">
        <PageHeader
          eyebrow="Job Detail"
          title="Opportunity overview"
          description="Review role details, compensation context, and apply directly on the source site."
        />
        {renderDetailView(viewState, backHref)}
      </main>
    </AppShell>
  );
}

function renderDetailView(viewState: JobDetailViewState, backHref: string) {
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
  return (
    <>
      <ButtonLink href={backHref} size="sm" variant="ghost">
        Back to search results
      </ButtonLink>

      <Card>
        <CardHeader className="gap-2">
          <CardTitle className="text-2xl">{job.title}</CardTitle>
          <CardDescription className="text-sm text-slate-700">
            {job.company} · {job.location}
          </CardDescription>
          <p className="text-sm font-medium text-slate-700">
            Salary: {job.salary_range || "Not specified"}
          </p>
        </CardHeader>
        <CardContent className="grid gap-4">
          <p className="whitespace-pre-wrap text-sm text-slate-700">
            {job.description}
          </p>
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
