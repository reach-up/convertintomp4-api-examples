// 1. Tasks and jobs: one request describes the whole pipeline.
// A job is a set of named tasks. A task reads the output of another task through `input`.
import { api, printTasks, SAMPLE_URL, waitForJob } from "./common.mjs";

const job = await api("POST", "/v1/jobs", {
  tag: "example-01",
  tasks: {
    "import-logo": { operation: "import/url", url: SAMPLE_URL },
    "to-jpg": { operation: "convert", input: "import-logo", output_format: "jpg", options: { background: "#ffffff" } },
    "export": { operation: "export/url", input: "to-jpg" },
  },
});
console.log("Created job", job.id, "status", job.status);
printTasks(job);

// Block until the job settles (90 s per call, repeated). Example 02 polls instead, example 03 listens.
const done = await waitForJob(job.id);
console.log("Job", done.id, done.status);
printTasks(done);
console.log("Download:", done.tasks.export.result?.url);
