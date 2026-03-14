"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuthSession } from "@/features/auth/session-provider";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";
import type { JobSource, SavedSearch } from "@/services/growth";
import type { NotificationAlertMode } from "@/services/preferences";

interface AccountSavedSearchesClientProps {
  initialSavedSearches: SavedSearch[];
}

const frequencyOptions: NotificationAlertMode[] = [
  "instant",
  "daily_digest",
  "weekly_digest",
];
const sourceOptions: JobSource[] = ["glints", "kalibrr", "jobstreet"];

export function AccountSavedSearchesClient({
  initialSavedSearches,
}: AccountSavedSearchesClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);

  const [savedSearches, setSavedSearches] = useState(initialSavedSearches);
  const [queryInput, setQueryInput] = useState("");
  const [locationInput, setLocationInput] = useState("");
  const [sourceInput, setSourceInput] = useState<JobSource | "">("");
  const [salaryMinInput, setSalaryMinInput] = useState("");
  const [frequencyInput, setFrequencyInput] =
    useState<NotificationAlertMode>("instant");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deletingID, setDeletingID] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  function handleUnauthorized() {
    clearBrowserSession();
    markAnonymous();
    router.replace(buildLoginHref("/account/saved-searches"));
  }

  async function handleCreateSavedSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatusMessage(null);

    const normalizedQuery = queryInput.trim();
    if (normalizedQuery.length < 2 || normalizedQuery.length > 200) {
      setStatusMessage("Query must be 2-200 characters.");
      return;
    }

    const salaryRaw = salaryMinInput.trim();
    let salaryMin: number | undefined;
    if (salaryRaw.length > 0) {
      const parsed = Number.parseInt(salaryRaw, 10);
      if (!Number.isFinite(parsed) || parsed < 0) {
        setStatusMessage("Salary minimum must be a number >= 0.");
        return;
      }
      salaryMin = parsed;
    }

    setIsSubmitting(true);
    try {
      const response = await sessionClient.createSavedSearch({
        query: normalizedQuery,
        location: locationInput.trim() || undefined,
        source: sourceInput || undefined,
        salary_min: salaryMin,
        frequency: frequencyInput,
        is_active: true,
      });
      setSavedSearches((previous) => [response.data, ...previous]);
      setQueryInput("");
      setLocationInput("");
      setSourceInput("");
      setSalaryMinInput("");
      setFrequencyInput("instant");
      setStatusMessage("Saved search added successfully.");
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to save saved search. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDeleteSavedSearch(id: string) {
    setStatusMessage(null);
    setDeletingID(id);
    try {
      await sessionClient.deleteSavedSearch(id);
      setSavedSearches((previous) => previous.filter((item) => item.id !== id));
      setStatusMessage("Saved search deleted successfully.");
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          handleUnauthorized();
          return;
        }
        setStatusMessage(error.message);
        return;
      }
      setStatusMessage("Failed to delete saved search.");
    } finally {
      setDeletingID(null);
    }
  }

  return (
    <section className="grid gap-4">
      <form
        onSubmit={handleCreateSavedSearch}
        className="grid gap-3 rounded-lg border border-gray-200 p-4"
        aria-label="Saved search form"
      >
        <h3 className="text-lg font-semibold text-gray-900">
          Add saved search
        </h3>

        <label className="grid gap-1 text-sm">
          <span className="font-medium text-gray-700">Query</span>
          <input
            type="text"
            value={queryInput}
            onChange={(event) => setQueryInput(event.target.value)}
            className="rounded-md border border-gray-300 px-3 py-2"
            placeholder="golang backend"
            required
          />
        </label>

        <div className="grid gap-3 md:grid-cols-2">
          <label className="grid gap-1 text-sm">
            <span className="font-medium text-gray-700">
              Location (optional)
            </span>
            <input
              type="text"
              value={locationInput}
              onChange={(event) => setLocationInput(event.target.value)}
              className="rounded-md border border-gray-300 px-3 py-2"
              placeholder="jakarta"
            />
          </label>

          <label className="grid gap-1 text-sm">
            <span className="font-medium text-gray-700">Source (optional)</span>
            <select
              value={sourceInput}
              onChange={(event) =>
                setSourceInput((event.target.value as JobSource | "") ?? "")
              }
              className="rounded-md border border-gray-300 px-3 py-2"
            >
              <option value="">All sources</option>
              {sourceOptions.map((source) => (
                <option key={source} value={source}>
                  {source}
                </option>
              ))}
            </select>
          </label>

          <label className="grid gap-1 text-sm">
            <span className="font-medium text-gray-700">
              Salary minimum (optional)
            </span>
            <input
              type="number"
              min={0}
              value={salaryMinInput}
              onChange={(event) => setSalaryMinInput(event.target.value)}
              className="rounded-md border border-gray-300 px-3 py-2"
              placeholder="12000000"
            />
          </label>

          <label className="grid gap-1 text-sm">
            <span className="font-medium text-gray-700">Alert frequency</span>
            <select
              value={frequencyInput}
              onChange={(event) =>
                setFrequencyInput(event.target.value as NotificationAlertMode)
              }
              className="rounded-md border border-gray-300 px-3 py-2"
            >
              {frequencyOptions.map((frequency) => (
                <option key={frequency} value={frequency}>
                  {frequency}
                </option>
              ))}
            </select>
          </label>
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-fit rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {isSubmitting ? "Saving..." : "Add saved search"}
        </button>

        {statusMessage ? (
          <p className="text-sm text-gray-700" role="status" aria-live="polite">
            {statusMessage}
          </p>
        ) : null}
      </form>

      <section className="grid gap-3 rounded-lg border border-gray-200 p-4">
        <h3 className="text-lg font-semibold text-gray-900">
          Saved searches list
        </h3>
        {savedSearches.length === 0 ? (
          <p className="text-sm text-gray-600">
            No saved searches yet. Add your first query.
          </p>
        ) : (
          <ul className="grid gap-2">
            {savedSearches.map((item) => (
              <li
                key={item.id}
                className="grid gap-2 rounded-md border border-gray-200 px-3 py-2 text-sm text-gray-700"
              >
                <p className="font-medium">{item.query}</p>
                <p className="text-xs text-gray-500">
                  {item.location
                    ? `Location: ${item.location}`
                    : "Location: all"}
                  {" · "}
                  {item.source ? `Source: ${item.source}` : "Source: all"}
                  {" · "}
                  Frequency: {item.frequency}
                </p>
                <div className="flex items-center gap-2">
                  <span className="rounded bg-gray-100 px-2 py-1 text-xs">
                    {item.is_active ? "active" : "inactive"}
                  </span>
                  <button
                    type="button"
                    onClick={() => void handleDeleteSavedSearch(item.id)}
                    disabled={deletingID === item.id}
                    className="rounded-md border border-red-300 px-2 py-1 text-xs text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {deletingID === item.id ? "Deleting..." : "Delete"}
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  );
}
