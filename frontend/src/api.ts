import type { Session, Verdict } from "./types";

type EvaluateResult = { product_name: string; verdict: Verdict; reasoning: string };

const API_URL = import.meta.env.VITE_API_URL;

async function loggedFetch(url: string, options: RequestInit) {
  const start = performance.now();
  const response = await fetch(url, options);
  const responseTimeSec = Math.round(((performance.now() - start) / 1000) * 100) / 100;
  console.log(`[api] ${options.method ?? "GET"} ${url} -> ${response.status} (${responseTimeSec.toFixed(2)}s)`);
  return { response, responseTimeSec };
}

export async function evaluate(product: string): Promise<EvaluateResult> {
  const { response } = await loggedFetch(`${API_URL}/evaluate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ product }),
  });

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    const err = new Error(body?.error ?? `Request failed with status ${response.status}`) as Error & { verdict?: string };
    if (body?.verdict) err.verdict = body.verdict;
    throw err;
  }

  return response.json();
}

export async function getSessions(): Promise<Session[]> {
  const { response } = await loggedFetch(`${API_URL}/sessions`, { method: "GET" });

  if (!response.ok) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  return response.json();
}
