import type { APIErrorItem, APIErrorResponse, APIResponse } from "@/lib/types/api";

/**
 * APIRequestError encapsulates api request error.
 */
export class APIRequestError extends Error {
  status: number;
  code?: string;
  errors: APIErrorItem[];

  constructor(
    message: string,
    status: number,
    code?: string,
    errors: APIErrorItem[] = [],
  ) {
    super(message);
    this.name = "APIRequestError";
    this.status = status;
    this.code = code;
    this.errors = errors;
  }
}

/**
 * fetchJSON fetches json.
 */
export async function fetchJSON<T>(
  input: string,
  init?: RequestInit,
): Promise<APIResponse<T>> {
  const response = await fetch(input, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (!response.ok) {
    let payload: APIErrorResponse | null = null;
    try {
      payload = (await response.json()) as APIErrorResponse;
    } catch {
      payload = null;
    }

    const fallbackMessage = `Request failed with status ${response.status}`;
    const message = payload?.meta.message?.trim() || fallbackMessage;
    const code = payload?.errors?.[0]?.code;
    throw new APIRequestError(message, response.status, code, payload?.errors ?? []);
  }

  return (await response.json()) as APIResponse<T>;
}
