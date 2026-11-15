// 5. Results and errors.
// Results: an export/url task returns { url, file: { name, size } }; the URL is signed and expires.
// Errors: { success: false, error: { code, message, details? } } with the HTTP status.
import { writeFileSync } from "node:fs";
import { api, ApiError, sleep, SAMPLE_URL, waitForJob } from "./common.mjs";

// Retry only what can succeed later: rate limits and server errors. Never retry a 4xx input error.
async function withRetry(fn, attempts = 4) {
  for (let i = 1; ; i++) {
    try {
      return await fn();
    } catch (e) {
      const retryable = e instanceof ApiError && (e.status === 429 || e.status >= 500);
      if (!retryable || i === attempts) throw e;
      const wait = e.retryAfter ? Number(e.retryAfter) * 1000 : 2 ** i * 1000;
      console.log(`  ${e.code}, retrying in ${wait} ms`);
      await sleep(wait);
    }
  }
}

// An input error: the task names an input that does not exist.
try {
  await api("POST", "/v1/jobs", { tasks: { "to-jpg": { operation: "convert", input: "nothing", output_format: "jpg" } } });
} catch (e) {
  if (!(e instanceof ApiError)) throw e;
  console.log("Expected error:", e.message);
}

// A job that succeeds: download the result.
const job = await withRetry(() =>
  api("POST", "/v1/jobs", {
    tasks: {
      "import-logo": { operation: "import/url", url: SAMPLE_URL },
      "to-png": { operation: "thumbnail", input: "import-logo", output_format: "png", options: { width: 64, height: 64, fit: "crop" } },
      "export": { operation: "export/url", input: "to-png" },
    },
  }),
);
const done = await withRetry(() => waitForJob(job.id));
if (done.status !== "completed") {
  // A failed task carries its own error; tasks that depended on it are cancelled.
  for (const [name, t] of Object.entries(done.tasks)) if (t.error) console.log(`Task ${name} failed: ${t.error}`);
  process.exit(1);
}
const { url, file } = done.tasks.export.result;
const res = await fetch(url);
if (!res.ok) throw new Error(`download failed: ${res.status}`);
const out = file?.name || "result.png";
writeFileSync(out, Buffer.from(await res.arrayBuffer()));
console.log("Saved", out, file?.size, "bytes");
