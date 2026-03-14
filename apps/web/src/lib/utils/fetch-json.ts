import type { APIResponse } from "@/lib/types/api";

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
    throw new Error(`Request failed with status ${response.status}`);
  }

  return (await response.json()) as APIResponse<T>;
}
