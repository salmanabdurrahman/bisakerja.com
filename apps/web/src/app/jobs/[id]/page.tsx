import { AppShell } from "@/components/layout/app-shell";
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
      <main className="grid gap-4" role="main">
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
      <a href={backHref} className="text-sm text-blue-700 underline">
        Back to search results
      </a>

      <article className="rounded-lg border border-gray-200 p-4">
        <h2 className="text-xl font-semibold text-gray-900">{job.title}</h2>
        <p className="mt-2 text-sm text-gray-700">{job.company}</p>
        <p className="mt-1 text-sm text-gray-600">{job.location}</p>
        <p className="mt-2 text-sm text-gray-700">
          Salary: {job.salary_range || "Not specified"}
        </p>
        <p className="mt-4 whitespace-pre-wrap text-sm text-gray-700">
          {job.description}
        </p>
        <a
          href={job.url}
          target="_blank"
          rel="noreferrer"
          className="mt-4 inline-flex rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Apply on source site
        </a>
      </article>
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
