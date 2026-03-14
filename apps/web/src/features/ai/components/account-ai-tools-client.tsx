"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { Button, ButtonLink } from "@/components/ui/button";
import { useAuthSession } from "@/features/auth/session-provider";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { clearBrowserSession } from "@/lib/auth/session-cookie";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { createSessionAPIClient } from "@/services/session-api-client";
import type {
  AICoverLetterDraftResult,
  AICoverLetterTone,
  AIFeature,
  AIJobFitSummaryResult,
  AISearchAssistantResult,
  AIUsage,
} from "@/services/ai";
import type { SubscriptionState } from "@/services/auth";

interface AccountAIToolsClientProps {
  subscriptionState: SubscriptionState | "status_unavailable";
  infoMessage: string | null;
}

const featureOrder: AIFeature[] = [
  "search_assistant",
  "job_fit_summary",
  "cover_letter_draft",
];

const featureLabels: Record<AIFeature, string> = {
  search_assistant: "Search assistant",
  job_fit_summary: "Job fit summary",
  cover_letter_draft: "Cover letter draft",
};

const toneOptions: AICoverLetterTone[] = [
  "professional",
  "confident",
  "friendly",
  "concise",
];

type UsageMap = Record<AIFeature, AIUsage | null>;

type ActionState =
  | "search_assistant"
  | "job_fit_summary"
  | "cover_letter_draft"
  | null;

