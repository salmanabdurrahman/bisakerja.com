import { redirect } from "next/navigation";

import { ButtonLink } from "@/components/ui/button";
import { AccountSavedSearchesClient } from "@/features/growth/components/account-saved-searches-client";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
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
    <AccountDashboardShell
      eyebrow="Growth"
      title="Saved searches"
      description="Store high-signal query presets so you can reuse them in one click."
    >
      {renderSavedSearchesView(viewState)}
    </AccountDashboardShell>
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
      <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
        <h3 className="bk-heading-card text-red-900">
          Failed to load saved searches
        </h3>
        <p className="bk-body text-red-800">
          Saved searches are currently unavailable. Please refresh the page.
        </p>
        <ButtonLink href="/account/saved-searches" variant="danger">
          Try again
        </ButtonLink>
      </section>
    );
  }

  return (
    <AccountSavedSearchesClient
      initialSavedSearches={viewState.savedSearches}
    />
  );
}
