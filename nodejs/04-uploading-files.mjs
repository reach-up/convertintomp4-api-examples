// 4. Uploading files.
// A) One call: POST /v1/convert with the file as multipart form data.
// B) In a job: an import/upload task returns a form URL; send the file there and the job continues.
import { readFileSync } from "node:fs";
import { api, printTasks, SAMPLE_FILE, waitForJob } from "./common.mjs";

const fileBytes = readFileSync(SAMPLE_FILE);

// A) one call
const form = new FormData();
form.append("file", new Blob([fileBytes]), "sample.png");
form.append("targetFormat", "webp");
const conversion = await api("POST", "/v1/convert", form);
console.log("A) conversion", conversion.jobId, conversion.status, "status at", conversion.statusUrl);

// B) upload inside a job
const job = await api("POST", "/v1/jobs", {
  tasks: {
    "upload": { operation: "import/upload", filename: "sample.png" },
    "to-jpg": { operation: "convert", input: "upload", output_format: "jpg" },
    "export": { operation: "export/url", input: "to-jpg" },
  },
});
const uploadForm = job.tasks.upload.result.form; // { url, method: "POST", field: "file" }
const body = new FormData();
body.append(uploadForm.field, new Blob([fileBytes]), "sample.png");
await api(uploadForm.method, new URL(uploadForm.url).pathname, body);
console.log("B) uploaded to job", job.id);

const done = await waitForJob(job.id);
printTasks(done);
console.log("Download:", done.tasks.export.result?.url);