export function AccountAIToolsClient({
  subscriptionState,
  infoMessage,
}: AccountAIToolsClientProps) {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();
  const sessionClient = useMemo(() => createSessionAPIClient(), []);

  const [usage, setUsage] = useState<UsageMap>({
    search_assistant: null,
    job_fit_summary: null,
    cover_letter_draft: null,
  });
  const [isUsageLoading, setIsUsageLoading] = useState(true);
  const [usageError, setUsageError] = useState<string | null>(null);

  const [activeAction, setActiveAction] = useState<ActionState>(null);
  const [searchMessage, setSearchMessage] = useState<string | null>(null);
  const [jobFitMessage, setJobFitMessage] = useState<string | null>(null);
  const [coverLetterMessage, setCoverLetterMessage] = useState<string | null>(
    null,
  );

  const [assistantPrompt, setAssistantPrompt] = useState("");
  const [assistantLocation, setAssistantLocation] = useState("");
  const [assistantJobTypes, setAssistantJobTypes] = useState("");
  const [assistantSalaryMin, setAssistantSalaryMin] = useState("");
  const [assistantResult, setAssistantResult] =
    useState<AISearchAssistantResult | null>(null);

  const [jobFitJobID, setJobFitJobID] = useState("");
  const [jobFitFocus, setJobFitFocus] = useState("");
  const [jobFitResult, setJobFitResult] =
    useState<AIJobFitSummaryResult | null>(null);

  const [coverLetterJobID, setCoverLetterJobID] = useState("");
  const [coverLetterTone, setCoverLetterTone] =
    useState<AICoverLetterTone>("professional");
  const [coverLetterHighlights, setCoverLetterHighlights] = useState("");
  const [coverLetterResult, setCoverLetterResult] =
    useState<AICoverLetterDraftResult | null>(null);

  const isPremiumActive = subscriptionState === "premium_active";

  useEffect(() => {
    void loadUsage();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function handleUnauthorized() {
    clearBrowserSession();
    markAnonymous();
    router.replace(buildLoginHref("/account/ai-tools"));
  }

  async function loadUsage() {
    setIsUsageLoading(true);
    setUsageError(null);
    try {
      const [searchUsage, fitUsage, coverUsage] = await Promise.all([
        sessionClient.getAIUsage("search_assistant"),
        sessionClient.getAIUsage("job_fit_summary"),
        sessionClient.getAIUsage("cover_letter_draft"),
      ]);
      setUsage({
        search_assistant: searchUsage.data,
        job_fit_summary: fitUsage.data,
        cover_letter_draft: coverUsage.data,
      });
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        handleUnauthorized();
        return;
      }
      setUsageError(resolveUsageErrorMessage(error));
    } finally {
      setIsUsageLoading(false);
    }
  }

  function applyQuotaFromResult(
    feature: AIFeature,
    result: {
      tier: "free" | "premium";
      daily_quota: number;
      used_today: number;
      quota_remaining: number;
      reset_at: string;
    },
  ) {
    setUsage((previous) => ({
      ...previous,
      [feature]: {
        feature,
        tier: result.tier,
        daily_quota: result.daily_quota,
        used: result.used_today,
        remaining: result.quota_remaining,
        reset_at: result.reset_at,
      },
    }));
  }

  async function handleSearchAssistantSubmit(
    event: FormEvent<HTMLFormElement>,
  ) {
    event.preventDefault();
    setSearchMessage(null);

    const prompt = assistantPrompt.trim();
    if (prompt.length < 5 || prompt.length > 500) {
      setSearchMessage("Prompt must be 5-500 characters.");
      return;
    }

    const normalizedSalary = assistantSalaryMin.trim();
    let salaryMin: number | undefined;
    if (normalizedSalary !== "") {
      const parsed = Number.parseInt(normalizedSalary, 10);
      if (!Number.isFinite(parsed) || parsed < 0) {
        setSearchMessage(
          "Salary minimum must be a number greater than or equal to 0.",
        );
        return;
      }
      salaryMin = parsed;
    }

    const parsedJobTypes = assistantJobTypes
      .split(",")
      .map((item) => item.trim().toLowerCase())
      .filter(Boolean);

    setActiveAction("search_assistant");
    try {
      const response = await sessionClient.generateAISearchAssistant({
        prompt,
        context: {
          location: assistantLocation.trim() || undefined,
          job_types: parsedJobTypes.length > 0 ? parsedJobTypes : undefined,
          salary_min: salaryMin,
        },
      });
      setAssistantResult(response.data);
      applyQuotaFromResult("search_assistant", response.data);
      setSearchMessage("Search assistant suggestion generated.");
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        handleUnauthorized();
        return;
      }
      setSearchMessage(resolveActionErrorMessage(error));
    } finally {
      setActiveAction(null);
    }
  }

  async function handleJobFitSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setJobFitMessage(null);

    const jobID = jobFitJobID.trim();
    if (jobID === "") {
      setJobFitMessage("Job ID is required.");
      return;
    }

    setActiveAction("job_fit_summary");
    try {
      const response = await sessionClient.generateAIJobFitSummary({
        job_id: jobID,
        focus: jobFitFocus.trim() || undefined,
      });
      setJobFitResult(response.data);
      applyQuotaFromResult("job_fit_summary", response.data);
      setJobFitMessage("Job fit summary generated.");
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        handleUnauthorized();
        return;
      }
      setJobFitMessage(resolveActionErrorMessage(error));
    } finally {
      setActiveAction(null);
    }
  }

  async function handleCoverLetterSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCoverLetterMessage(null);

    const jobID = coverLetterJobID.trim();
    if (jobID === "") {
      setCoverLetterMessage("Job ID is required.");
      return;
    }

    const parsedHighlights = coverLetterHighlights
      .split("\n")
      .map((item) => item.trim())
      .filter(Boolean)
      .slice(0, 5);

    setActiveAction("cover_letter_draft");
    try {
      const response = await sessionClient.generateAICoverLetterDraft({
        job_id: jobID,
        tone: coverLetterTone,
        highlights: parsedHighlights.length > 0 ? parsedHighlights : undefined,
      });
      setCoverLetterResult(response.data);
      applyQuotaFromResult("cover_letter_draft", response.data);
      setCoverLetterMessage("Cover letter draft generated.");
    } catch (error) {
      if (error instanceof APIRequestError && error.status === 401) {
        handleUnauthorized();
        return;
      }
      setCoverLetterMessage(resolveActionErrorMessage(error));
    } finally {
      setActiveAction(null);
    }
  }

  return (
    <section className="grid gap-5">
      {infoMessage ? (
        <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          {infoMessage}
        </p>
      ) : null}

      {!isPremiumActive ? (
        <section className="bk-card grid gap-3 border-[#E5E5E5] bg-[#F9F9F9] p-5 sm:p-6">
          <p className="text-sm font-medium text-black">
            Upgrade to premium to unlock job-fit insights and full cover letter
            drafting.
          </p>
          <div>
            <ButtonLink href="/pricing" variant="secondary" size="sm">
              View premium plans
            </ButtonLink>
          </div>
        </section>
      ) : null}

      <section className="bk-card grid gap-4 p-5 sm:p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h3 className="text-[24px] font-normal text-black">AI usage today</h3>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={isUsageLoading}
            onClick={() => void loadUsage()}
          >
            {isUsageLoading ? "Refreshing..." : "Refresh usage"}
          </Button>
        </div>

        {usageError ? (
          <p className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {usageError}
          </p>
        ) : null}

        <div className="grid gap-3 md:grid-cols-3">
          {featureOrder.map((feature) => (
            <UsageCard
              key={feature}
              label={featureLabels[feature]}
              usage={usage[feature]}
              isLoading={isUsageLoading}
            />
          ))}
        </div>
      </section>

      <section className="bk-card grid gap-4 p-5 sm:p-6">
        <h3 className="text-[24px] font-normal text-black">Search assistant</h3>
        <p className="text-[14px] text-[#666666]">
          Turn rough ideas into a better query and filter strategy before
          searching jobs.
        </p>
        <form
          onSubmit={handleSearchAssistantSubmit}
          className="grid gap-3"
          aria-label="Search assistant form"
        >
          <label className="grid gap-1 text-sm">
            <span className="font-medium text-slate-700">Prompt</span>
            <textarea
              value={assistantPrompt}
              onChange={(event) => setAssistantPrompt(event.target.value)}
              placeholder="I want remote Golang backend roles with strong API ownership."
              className="bk-textarea"
              required
            />
          </label>
          <div className="grid gap-3 md:grid-cols-3">
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">
                Location (optional)
              </span>
              <input
                value={assistantLocation}
                onChange={(event) => setAssistantLocation(event.target.value)}
                placeholder="Jakarta"
                className="bk-input"
              />
            </label>
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">
                Job types (optional)
              </span>
              <input
                value={assistantJobTypes}
                onChange={(event) => setAssistantJobTypes(event.target.value)}
                placeholder="fulltime, contract"
                className="bk-input"
              />
            </label>
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">
                Salary minimum (optional)
              </span>
              <input
                type="number"
                min={0}
                value={assistantSalaryMin}
                onChange={(event) => setAssistantSalaryMin(event.target.value)}
                placeholder="15000000"
                className="bk-input"
              />
            </label>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="submit"
              variant="secondary"
              disabled={activeAction === "search_assistant"}
            >
              {activeAction === "search_assistant"
                ? "Generating..."
                : "Generate suggestion"}
            </Button>
          </div>
          {searchMessage ? <StatusMessage text={searchMessage} /> : null}
        </form>

        {assistantResult ? (
          <div className="grid gap-2 rounded-2xl border border-[#E5E5E5] bg-[#F9F9F9] px-4 py-3 text-sm text-[#444444]">
            <p>
              Suggested query:{" "}
              <span className="font-medium text-black">
                {assistantResult.suggested_query}
              </span>
            </p>
            <p>{assistantResult.summary}</p>
            <p className="text-xs text-[#666666]">
              Filters: locations{" "}
              {assistantResult.suggested_filters.locations.length > 0
                ? assistantResult.suggested_filters.locations.join(", ")
                : "not specified"}
              {" · "}job types{" "}
              {assistantResult.suggested_filters.job_types.length > 0
                ? assistantResult.suggested_filters.job_types.join(", ")
                : "not specified"}
            </p>
          </div>
        ) : null}
      </section>

      <section className="bk-card grid gap-4 p-5 sm:p-6">
        <h3 className="text-[24px] font-normal text-black">Job fit summary</h3>
        <p className="text-[14px] text-[#666666]">
          Estimate how well your profile aligns with one job and get practical
          next steps.
        </p>
        <form
          onSubmit={handleJobFitSubmit}
          className="grid gap-3"
          aria-label="Job fit summary form"
        >
          <div className="grid gap-3 md:grid-cols-2">
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">Job ID</span>
              <input
                value={jobFitJobID}
                onChange={(event) => setJobFitJobID(event.target.value)}
                placeholder="job_123"
                className="bk-input"
                required
              />
            </label>
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">
                Focus (optional)
              </span>
              <input
                value={jobFitFocus}
                onChange={(event) => setJobFitFocus(event.target.value)}
                placeholder="Prioritize architecture depth and impact."
                className="bk-input"
              />
            </label>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="submit"
              variant="secondary"
              disabled={activeAction === "job_fit_summary"}
            >
              {activeAction === "job_fit_summary"
                ? "Generating..."
                : "Generate job fit"}
            </Button>
          </div>
          {jobFitMessage ? <StatusMessage text={jobFitMessage} /> : null}
        </form>

        {jobFitResult ? (
          <div className="grid gap-2 rounded-2xl border border-[#E5E5E5] bg-[#F9F9F9] px-4 py-3 text-sm text-[#444444]">
            <p>
              Fit score:{" "}
              <span className="font-medium text-black">
                {jobFitResult.fit_score}
              </span>
              {" · "}Verdict:{" "}
              <span className="font-medium text-black">
                {jobFitResult.verdict}
              </span>
            </p>
            <p>{jobFitResult.summary}</p>
            <p className="text-xs text-[#666666]">
              Strengths: {joinOrFallback(jobFitResult.strengths)}
            </p>
            <p className="text-xs text-[#666666]">
              Gaps: {joinOrFallback(jobFitResult.gaps)}
            </p>
            <p className="text-xs text-[#666666]">
              Next actions: {joinOrFallback(jobFitResult.next_actions)}
            </p>
          </div>
        ) : null}
      </section>

      <section className="bk-card grid gap-4 p-5 sm:p-6">
        <h3 className="text-[24px] font-normal text-black">
          Cover letter draft
        </h3>
        <p className="text-[14px] text-[#666666]">
          Generate a role-specific draft and key talking points that you can
          edit before sending.
        </p>
        <form
          onSubmit={handleCoverLetterSubmit}
          className="grid gap-3"
          aria-label="Cover letter draft form"
        >
          <div className="grid gap-3 md:grid-cols-2">
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">Job ID</span>
              <input
                value={coverLetterJobID}
                onChange={(event) => setCoverLetterJobID(event.target.value)}
                placeholder="job_123"
                className="bk-input"
                required
              />
            </label>
            <label className="grid gap-1 text-sm">
              <span className="font-medium text-slate-700">Tone</span>
              <select
                value={coverLetterTone}
                onChange={(event) =>
                  setCoverLetterTone(event.target.value as AICoverLetterTone)
                }
                className="bk-select"
              >
                {toneOptions.map((tone) => (
                  <option key={tone} value={tone}>
                    {tone}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <label className="grid gap-1 text-sm">
            <span className="font-medium text-slate-700">
              Highlights (optional, one per line)
            </span>
            <textarea
              value={coverLetterHighlights}
              onChange={(event) => setCoverLetterHighlights(event.target.value)}
              placeholder={
                "Led API reliability initiatives\nBuilt Golang services at scale"
              }
              className="bk-textarea"
            />
          </label>

          <div className="flex flex-wrap gap-2">
            <Button
              type="submit"
              variant="secondary"
              disabled={activeAction === "cover_letter_draft"}
            >
              {activeAction === "cover_letter_draft"
                ? "Generating..."
                : "Generate cover letter"}
            </Button>
          </div>
          {coverLetterMessage ? (
            <StatusMessage text={coverLetterMessage} />
          ) : null}
        </form>

        {coverLetterResult ? (
          <div className="grid gap-2 rounded-2xl border border-[#E5E5E5] bg-[#F9F9F9] px-4 py-3 text-sm text-[#444444]">
            <p className="text-xs text-[#666666]">
              Tone:{" "}
              <span className="font-medium text-black">
                {coverLetterResult.tone}
              </span>
            </p>
            <p className="whitespace-pre-wrap leading-relaxed">
              {coverLetterResult.draft}
            </p>
            <p className="text-xs text-[#666666]">
              Key points: {joinOrFallback(coverLetterResult.key_points)}
            </p>
          </div>
        ) : null}
      </section>
    </section>
  );
}

interface UsageCardProps {
  label: string;
  usage: AIUsage | null;
  isLoading: boolean;
}

function UsageCard({ label, usage, isLoading }: UsageCardProps) {
  return (
    <div className="grid gap-2 rounded-2xl border border-[#E5E5E5] bg-[#F9F9F9] px-4 py-3">
      <p className="text-[13px] font-medium text-black">{label}</p>
      {isLoading ? (
        <p className="text-sm text-[#666666]">Loading...</p>
      ) : usage ? (
        <>
          <p className="text-sm text-[#555555]">
            {usage.remaining} / {usage.daily_quota} remaining
          </p>
          <p className="text-xs text-[#777777]">Tier: {usage.tier}</p>
          <p className="text-xs text-[#777777]">
            Reset: {new Date(usage.reset_at).toLocaleString("en-US")}
          </p>
        </>
      ) : (
        <p className="text-sm text-[#666666]">No usage data.</p>
      )}
    </div>
  );
}

function StatusMessage({ text }: { text: string }) {
  return (
    <p
      className="rounded-xl border border-[#E5E5E5] bg-[#F9F9F9] px-3 py-2 text-sm text-[#555555]"
      role="status"
      aria-live="polite"
    >
      {text}
    </p>
  );
}

function joinOrFallback(items: string[]): string {
  if (items.length === 0) {
    return "No details provided.";
  }
  return items.join(" · ");
}

function resolveUsageErrorMessage(error: unknown): string {
  if (error instanceof APIRequestError) {
    return error.message;
  }
  return "Unable to load AI usage. Please refresh shortly.";
}

function resolveActionErrorMessage(error: unknown): string {
  if (error instanceof APIRequestError) {
    if (error.status === 403 || error.code === "FORBIDDEN") {
      return "This AI action is available for premium users only.";
    }
    if (error.code === "AI_QUOTA_EXCEEDED" || error.status === 429) {
      return "Your daily AI quota is exhausted. Please wait until reset.";
    }
    if (
      error.code === "AI_PROVIDER_RATE_LIMITED" ||
      error.code === "AI_PROVIDER_UNAVAILABLE"
    ) {
      return "AI provider is temporarily unavailable. Please retry in a moment.";
    }
    if (error.code === "AI_PROVIDER_UPSTREAM_ERROR") {
      return "AI response could not be processed. Please try again.";
    }
    if (error.code === "NOT_FOUND") {
      return "Job ID was not found. Please verify and try again.";
    }
    return error.message;
  }
  return "AI request failed. Please try again.";
}
