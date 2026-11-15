// 6. A complex job in one request:
//   page 1: a capture of a web page as PDF
//   page 2: an image converted to PDF
//   merge both into one PDF, make a 200 px thumbnail of the image,
//   pack the PDF and the thumbnail into a zip and export it.
import { api, printTasks, SAMPLE_URL, waitForJob } from "./common.mjs";

const job = await api("POST", "/v1/jobs", {
  tag: "example-06",
  tasks: {
    "page": { operation: "capture-website", url: "https://convertintomp4.com/api-docs", output_format: "pdf", options: { page_format: "A4", print_background: true } },
    "logo": { operation: "import/url", url: SAMPLE_URL },
    "logo-pdf": { operation: "convert", input: "logo", output_format: "pdf" },
    "merged": { operation: "merge", input: ["page", "logo-pdf"], output_format: "pdf" },
    "thumb": { operation: "thumbnail", input: "logo", output_format: "jpg", options: { width: 200, height: 200, fit: "max" } },
    "package": { operation: "archive/create", input: ["merged", "thumb"], output_format: "zip" },
    "export": { operation: "export/url", input: "package" },
  },
});
console.log("Created job", job.id, "with", Object.keys(job.tasks).length, "tasks");

const done = await waitForJob(job.id);
console.log("Job", done.status);
printTasks(done);
if (done.status === "completed") console.log("Zip:", done.tasks.export.result.url);
