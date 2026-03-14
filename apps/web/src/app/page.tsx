import { AppShell } from "@/components/layout/app-shell";
import { StateCard } from "@/components/ui/state-card";
import { buildAPIURL } from "@/services/http-client";

export default function Home() {
  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <StateCard
          title="Backend Connectivity"
          description={`API base URL: ${buildAPIURL("/")}`}
        />
        <StateCard
          title="Phase 0 Scope"
          description="App Router, domain-based structure, typed service layer, and baseline testing are now bootstrapped."
        />
      </main>
    </AppShell>
  );
}
