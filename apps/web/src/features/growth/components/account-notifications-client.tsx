"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuthSession } from "@/features/auth/session-provider";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import type { APIPagination } from "@/lib/types/api";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";
import type { NotificationRecord } from "@/services/growth";

interface AccountNotificationsClientProps {
  initialNotifications: NotificationRecord[];
  initialPagination: APIPagination | null;
}

const pageLimit = 20;

export function AccountNotificationsClient({
  initialNotifications,
  initialPagination,
}: AccountNotificationsClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);

  const [notifications, setNotifications] = useState(initialNotifications);
  const [pagination, setPagination] = useState<APIPagination | null>(
    initialPagination,
  );
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [markingID, setMarkingID] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  function handleUnauthorized() {
    clearBrowserSession();
    markAnonymous();
    router.replace(buildLoginHref("/account/notifications"));
  }

  async function loadNotifications(page: number, unreadFilter: boolean) {
    setStatusMessage(null);
    setIsLoading(true);
    try {
      const response = await sessionClient.listNotifications({
        page,
        limit: pageLimit,
        unread_only: unreadFilter,
      });
      setNotifications(response.data);
      setPagination(response.meta.pagination ?? null);
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to load notifications. Try again.");
    } finally {
      setIsLoading(false);
    }
  }

  async function handleMarkRead(notificationID: string) {
    setStatusMessage(null);
    setMarkingID(notificationID);
    try {
      const response =
        await sessionClient.markNotificationAsRead(notificationID);
      setNotifications((previous) =>
        previous.map((item) =>
          item.id === notificationID ? response.data : item,
        ),
      );
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to mark notification as read.");
    } finally {
      setMarkingID(null);
    }
  }

  const currentPage = pagination?.page ?? 1;
  const totalPages = pagination?.total_pages ?? 1;
  const canPrev = currentPage > 1;
  const canNext = currentPage < totalPages;

  return (
    <section className="grid gap-4 rounded-lg border border-gray-200 p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h3 className="text-lg font-semibold text-gray-900">
          Notification center
        </h3>
        <label className="inline-flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={unreadOnly}
            onChange={(event) => {
              const nextUnread = event.target.checked;
              setUnreadOnly(nextUnread);
              void loadNotifications(1, nextUnread);
            }}
          />
          <span>Unread only</span>
        </label>
      </div>

      {statusMessage ? (
        <p className="text-sm text-gray-700" role="status" aria-live="polite">
          {statusMessage}
        </p>
      ) : null}

      {notifications.length === 0 ? (
        <p className="text-sm text-gray-600">
          No notifications found for the current filter.
        </p>
      ) : (
        <ul className="grid gap-2">
          {notifications.map((notification) => {
            const isUnread = !notification.read_at;
            return (
              <li
                key={notification.id}
                className="grid gap-2 rounded-md border border-gray-200 px-3 py-2 text-sm"
              >
                <p className="font-medium text-gray-900">
                  Job ID: {notification.job_id}
                </p>
                <p className="text-gray-700">
                  Channel: {notification.channel} · Status:{" "}
                  {notification.status}
                </p>
                <p className="text-xs text-gray-500">
                  Sent at:{" "}
                  {new Date(notification.sent_at).toLocaleString("id-ID")}
                </p>
                <p className="text-xs text-gray-500">
                  {notification.read_at
                    ? `Read at: ${new Date(notification.read_at).toLocaleString(
                        "id-ID",
                      )}`
                    : "Unread"}
                </p>
                {isUnread ? (
                  <button
                    type="button"
                    onClick={() => void handleMarkRead(notification.id)}
                    disabled={markingID === notification.id}
                    className="w-fit rounded-md border border-blue-300 px-2 py-1 text-xs text-blue-700 hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {markingID === notification.id
                      ? "Menyimpan..."
                      : "Mark as read"}
                  </button>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}

      <div className="flex items-center gap-3">
        <button
          type="button"
          disabled={!canPrev || isLoading}
          onClick={() => void loadNotifications(currentPage - 1, unreadOnly)}
          className="rounded-md border border-gray-300 px-3 py-1 text-sm disabled:cursor-not-allowed disabled:opacity-60"
        >
          Prev
        </button>
        <p className="text-sm text-gray-600">
          Page {currentPage} / {totalPages}
        </p>
        <button
          type="button"
          disabled={!canNext || isLoading}
          onClick={() => void loadNotifications(currentPage + 1, unreadOnly)}
          className="rounded-md border border-gray-300 px-3 py-1 text-sm disabled:cursor-not-allowed disabled:opacity-60"
        >
          Next
        </button>
      </div>
    </section>
  );
}
