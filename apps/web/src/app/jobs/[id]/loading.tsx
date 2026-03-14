import { AppShell } from "@/components/layout/app-shell";
import { JobsStatePanel } from "@/features/jobs/components/jobs-state-panel";

export default function JobDetailLoadingPage() {
  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <JobsStatePanel
          title="Loading job detail"
          description="Please wait while we load the selected job."
        />
      </main>
    </AppShell>
  );
}
