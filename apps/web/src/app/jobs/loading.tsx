import { AppShell } from "@/components/layout/app-shell";
import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";

export default function JobsLoadingPage() {
  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <h2 className="text-xl font-semibold">Jobs discovery</h2>
        <JobsStatePanel
          title="Loading jobs"
          description="Please wait while we fetch the latest job listings."
        />
      </main>
    </AppShell>
  );
}
