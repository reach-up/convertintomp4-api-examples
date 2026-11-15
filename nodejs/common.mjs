// Shared helpers for the ConvertIntoMP4 API examples.
// API reference: https://convertintomp4.com/api-docs
import { fileURLToPath } from "node:url";
import path from "node:path";

export const API_URL = (process.env.CIM4_API_URL || "https://api.convertintomp4.com").replace(/\/$/, "");
export const API_KEY = process.env.CIM4_API_KEY;
// Sandbox mode (X-Sandbox: true): files up to 1 MB, 100 requests a month per key. Set CIM4_SANDBOX=0 for live mode.
export const SANDBOX = process.env.CIM4_SANDBOX !== "0";
export const SAMPLE_URL = "https://convertintomp4.com/icon-192.png";
export const SAMPLE_FILE = path.join(path.dirname(fileURLToPath(import.meta.url)), "..", "fixtures", "sample.png");

if (!API_KEY) {
  console.error("Set CIM4_API_KEY to an API key from https://convertintomp4.com/account/api/keys");
  process.exit(1);
}

export function authHeaders(extra = {}) {
  return {
    "X-Api-Key": API_KEY,
    "User-Agent": "convertintomp4-api-examples/1.0 (node)",
    ...(SANDBOX ? { "X-Sandbox": "true" } : {}),
    ...extra,
  };
}

// Every error from /v1/jobs has one shape: { success: false, error: { code, message, details? } }
export class ApiError extends Error {
  constructor(status, code, message, details, retryAfter) {
    super(`${status} ${code}: ${message}`);
    Object.assign(this, { status, code, details, retryAfter });
  }
}

export async function api(method, urlPath, body, attempt = 1) {
  const isForm = body instanceof FormData;
  const res = await fetch(API_URL + urlPath, {
    method,
    headers: authHeaders(body && !isForm ? { "Content-Type": "application/json" } : {}),
    body: body === undefined ? undefined : isForm ? body : JSON.stringify(body),
  });
  const text = await res.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = { error: { code: "NOT_JSON", message: text.slice(0, 200) } };
  }
  if (res.status === 429 && attempt < 4) {
    // Rate limited (sandbox: 1 POST a second). Wait and send the same request again.
    await sleep(Number(res.headers.get("retry-after")) * 1000 || 1500 * attempt);
    return api(method, urlPath, body, attempt + 1);
  }
  if (!res.ok) {
    const err = data?.error ?? {};
    throw new ApiError(res.status, err.code ?? "HTTP_" + res.status, err.message ?? res.statusText, err.details, res.headers.get("retry-after"));
  }
  return data;
}

export const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

export function printTasks(job) {
  for (const [name, t] of Object.entries(job.tasks)) {
    console.log(`  ${name.padEnd(12)} ${t.operation.padEnd(16)} ${t.status}${t.error ? "  " + t.error : ""}`);
  }
}

// Wait for a job to settle. One /wait call holds at most 90 s; a longer job answers 408 TIMEOUT
// and we simply ask again.
export async function waitForJob(jobId, maxSeconds = 600) {
  const deadline = Date.now() + maxSeconds * 1000;
  for (;;) {
    try {
      return await api("GET", `/v1/jobs/${jobId}/wait?timeout=90`);
    } catch (e) {
      if (!(e instanceof ApiError && e.status === 408) || Date.now() > deadline) throw e;
    }
  }
}
