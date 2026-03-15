import sanitizeHtml from "sanitize-html";

const htmlTagPattern = /<\/?[a-z][\s\S]*>/i;

const sanitizeOptions: sanitizeHtml.IOptions = {
  allowedTags: [
    "p",
    "br",
    "ul",
    "ol",
    "li",
    "strong",
    "em",
    "b",
    "i",
    "u",
    "a",
    "h2",
    "h3",
    "h4",
    "blockquote",
  ],
  allowedAttributes: {
    a: ["href", "target", "rel"],
  },
  allowedSchemes: ["http", "https", "mailto"],
  transformTags: {
    a: (_tagName, attributes) => ({
      tagName: "a",
      attribs: {
        ...(attributes.href ? { href: attributes.href } : {}),
        target: "_blank",
        rel: "nofollow noopener noreferrer",
      },
    }),
  },
};

/**
 * isHTMLDescription returns true when a job description contains HTML tags.
 */
export function isHTMLDescription(description: string): boolean {
  return htmlTagPattern.test(description);
}

/**
 * sanitizeJobDescription sanitizes rich-text HTML from external job sources.
 */
export function sanitizeJobDescription(description: string): string {
  const normalized = description.trim();
  if (!normalized) {
    return "";
  }
  if (!isHTMLDescription(normalized)) {
    return normalized;
  }
  return sanitizeHtml(normalized, sanitizeOptions).trim();
}
