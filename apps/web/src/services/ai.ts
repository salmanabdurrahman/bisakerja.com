import type { APIResponse } from "@/lib/types/api";
import { fetchJSON } from "@/lib/utils/fetch-json";
import { buildAPIURL } from "@/services/http-client";

/**
 * AIFeature defines supported AI feature keys.
 */
export type AIFeature =
  | "search_assistant"
  | "job_fit_summary"
  | "cover_letter_draft";

/**
 * AITier defines user tier for AI capability.
 */
export type AITier = "free" | "premium";

/**
 * AIUsage defines usage/quota snapshot for one feature.
 */
export interface AIUsage {
  feature: AIFeature;
  tier: AITier;
  daily_quota: number;
  used: number;
  remaining: number;
  reset_at: string;
}

/**
 * AISearchAssistantContext defines optional context sent to AI search assistant.
 */
export interface AISearchAssistantContext {
  location?: string;
  job_types?: string[];
  salary_min?: number;
}

/**
 * GenerateAISearchAssistantInput defines input payload for search assistant.
 */
export interface GenerateAISearchAssistantInput {
  prompt: string;
  context?: AISearchAssistantContext;
}

/**
 * AISearchAssistantResult defines response payload for search assistant generation.
 */
export interface AISearchAssistantResult {
  feature: AIFeature;
  prompt: string;
  suggested_query: string;
  suggested_filters: {
    locations: string[];
    job_types: string[];
    salary_min?: number | null;
  };
  summary: string;
  tier: AITier;
  provider: string;
  model: string;
  daily_quota: number;
  used_today: number;
  quota_remaining: number;
  reset_at: string;
}

/**
 * GenerateAIJobFitSummaryInput defines input payload for job fit summary generation.
 */
export interface GenerateAIJobFitSummaryInput {
  job_id: string;
  focus?: string;
}

/**
 * AIJobFitSummaryResult defines response payload for job fit summary generation.
 */
export interface AIJobFitSummaryResult {
  feature: AIFeature;
  job_id: string;
  fit_score: number;
  verdict: "strong_match" | "moderate_match" | "low_match";
  strengths: string[];
  gaps: string[];
  next_actions: string[];
  summary: string;
  tier: AITier;
  provider: string;
  model: string;
  daily_quota: number;
  used_today: number;
  quota_remaining: number;
  reset_at: string;
}

/**
 * AICoverLetterTone defines supported cover letter tone values.
 */
export type AICoverLetterTone =
  | "professional"
  | "confident"
  | "friendly"
  | "concise";

/**
 * GenerateAICoverLetterDraftInput defines input payload for cover letter draft generation.
 */
export interface GenerateAICoverLetterDraftInput {
  job_id: string;
  tone?: AICoverLetterTone;
  highlights?: string[];
}

/**
 * AICoverLetterDraftResult defines response payload for cover letter draft generation.
 */
export interface AICoverLetterDraftResult {
  feature: AIFeature;
  job_id: string;
  tone: AICoverLetterTone;
  draft: string;
  key_points: string[];
  summary: string;
  tier: AITier;
  provider: string;
  model: string;
  daily_quota: number;
  used_today: number;
  quota_remaining: number;
  reset_at: string;
}

/**
 * getAIUsage retrieves usage snapshot for one AI feature.
 */
export async function getAIUsage(
  accessToken: string,
  feature?: AIFeature,
  init?: RequestInit,
): Promise<APIResponse<AIUsage>> {
  const params = new URLSearchParams();
  if (feature) {
    params.set("feature", feature);
  }
  const query = params.toString();
  const endpoint = query ? `/ai/usage?${query}` : "/ai/usage";

  return fetchJSON<AIUsage>(buildAPIURL(endpoint), {
    method: "GET",
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * generateAISearchAssistant generates AI search assistant result.
 */
export async function generateAISearchAssistant(
  accessToken: string,
  input: GenerateAISearchAssistantInput,
  init?: RequestInit,
): Promise<APIResponse<AISearchAssistantResult>> {
  return fetchJSON<AISearchAssistantResult>(
    buildAPIURL("/ai/search-assistant"),
    {
      method: "POST",
      body: JSON.stringify(input),
      cache: "no-store",
      ...init,
      headers: {
        Authorization: `Bearer ${accessToken}`,
        ...(init?.headers ?? {}),
      },
    },
  );
}

/**
 * generateAIJobFitSummary generates AI job fit summary result.
 */
export async function generateAIJobFitSummary(
  accessToken: string,
  input: GenerateAIJobFitSummaryInput,
  init?: RequestInit,
): Promise<APIResponse<AIJobFitSummaryResult>> {
  return fetchJSON<AIJobFitSummaryResult>(buildAPIURL("/ai/job-fit-summary"), {
    method: "POST",
    body: JSON.stringify(input),
    cache: "no-store",
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...(init?.headers ?? {}),
    },
  });
}

/**
 * generateAICoverLetterDraft generates AI cover letter draft result.
 */
export async function generateAICoverLetterDraft(
  accessToken: string,
  input: GenerateAICoverLetterDraftInput,
  init?: RequestInit,
): Promise<APIResponse<AICoverLetterDraftResult>> {
  return fetchJSON<AICoverLetterDraftResult>(
    buildAPIURL("/ai/cover-letter-draft"),
    {
      method: "POST",
      body: JSON.stringify(input),
      cache: "no-store",
      ...init,
      headers: {
        Authorization: `Bearer ${accessToken}`,
        ...(init?.headers ?? {}),
      },
    },
  );
}
