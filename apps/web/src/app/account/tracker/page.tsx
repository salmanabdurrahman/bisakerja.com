import { redirect } from "next/navigation";

import { ButtonLink } from "@/components/ui/button";
import { AccountTrackerClient } from "@/features/tracker/components/account-tracker-client";
import { AccountDashboardShell } from "@/features/profile/components/account-dashboard-shell";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, type SubscriptionState } from "@/services/auth";
import { getBillingStatus } from "@/services/billing";
import { getJobDetail } from "@/services/jobs";
import { listBookmarks, listTrackedApplications } from "@/services/tracker";
import type {
  EnrichedBookmark,
  EnrichedApplication,
} from "@/features/tracker/components/account-tracker-client";

type TrackerPageViewState =
  | {
      kind: "ready";
      bookmarks: EnrichedBookmark[];
      applications: EnrichedApplication[];
      subscriptionState: SubscriptionState | "status_unavailable";
    }
  | { kind: "error" };

export default async function AccountTrackerPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/tracker"));
  }

  const viewState = await loadTrackerViewState(accessToken);

  return (
    <AccountDashboardShell
      eyebrow="Tracker"
      title="Application tracker"
      description="Bookmark interesting jobs and track the status of your applications."
    >
      {renderTrackerView(viewState)}
    </AccountDashboardShell>
  );
}

async function loadTrackerViewState(
  accessToken: string,
): Promise<TrackerPageViewState> {
  try {
    const [bookmarksResponse, appsResponse] = await Promise.all([
      listBookmarks(accessToken),
      listTrackedApplications(accessToken),
    ]);

    // Enrich bookmarks and applications with job title/company in parallel
    const allJobIDs = [
      ...bookmarksResponse.data.map((b) => b.job_id),
      ...appsResponse.data.map((a) => a.job_id),
    ];
    const uniqueJobIDs = [...new Set(allJobIDs)];

    const jobDetailMap = new Map<string, { title: string; company: string }>();
    if (uniqueJobIDs.length > 0) {
      const results = await Promise.allSettled(
        uniqueJobIDs.map((id) => getJobDetail(id)),
      );
      for (let i = 0; i < uniqueJobIDs.length; i++) {
        const result = results[i];
        if (result.status === "fulfilled") {
          jobDetailMap.set(uniqueJobIDs[i], {
            title: result.value.data.title,
            company: result.value.data.company,
          });
        }
      }
    }

    const enrichedBookmarks: EnrichedBookmark[] = bookmarksResponse.data.map(
      (b) => ({
        ...b,
        job_title: jobDetailMap.get(b.job_id)?.title ?? null,
        job_company: jobDetailMap.get(b.job_id)?.company ?? null,
      }),
    );

    const enrichedApplications: EnrichedApplication[] = appsResponse.data.map(
      (a) => ({
        ...a,
        job_title: jobDetailMap.get(a.job_id)?.title ?? null,
        job_company: jobDetailMap.get(a.job_id)?.company ?? null,
      }),
    );

    let profileSubscriptionState: SubscriptionState | null = null;
    let fallbackPremium = false;

    try {
      const profileResponse = await getMe(accessToken);
      profileSubscriptionState =
        profileResponse.data.subscription_state ?? null;
      fallbackPremium = profileResponse.data.is_premium;
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        throw error;
      }
    }

    let subscriptionState: SubscriptionState | "status_unavailable" =
      profileSubscriptionState ?? (fallbackPremium ? "premium_active" : "free");

    try {
      const billingResponse = await getBillingStatus(accessToken);
      subscriptionState = billingResponse.data.subscription_state;
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        throw error;
      }
      if (!profileSubscriptionState && !fallbackPremium) {
        subscriptionState = "status_unavailable";
      }
    }

    return {
      kind: "ready",
      bookmarks: enrichedBookmarks,
      applications: enrichedApplications,
      subscriptionState,
    };
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/tracker"));
    }
    return { kind: "error" };
  }
}

function renderTrackerView(viewState: TrackerPageViewState) {
  if (viewState.kind === "error") {
    return (
      <section className="bk-card grid gap-3 border-red-200 bg-red-50 p-5">
        <h3 className="bk-heading-card text-red-900">
          Failed to load application tracker
        </h3>
        <p className="bk-body text-red-800">
          Application tracker is currently unavailable. Please refresh the page.
        </p>
        <ButtonLink href="/account/tracker" variant="danger">
          Try again
        </ButtonLink>
      </section>
    );
  }

  return (
    <AccountTrackerClient
      initialBookmarks={viewState.bookmarks}
      initialApplications={viewState.applications}
      subscriptionState={viewState.subscriptionState}
    />
  );
}
