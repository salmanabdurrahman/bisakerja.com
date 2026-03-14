import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/ui/page-header";
import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";

export default function JobsLoadingPage() {
  return (
    <AppShell>
      <main className="grid gap-5" role="main">
        <PageHeader
          eyebrow="Discovery"
          title="Jobs discovery"
          description="Fetching the latest opportunities and ranking them for your current search."
        />
        <JobsStatePanel
          title="Loading jobs"
          description="Please wait while we fetch the latest job listings."
        />
      </main>
    </AppShell>
  );
}
