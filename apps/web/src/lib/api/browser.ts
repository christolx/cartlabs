import type { components } from "./schema";

export type Problem = components["schemas"]["Problem"];

export class BrowserAPIError extends Error {
  constructor(
    public readonly status: number,
    message: string,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = "BrowserAPIError";
  }
}

export async function browserRequest<T>(
  path: string,
  token?: string,
  init?: RequestInit,
): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init?.body && !headers.has("Content-Type"))
    headers.set("Content-Type", "application/json");

  let response: Response;
  try {
    response = await fetch(`/api/backend${path}`, {
      ...init,
      headers,
      credentials: "same-origin",
    });
  } catch {
    throw new BrowserAPIError(
      503,
      "API unavailable. Check service connection and try again.",
    );
  }

  if (!response.ok) {
    const problem = (await response.json().catch(() => null)) as Problem | null;
    throw new BrowserAPIError(
      response.status,
      problem?.detail ?? `Request failed with status ${response.status}`,
      problem?.requestId ?? response.headers.get("x-request-id") ?? undefined,
    );
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

export function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
