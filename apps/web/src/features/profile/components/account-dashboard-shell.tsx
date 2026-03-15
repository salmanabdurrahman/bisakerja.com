import type { PropsWithChildren, ReactNode } from "react";

import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/ui/page-header";
import { AccountDashboardNav } from "@/features/profile/components/account-dashboard-nav";

interface AccountDashboardShellProps extends PropsWithChildren {
  title: string;
  description?: string;
  eyebrow?: string;
  actions?: ReactNode;
}

export function AccountDashboardShell({
  title,
  description,
  eyebrow,
  actions,
  children,
}: AccountDashboardShellProps) {
  return (
    <AppShell>
      <main className="grid gap-5" role="main">
        <PageHeader
          eyebrow={eyebrow}
          title={title}
          description={description}
          actions={actions}
        />
        <AccountDashboardNav />
        {children}
      </main>
    </AppShell>
  );
}
