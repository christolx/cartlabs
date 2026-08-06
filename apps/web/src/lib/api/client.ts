import { connection } from "next/server";
import type { components } from "./schema";

export type ProductPage = components["schemas"]["ProductPage"];
export type ProductDetail = components["schemas"]["ProductDetail"];
export type Category = components["schemas"]["Category"];

export class APIError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
  }
}

function baseURL() {
  return (
    process.env.API_INTERNAL_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080/api/v1"
  ).replace(/\/$/, "");
}

export async function apiGet<T>(path: string): Promise<T> {
  await connection();
  const response = await fetch(`${baseURL()}${path}`, {
    headers: { Accept: "application/json" },
    cache: "no-store",
  });
  if (!response.ok) {
    let message = `API request failed with status ${response.status}`;
    try {
      const problem = (await response.json()) as { detail?: string };
      if (problem.detail) message = problem.detail;
    } catch {}
    throw new APIError(response.status, message);
  }
  return (await response.json()) as T;
}

export function formatMoney(value: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(value);
}
