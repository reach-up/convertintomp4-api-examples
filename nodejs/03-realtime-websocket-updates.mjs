// 3. Realtime updates over a WebSocket instead of polling.
// wss://api.convertintomp4.com/ws/user/jobs streams every job of the key's account:
// {type: "started" | "progress" | "completed" | "failed", jobId, progress?, outputUrl?, error?}.
// For multi-task jobs, also pass `webhook_url` when you create the job: you get one signed
// job.finished / job.failed POST when the whole job settles (see verifyWebhook below).
import WebSocket from "ws";
import { createHmac, timingSafeEqual } from "node:crypto";
import { readFileSync } from "node:fs";
import { API_URL, api, authHeaders, SAMPLE_FILE } from "./common.mjs";

const ws = new WebSocket(API_URL.replace(/^http/, "ws") + "/ws/user/jobs", { headers: authHeaders() });
let jobId = null;

ws.on("message", (raw) => {
  const msg = JSON.parse(raw.toString());
  if (msg.type === "connected") return;
  if (jobId && msg.jobId !== jobId) return; // other jobs of the same account
  console.log(`  ${msg.type}${msg.progress !== undefined ? " " + msg.progress + "%" : ""}`);
  if (msg.type === "completed") {
    console.log("Download:", API_URL + msg.outputUrl);
    ws.close();
  }
  if (msg.type === "failed") {
    console.log("Failed:", msg.error);
    ws.close();
  }
});

ws.on("open", async () => {
  // Start the conversion only once the socket is open, so no event is missed.
  const form = new FormData();
  form.append("file", new Blob([readFileSync(SAMPLE_FILE)]), "sample.png");
  form.append("targetFormat", "jpg");
  const created = await api("POST", "/v1/convert", form);
  jobId = created.jobId;
  console.log("Created conversion", jobId);
});

// Webhook receivers: check the signature on the raw body before trusting it.
// The secret comes from GET /v1/jobs/webhook-secret.
export function verifyWebhook(rawBody, signatureHeader, secret) {
  const expected = "sha256=" + createHmac("sha256", secret).update(rawBody).digest("hex");
  const a = Buffer.from(expected);
  const b = Buffer.from(signatureHeader ?? "");
  return a.length === b.length && timingSafeEqual(a, b);
}
