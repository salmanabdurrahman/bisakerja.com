import { describe, expect, it } from "vitest";

import {
  isHTMLDescription,
  sanitizeJobDescription,
} from "@/lib/utils/sanitize-job-description";

describe("sanitizeJobDescription", () => {
  it("keeps plain text unchanged", () => {
    const description = "Build reliable backend APIs";
    expect(isHTMLDescription(description)).toBe(false);
    expect(sanitizeJobDescription(description)).toBe(description);
  });

  it("removes unsafe tags and javascript links", () => {
    const description =
      '<p>Hello</p><script>alert("xss")</script><a href="javascript:alert(1)">Click</a>';

    const sanitized = sanitizeJobDescription(description);

    expect(sanitized).toContain("<p>Hello</p>");
    expect(sanitized).not.toContain("<script");
    expect(sanitized).not.toContain("javascript:");
    expect(sanitized).toContain('rel="nofollow noopener noreferrer"');
  });

  it("preserves common rich text tags", () => {
    const description =
      "<ul><li>Design APIs</li><li>Ship features</li></ul><p><strong>Fast feedback</strong></p>";

    const sanitized = sanitizeJobDescription(description);

    expect(sanitized).toContain("<ul>");
    expect(sanitized).toContain("<li>Design APIs</li>");
    expect(sanitized).toContain("<strong>Fast feedback</strong>");
  });
});
