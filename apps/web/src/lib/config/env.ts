const defaultAPIBaseURL = "http://localhost:8080/api/v1";

export function getAPIBaseURL(): string {
  const value = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  return value && value.length > 0 ? value : defaultAPIBaseURL;
}
