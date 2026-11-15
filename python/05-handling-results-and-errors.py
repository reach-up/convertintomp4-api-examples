"""5. Results and errors.
Results: an export/url task returns {"url", "file": {"name", "size"}}; the URL is signed and expires.
Errors: {"success": false, "error": {"code", "message", "details"?}} with the HTTP status."""
import sys
import time

import requests

from common import SAMPLE_URL, ApiError, api, wait_for_job


def with_retry(fn, attempts=4):
    """Retry only what can succeed later: rate limits and server errors. Never retry a 4xx input error."""
    for i in range(1, attempts + 1):
        try:
            return fn()
        except ApiError as e:
            if not (e.status == 429 or e.status >= 500) or i == attempts:
                raise
            wait = float(e.retry_after) if e.retry_after else 2 ** i
            print(f"  {e.code}, retrying in {wait} s")
            time.sleep(wait)


# An input error: the task names an input that does not exist.
try:
    api("POST", "/v1/jobs", {"tasks": {"to-jpg": {"operation": "convert", "input": "nothing", "output_format": "jpg"}}})
except ApiError as e:
    print("Expected error:", e)

# A job that succeeds: download the result.
job = with_retry(lambda: api("POST", "/v1/jobs", {"tasks": {
    "import-logo": {"operation": "import/url", "url": SAMPLE_URL},
    "to-png": {"operation": "thumbnail", "input": "import-logo", "output_format": "png", "options": {"width": 64, "height": 64, "fit": "crop"}},
    "export": {"operation": "export/url", "input": "to-png"},
}}))
done = with_retry(lambda: wait_for_job(job['id']))
if done["status"] != "completed":
    # A failed task carries its own error; tasks that depended on it are cancelled.
    for name, t in done["tasks"].items():
        if t.get("error"):
            print(f"Task {name} failed: {t['error']}")
    sys.exit(1)
result = done["tasks"]["export"]["result"]
res = requests.get(result["url"], timeout=120)
res.raise_for_status()
out = (result.get("file") or {}).get("name") or "result.png"
with open(out, "wb") as f:
    f.write(res.content)
print("Saved", out, len(res.content), "bytes")
