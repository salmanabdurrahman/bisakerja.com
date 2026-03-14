import { redirect } from "next/navigation";

import { AppShell } from "@/components/layout/app-shell";
import { AccountNotificationsClient } from "@/features/growth/components/account-notifications-client";
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
    <AppShell>
      <main className="grid gap-4" role="main">
        <div className="grid gap-1">
          <h2 className="text-xl font-semibold">Notifications</h2>
          <p className="text-sm text-gray-600">
            View notification history and mark updates as read.
          </p>
        </div>
        {renderNotificationsView(viewState)}
      </main>
    </AppShell>
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
      <section className="grid gap-3 rounded-lg border border-red-200 bg-red-50 p-4">
        <h3 className="text-lg font-semibold text-red-900">
          Failed to load notifications
        </h3>
        <p className="text-sm text-red-800">
          Notifications are currently unavailable. Please refresh the page.
        </p>
        <a
          href="/account/notifications"
          className="w-fit rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Try again
        </a>
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
