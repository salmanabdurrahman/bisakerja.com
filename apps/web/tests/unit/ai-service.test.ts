import { afterEach, describe, expect, it, vi } from "vitest";

import {
  generateAICoverLetterDraft,
  generateAIJobFitSummary,
  generateAISearchAssistant,
  getAIUsage,
} from "@/services/ai";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

describe("ai services", () => {
  it("requests usage with feature query parameter", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: { code: 200, status: "success", message: "AI usage retrieved" },
          data: {
            feature: "search_assistant",
            tier: "free",
            daily_quota: 5,
            used: 1,
            remaining: 4,
            reset_at: "2030-01-01T00:00:00Z",
          },
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await getAIUsage("access-token", "job_fit_summary");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/ai/usage?feature=job_fit_summary",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );
  });

  it("posts search assistant payload", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          meta: { code: 200, status: "success", message: "AI generated" },
          data: {
            feature: "search_assistant",
            prompt: "remote backend",
            suggested_query: "remote backend golang",
            suggested_filters: {
              locations: [],
              job_types: [],
              salary_min: null,
            },
            summary: "Use remote-focused backend terms.",
            tier: "free",
            provider: "openai_compatible",
            model: "gpt-4.1-mini",
            daily_quota: 5,
            used_today: 1,
            quota_remaining: 4,
            reset_at: "2030-01-01T00:00:00Z",
          },
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await generateAISearchAssistant("access-token", {
      prompt: "remote backend",
      context: { location: "Jakarta" },
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/ai/search-assistant",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer access-token",
        }),
      }),
    );

    const [, requestConfig] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(JSON.parse(String(requestConfig.body))).toEqual({
      prompt: "remote backend",
      context: { location: "Jakarta" },
    });
  });

  it("posts premium ai payloads for job fit and cover letter", async () => {
    const fetchMock = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({
            meta: { code: 200, status: "success", message: "AI generated" },
            data: {
              feature: "job_fit_summary",
              job_id: "job_1",
              fit_score: 81,
              verdict: "strong_match",
              strengths: [],
              gaps: [],
              next_actions: [],
              summary: "Good fit.",
              tier: "premium",
              provider: "openai_compatible",
              model: "gpt-4.1-mini",
              daily_quota: 30,
              used_today: 2,
              quota_remaining: 28,
              reset_at: "2030-01-01T00:00:00Z",
            },
          }),
          { status: 200 },
        ),
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await generateAIJobFitSummary("access-token", {
      job_id: "job_1",
      focus: "architecture",
    });
    await generateAICoverLetterDraft("access-token", {
      job_id: "job_1",
      tone: "professional",
      highlights: ["Built scalable APIs"],
    });

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/ai/job-fit-summary",
      expect.objectContaining({
        method: "POST",
      }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/ai/cover-letter-draft",
      expect.objectContaining({
        method: "POST",
      }),
    );
  });
});
