"""3. Realtime updates over a WebSocket instead of polling.
wss://api.convertintomp4.com/ws/user/jobs streams every job of the key's account:
{"type": "started" | "progress" | "completed" | "failed", "jobId", "progress"?, "outputUrl"?, "error"?}.
For multi-task jobs, also pass `webhook_url` when you create the job: you get one signed
job.finished / job.failed POST when the whole job settles (see verify_webhook below)."""
import asyncio
import hashlib
import hmac
import json

import websockets

from common import API_URL, SAMPLE_FILE, api, auth_headers


async def main():
    url = API_URL.replace("http", "ws", 1) + "/ws/user/jobs"
    async with websockets.connect(url, additional_headers=auth_headers()) as ws:
        await ws.recv()  # {"type": "connected"}
        # Start the conversion only once the socket is open, so no event is missed.
        created = await asyncio.to_thread(api, "POST", "/v1/convert", files={"file": ("sample.png", SAMPLE_FILE.read_bytes())}, data={"targetFormat": "jpg"})
        job_id = created["jobId"]
        print("Created conversion", job_id)
        async for raw in ws:
            msg = json.loads(raw)
            if msg.get("jobId") != job_id:
                continue  # other jobs of the same account
            print(" ", msg["type"], f"{msg['progress']}%" if "progress" in msg else "")
            if msg["type"] == "completed":
                print("Download:", API_URL + msg["outputUrl"])
                return
            if msg["type"] == "failed":
                print("Failed:", msg.get("error"))
                return


def verify_webhook(raw_body: bytes, signature_header: str, secret: str) -> bool:
    """Webhook receivers: check X-ConvertIntoMP4-Signature on the raw body before trusting it.
    The secret comes from GET /v1/jobs/webhook-secret."""
    expected = "sha256=" + hmac.new(secret.encode(), raw_body, hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, signature_header or "")


if __name__ == "__main__":
    asyncio.run(main())
