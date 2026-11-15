"""6. A complex job in one request:
  page 1: a capture of a web page as PDF
  page 2: an image converted to PDF
  merge both into one PDF, make a 200 px thumbnail of the image,
  pack the PDF and the thumbnail into a zip and export it."""
from common import SAMPLE_URL, api, print_tasks, wait_for_job

job = api("POST", "/v1/jobs", {"tag": "example-06", "tasks": {
    "page": {"operation": "capture-website", "url": "https://convertintomp4.com/api-docs", "output_format": "pdf",
             "options": {"page_format": "A4", "print_background": True}},
    "logo": {"operation": "import/url", "url": SAMPLE_URL},
    "logo-pdf": {"operation": "convert", "input": "logo", "output_format": "pdf"},
    "merged": {"operation": "merge", "input": ["page", "logo-pdf"], "output_format": "pdf"},
    "thumb": {"operation": "thumbnail", "input": "logo", "output_format": "jpg", "options": {"width": 200, "height": 200, "fit": "max"}},
    "package": {"operation": "archive/create", "input": ["merged", "thumb"], "output_format": "zip"},
    "export": {"operation": "export/url", "input": "package"},
}})
print("Created job", job["id"], "with", len(job["tasks"]), "tasks")

done = wait_for_job(job['id'])
print("Job", done["status"])
print_tasks(done)
if done["status"] == "completed":
    print("Zip:", done["tasks"]["export"]["result"]["url"])
