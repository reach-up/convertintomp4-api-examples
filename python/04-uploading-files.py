"""4. Uploading files.
A) One call: POST /v1/convert with the file as multipart form data.
B) In a job: an import/upload task returns a form URL; send the file there and the job continues."""
from urllib.parse import urlparse

from common import SAMPLE_FILE, api, print_tasks, wait_for_job

# A) one call
file_bytes = SAMPLE_FILE.read_bytes()  # bytes, not a stream: a retried request re-sends the whole file
conversion = api("POST", "/v1/convert", files={"file": ("sample.png", file_bytes)}, data={"targetFormat": "webp"})
print("A) conversion", conversion["jobId"], conversion["status"], "status at", conversion["statusUrl"])

# B) upload inside a job
job = api("POST", "/v1/jobs", {"tasks": {
    "upload": {"operation": "import/upload", "filename": "sample.png"},
    "to-jpg": {"operation": "convert", "input": "upload", "output_format": "jpg"},
    "export": {"operation": "export/url", "input": "to-jpg"},
}})
form = job["tasks"]["upload"]["result"]["form"]  # {"url", "method": "POST", "field": "file"}
api(form["method"], urlparse(form["url"]).path, files={form["field"]: ("sample.png", file_bytes)})
print("B) uploaded to job", job["id"])

done = wait_for_job(job['id'])
print_tasks(done)
print("Download:", (done["tasks"]["export"].get("result") or {}).get("url"))
