"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { useAuthSession } from "@/features/auth/session-provider";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";
import { listJobs } from "@/services/jobs";
import type { JobListItem } from "@/services/jobs";
import type { SubscriptionState } from "@/services/auth";
import type {
  Bookmark,
  TrackedApplication,
  ApplicationStatus,
} from "@/services/tracker";

export interface EnrichedBookmark extends Bookmark {
  job_title: string | null;
  job_company: string | null;
}

export interface EnrichedApplication extends TrackedApplication {
  job_title: string | null;
  job_company: string | null;
}

interface AccountTrackerClientProps {
  initialBookmarks: EnrichedBookmark[];
  initialApplications: EnrichedApplication[];
  subscriptionState: SubscriptionState | "status_unavailable";
}

const FREE_TIER_APPLICATION_LIMIT = 5;
const APPLICATION_STATUSES: ApplicationStatus[] = [
  "applied",
  "interview",
  "offer",
  "rejected",
  "withdrawn",
];

function statusBadgeClass(status: ApplicationStatus): string {
  switch (status) {
    case "applied":
      return "bg-blue-50 border-blue-200 text-blue-800";
    case "interview":
      return "bg-yellow-50 border-yellow-200 text-yellow-800";
    case "offer":
      return "bg-green-50 border-green-200 text-green-800";
    case "rejected":
      return "bg-red-50 border-red-200 text-red-800";
    case "withdrawn":
      return "bg-slate-50 border-slate-200 text-slate-600";
  }
}

