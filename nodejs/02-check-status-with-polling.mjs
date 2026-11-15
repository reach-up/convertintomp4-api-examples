// 2. Check status with polling: GET /v1/jobs/{id} until the job is completed, failed or cancelled.
import { api, printTasks, sleep, SAMPLE_URL } from "./common.mjs";

const job = await api("POST", "/v1/jobs", {
  tasks: {
    "import-logo": { operation: "import/url", url: SAMPLE_URL },
    "to-webp": { operation: "convert", input: "import-logo", output_format: "webp" },
    "export": { operation: "export/url", input: "to-webp" },
  },
});
console.log("Created job", job.id);

const deadline = Date.now() + 10 * 60_000;
let delay = 1000;
let current = job;
while (!["completed", "failed", "cancelled"].includes(current.status)) {
  if (Date.now() > deadline) throw new Error("gave up after 10 minutes");
  await sleep(delay);
  delay = Math.min(delay * 1.5, 10_000); // back off: most jobs finish in seconds, long videos take minutes
  current = await api("GET", `/v1/jobs/${job.id}`);
  const progress = Object.values(current.tasks).map((t) => `${t.progress ?? 0}%`).join(" ");
  console.log(`  ${current.status} ${progress}`);
}
printTasks(current);
if (current.status === "completed") console.log("Download:", current.tasks.export.result.url);
