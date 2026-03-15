import { redirect } from "next/navigation";

import { ButtonLink } from "@/components/ui/button";
import { AccountNotificationsClient } from "@/features/growth/components/account-notifications-client";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import type { APIPagination } from "@/lib/types/api";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { listNotifications } from "@/services/growth";

type NotificationsPageViewState =
  | {
      kind: "ready";
      notifications: Awaited<ReturnType<typeof listNotifications>>["data"];
      pagination: APIPagination | null;
    }
  | { kind: "error" };

export default async function AccountNotificationsPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/notifications"));
  }

  const viewState = await loadNotificationsViewState(accessToken);

  return (
    <AccountDashboardShell
      eyebrow="Growth"
      title="Notifications"
      description="Review delivery history, focus unread updates, and close the loop with read status."
    >
      {renderNotificationsView(viewState)}
    </AccountDashboardShell>
  );
}

async function loadNotificationsViewState(
  accessToken: string,
): Promise<NotificationsPageViewState> {
  try {
    const response = await listNotifications(accessToken, {
      page: 1,
      limit: 20,
    });
    return {
      kind: "ready",
      notifications: response.data,
      pagination: response.meta.pagination ?? null,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/notifications"));
    }
    return { kind: "error" };
  }
}

function renderNotificationsView(viewState: NotificationsPageViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
        <h3 className="text-lg font-semibold text-red-900">
          Failed to load notifications
        </h3>
        <p className="text-sm text-red-800">
          Notifications are currently unavailable. Please refresh the page.
        </p>
        <ButtonLink href="/account/notifications" variant="danger">
          Try again
        </ButtonLink>
      </section>
    );
  }

  return (
    <AccountNotificationsClient
      initialNotifications={viewState.notifications}
      initialPagination={viewState.pagination}
    />
  );
}
