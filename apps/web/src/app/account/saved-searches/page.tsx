import { redirect } from "next/navigation";

import { AppShell } from "@/components/layout/app-shell";
import { AccountSavedSearchesClient } from "@/features/growth/components/account-saved-searches-client";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { listSavedSearches } from "@/services/growth";

type SavedSearchesPageViewState =
  | {
      kind: "ready";
      savedSearches: Awaited<ReturnType<typeof listSavedSearches>>["data"];
    }
  | { kind: "error" };

export default async function AccountSavedSearchesPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/saved-searches"));
  }

  const viewState = await loadSavedSearchesViewState(accessToken);

  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <div className="grid gap-1">
          <h2 className="text-xl font-semibold">Saved searches</h2>
          <p className="text-sm text-gray-600">
            Save favorite queries to speed up job discovery.
          </p>
        </div>
        {renderSavedSearchesView(viewState)}
      </main>
    </AppShell>
  );
}

async function loadSavedSearchesViewState(
  accessToken: string,
): Promise<SavedSearchesPageViewState> {
  try {
    const response = await listSavedSearches(accessToken);
    return {
      kind: "ready",
      savedSearches: response.data,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/saved-searches"));
    }
    return { kind: "error" };
  }
}

function renderSavedSearchesView(viewState: SavedSearchesPageViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="grid gap-3 rounded-lg border border-red-200 bg-red-50 p-4">
        <h3 className="text-lg font-semibold text-red-900">
          Failed to load saved searches
        </h3>
        <p className="text-sm text-red-800">
          Saved searches are currently unavailable. Please refresh
          the page.
        </p>
        <a
          href="/account/saved-searches"
          className="w-fit rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
        >
          Try again
        </a>
      </section>
    );
  }

  return (
    <AccountSavedSearchesClient
      initialSavedSearches={viewState.savedSearches}
    />
  );
}
