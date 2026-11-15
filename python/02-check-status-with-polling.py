"""2. Check status with polling: GET /v1/jobs/{id} until the job is completed, failed or cancelled."""
import time

from common import SAMPLE_URL, api, print_tasks

job = api("POST", "/v1/jobs", {"tasks": {
    "import-logo": {"operation": "import/url", "url": SAMPLE_URL},
    "to-webp": {"operation": "convert", "input": "import-logo", "output_format": "webp"},
    "export": {"operation": "export/url", "input": "to-webp"},
}})
print("Created job", job["id"])

deadline = time.time() + 600
delay = 1.0
current = job
while current["status"] not in ("completed", "failed", "cancelled"):
    if time.time() > deadline:
        raise SystemExit("gave up after 10 minutes")
    time.sleep(delay)
    delay = min(delay * 1.5, 10)  # back off: most jobs finish in seconds, long videos take minutes
    current = api("GET", f"/v1/jobs/{job['id']}")
    print(" ", current["status"], " ".join(f"{t.get('progress', 0)}%" for t in current["tasks"].values()))
print_tasks(current)
if current["status"] == "completed":
    print("Download:", current["tasks"]["export"]["result"]["url"])
