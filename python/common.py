"""Shared helpers for the ConvertIntoMP4 API examples. API reference: https://convertintomp4.com/api-docs"""
import os
import sys
import time
from pathlib import Path

import requests

API_URL = os.environ.get("CIM4_API_URL", "https://api.convertintomp4.com").rstrip("/")
API_KEY = os.environ.get("CIM4_API_KEY")
# Sandbox mode (X-Sandbox: true): files up to 1 MB, 100 requests a month per key. CIM4_SANDBOX=0 for live mode.
SANDBOX = os.environ.get("CIM4_SANDBOX") != "0"
SAMPLE_URL = "https://convertintomp4.com/icon-192.png"
SAMPLE_FILE = Path(__file__).resolve().parent.parent / "fixtures" / "sample.png"

if not API_KEY:
    sys.exit("Set CIM4_API_KEY to an API key from https://convertintomp4.com/account/api/keys")


def auth_headers():
    headers = {"X-Api-Key": API_KEY, "User-Agent": "convertintomp4-api-examples/1.0 (python)"}
    if SANDBOX:
        headers["X-Sandbox"] = "true"
    return headers


class ApiError(Exception):
    """Every /v1/jobs error has one shape: {"success": false, "error": {"code", "message", "details"?}}."""

    def __init__(self, status, code, message, details=None, retry_after=None):
        super().__init__(f"{status} {code}: {message}")
        self.status, self.code, self.details, self.retry_after = status, code, details, retry_after


def api(method, path, json_body=None, files=None, data=None, attempt=1):
    res = requests.request(method, API_URL + path, headers=auth_headers(), json=json_body, files=files, data=data, timeout=330)
    if res.status_code == 429 and attempt < 4:
        # Rate limited (sandbox: 1 POST a second). Wait and send the same request again.
        time.sleep(float(res.headers.get("Retry-After") or 1.5 * attempt))
        return api(method, path, json_body, files, data, attempt + 1)
    try:
        body = res.json() if res.content else None
    except ValueError:
        body = {"error": {"code": "NOT_JSON", "message": res.text[:200]}}
    if not res.ok:
        err = (body or {}).get("error") or {}
        raise ApiError(res.status_code, err.get("code", f"HTTP_{res.status_code}"), err.get("message", res.reason),
                       err.get("details"), res.headers.get("Retry-After"))
    return body


def print_tasks(job):
    for name, t in job["tasks"].items():
        print(f"  {name:<12} {t['operation']:<16} {t['status']}" + (f"  {t['error']}" if t.get("error") else ""))


def wait_for_job(job_id, max_seconds=600):
    """Wait for a job to settle. One /wait call holds at most 90 s; a longer job answers 408 TIMEOUT
    and we simply ask again."""
    deadline = time.time() + max_seconds
    while True:
        try:
            return api("GET", f"/v1/jobs/{job_id}/wait?timeout=90")
        except ApiError as e:
            if e.status != 408 or time.time() > deadline:
                raise