export function AccountTrackerClient({
  initialBookmarks,
  initialApplications,
  subscriptionState,
}: AccountTrackerClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);

  const [bookmarks, setBookmarks] = useState(initialBookmarks);
  const [applications, setApplications] = useState(initialApplications);

  // Job title autocomplete state for Track New Application
  const [titleQuery, setTitleQuery] = useState("");
  const [titleResults, setTitleResults] = useState<JobListItem[]>([]);
  const [selectedTitle, setSelectedTitle] = useState("");
  const [selectedJobID, setSelectedJobID] = useState("");
  const [notesInput, setNotesInput] = useState("");

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deletingBookmarkID, setDeletingBookmarkID] = useState<string | null>(
    null,
  );
  const [updatingAppID, setUpdatingAppID] = useState<string | null>(null);
  const [deletingAppID, setDeletingAppID] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  // Debounced job title search — identical pattern to AI tools
  useEffect(() => {
    if (titleQuery.length < 2) {
      setTitleResults([]);
      if (titleQuery === "") {
        setSelectedJobID("");
        setSelectedTitle("");
      }
      return;
    }
    if (titleQuery === selectedTitle) {
      return;
    }

    const timeoutId = setTimeout(() => {
      listJobs({ q: titleQuery, limit: 10 })
        .then((res) => {
          setTitleResults(res.data);
        })
        .catch(() => {
          setTitleResults([]);
        });
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [titleQuery, selectedTitle]);

  function handleUnauthorized() {
    clearBrowserSession();
    markAnonymous();
    router.replace(buildLoginHref("/account/tracker"));
  }

  const activeApplications = applications.filter(
    (app) => app.status !== "rejected" && app.status !== "withdrawn",
  ).length;

  const isFreeTier = subscriptionState === "free";
  const limitReached =
    isFreeTier && activeApplications >= FREE_TIER_APPLICATION_LIMIT;

  async function handleDeleteBookmark(jobID: string) {
    setStatusMessage(null);
    setDeletingBookmarkID(jobID);
    try {
      await sessionClient.deleteBookmark(jobID);
      setBookmarks((prev) => prev.filter((b) => b.job_id !== jobID));
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to delete bookmark.");
    } finally {
      setDeletingBookmarkID(null);
    }
  }

  async function handleCreateApplication(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setStatusMessage(null);
    if (limitReached) return;

    const trimmedJobId = selectedJobID.trim();
    if (!trimmedJobId) {
      setStatusMessage("Please select a job from the search results.");
      return;
    }

    setIsSubmitting(true);
    try {
      const response = await sessionClient.createTrackedApplication({
        job_id: trimmedJobId,
        notes: notesInput.trim() || undefined,
      });
      // Enrich the new application with the title we already know from autocomplete
      const enriched: EnrichedApplication = {
        ...response.data,
        job_title: selectedTitle || null,
        job_company: null,
      };
      setApplications((prev) => [enriched, ...prev]);
      setTitleQuery("");
      setSelectedTitle("");
      setSelectedJobID("");
      setTitleResults([]);
      setNotesInput("");
      setStatusMessage("Application tracked successfully.");
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to track application.");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleUpdateApplicationStatus(
    id: string,
    newStatus: ApplicationStatus,
  ) {
    setStatusMessage(null);
    setUpdatingAppID(id);
    try {
      const response = await sessionClient.updateApplicationStatus(id, {
        status: newStatus,
      });
      setApplications((prev) =>
        prev.map((app) =>
          app.id === id
            ? {
                ...response.data,
                job_title: app.job_title,
                job_company: app.job_company,
              }
            : app,
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
      setStatusMessage("Failed to update status.");
    } finally {
      setUpdatingAppID(null);
    }
  }

  async function handleDeleteApplication(id: string) {
    setStatusMessage(null);
    setDeletingAppID(id);
    try {
      await sessionClient.deleteTrackedApplication(id);
      setApplications((prev) => prev.filter((app) => app.id !== id));
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to delete tracked application.");
    } finally {
      setDeletingAppID(null);
    }
  }

  return (
    <div className="grid gap-4">
      {/* Section A: Bookmarks */}
      <section className="bk-card grid gap-3 p-5">
        <h3 className="bk-heading-card">Bookmarks</h3>
        {bookmarks.length === 0 ? (
          <p className="bk-body">No bookmarks yet.</p>
        ) : (
          <ul className="grid gap-2">
            {bookmarks.map((b) => (
              <li
                key={b.job_id}
                className="flex items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-3 py-2"
              >
                <div>
                  <p className="bk-body font-medium text-black">
                    {b.job_title ?? b.job_id}
                  </p>
                  {b.job_company ? (
                    <p className="bk-body-sm text-[#555555]">{b.job_company}</p>
                  ) : null}
                  <p className="bk-body-sm text-[#777777]">
                    Bookmarked on: {new Date(b.created_at).toLocaleDateString()}
                  </p>
                </div>
                <Button
                  type="button"
                  variant="danger"
                  size="sm"
                  disabled={deletingBookmarkID === b.job_id}
                  onClick={() => void handleDeleteBookmark(b.job_id)}
                >
                  {deletingBookmarkID === b.job_id ? "Deleting..." : "Delete"}
                </Button>
              </li>
            ))}
          </ul>
        )}
      </section>

      {/* Section B: Track new application */}
      <section className="bk-card grid gap-3 p-5">
        <h3 className="bk-heading-card">Track new application</h3>
        {limitReached ? (
          <p className="rounded-xl border border-yellow-200 bg-yellow-50 px-3 py-2 bk-body text-yellow-800">
            Free tier is limited to {FREE_TIER_APPLICATION_LIMIT} active
            applications. Upgrade to premium to track more.
          </p>
        ) : (
          <form onSubmit={handleCreateApplication} className="grid gap-3">
            <div className="relative">
              <label className="grid gap-1 bk-label">
                <span className="font-medium text-slate-700">Job title</span>
                <input
                  type="text"
                  required
                  className="bk-input"
                  placeholder="Search job title..."
                  value={titleQuery}
                  onChange={(e) => setTitleQuery(e.target.value)}
                />
              </label>
              {titleResults.length > 0 && titleQuery !== selectedTitle ? (
                <ul className="absolute z-10 mt-1 max-h-60 overflow-y-auto w-full rounded-xl border border-[#E5E5E5] bg-white shadow-md">
                  {titleResults.map((result) => (
                    <li
                      key={result.id}
                      className="cursor-pointer px-3 py-2 bk-body text-[#444444] hover:bg-[#F9F9F9]"
                      onClick={() => {
                        setSelectedJobID(result.id);
                        setSelectedTitle(result.title);
                        setTitleQuery(result.title);
                        setTitleResults([]);
                      }}
                    >
                      <span className="font-medium">{result.title}</span>
                      {result.company ? (
                        <span className="bk-body-sm text-[#777777] ml-2">
                          · {result.company}
                        </span>
                      ) : null}
                    </li>
                  ))}
                </ul>
              ) : null}
            </div>
            <label className="grid gap-1 bk-label">
              <span className="font-medium text-slate-700">
                Notes (optional)
              </span>
              <input
                type="text"
                className="bk-input"
                placeholder="e.g. Applied via LinkedIn"
                value={notesInput}
                onChange={(e) => setNotesInput(e.target.value)}
              />
            </label>
            <div>
              <Button type="submit" variant="secondary" disabled={isSubmitting}>
                {isSubmitting ? "Tracking..." : "Track application"}
              </Button>
            </div>
            {statusMessage ? (
              <p
                className="rounded-xl border border-[#E5E5E5] bg-[#F9F9F9] px-3 py-2 bk-body text-[#555555]"
                role="status"
                aria-live="polite"
              >
                {statusMessage}
              </p>
            ) : null}
          </form>
        )}
      </section>

      {/* Section C: Tracked applications */}
      <section className="bk-card grid gap-3 p-5">
        <h3 className="bk-heading-card flex items-center justify-between">
          <span>Tracked applications</span>
          {isFreeTier && (
            <span className="text-sm font-normal text-slate-500">
              {activeApplications}/{FREE_TIER_APPLICATION_LIMIT} active
              applications used
            </span>
          )}
        </h3>
        {applications.length === 0 ? (
          <p className="bk-body">No tracked applications yet.</p>
        ) : (
          <ul className="grid gap-2">
            {applications.map((app) => (
              <li
                key={app.id}
                className="grid gap-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-3"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="bk-body font-medium text-black">
                      {app.job_title ?? app.job_id}
                    </p>
                    {app.job_company ? (
                      <p className="bk-body-sm text-[#555555]">
                        {app.job_company}
                      </p>
                    ) : null}
                  </div>
                  <span
                    className={`px-2 py-1 rounded-full border bk-body-sm font-medium ${statusBadgeClass(app.status)}`}
                  >
                    {app.status}
                  </span>
                </div>
                {app.notes && (
                  <p className="bk-body-sm text-[#555555]">{app.notes}</p>
                )}
                <div className="flex items-center justify-between text-[#777777] bk-body-sm">
                  <span>
                    Tracked: {new Date(app.created_at).toLocaleDateString()}
                  </span>
                  <div className="flex items-center gap-2">
                    <select
                      className="bk-select text-sm py-1 h-auto"
                      value={app.status}
                      disabled={updatingAppID === app.id}
                      onChange={(e) =>
                        void handleUpdateApplicationStatus(
                          app.id,
                          e.target.value as ApplicationStatus,
                        )
                      }
                    >
                      {APPLICATION_STATUSES.map((status) => (
                        <option key={status} value={status}>
                          {status}
                        </option>
                      ))}
                    </select>
                    <Button
                      type="button"
                      variant="danger"
                      size="sm"
                      disabled={deletingAppID === app.id}
                      onClick={() => void handleDeleteApplication(app.id)}
                    >
                      {deletingAppID === app.id ? "Deleting..." : "Delete"}
                    </Button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
