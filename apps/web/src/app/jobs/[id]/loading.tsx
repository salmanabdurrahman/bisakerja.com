import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/ui/page-header";
import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";

export default function JobDetailLoadingPage() {
  return (
    <AppShell>
      <main className="grid gap-5" role="main">
        <PageHeader
          eyebrow="Job Detail"
          title="Opportunity overview"
          description="Loading role details from the selected source."
        />
        <JobsStatePanel
          title="Loading job detail"
          description="Please wait while we load the selected job."
        />
      </main>
    </AppShell>
  );
}
