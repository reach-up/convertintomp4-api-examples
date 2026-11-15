"""1. Tasks and jobs: one request describes the whole pipeline.
A job is a set of named tasks. A task reads the output of another task through `input`."""
from common import SAMPLE_URL, api, print_tasks, wait_for_job

job = api("POST", "/v1/jobs", {
    "tag": "example-01",
    "tasks": {
        "import-logo": {"operation": "import/url", "url": SAMPLE_URL},
        "to-jpg": {"operation": "convert", "input": "import-logo", "output_format": "jpg", "options": {"background": "#ffffff"}},
        "export": {"operation": "export/url", "input": "to-jpg"},
    },
})
print("Created job", job["id"], "status", job["status"])
print_tasks(job)

# Block until the job settles (90 s per call, repeated). Example 02 polls instead, example 03 listens.
done = wait_for_job(job['id'])
print("Job", done["id"], done["status"])
print_tasks(done)
print("Download:", (done["tasks"]["export"].get("result") or {}).get("url"))
