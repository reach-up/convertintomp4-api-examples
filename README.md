# ConvertIntoMP4 API examples

Runnable examples for the [ConvertIntoMP4 file conversion API](https://convertintomp4.com/api-docs) in **Node.js, Python, Go and .NET**. Each language has the same six examples:

| # | Example | What it shows |
|---|---|---|
| 01 | Tasks and jobs | One request runs a pipeline: import a file from a URL, convert it, export a download link |
| 02 | Check status with polling | `GET /v1/jobs/{id}` with back-off until the job settles |
| 03 | Realtime updates | WebSocket stream of job events (`/ws/user/jobs`), plus verifying signed job webhooks |
| 04 | Uploading files | One-call upload to `POST /v1/convert`, and an `import/upload` task inside a job |
| 05 | Results and errors | Download the result, read the error shape, retry only what can succeed later |
| 06 | Complex job | Web page capture to PDF, image to PDF, merge, thumbnail, zip, export: 7 tasks in one request |

## Setup

1. Create an API key in your dashboard: https://convertintomp4.com/account/api/keys
2. `export CIM4_API_KEY=your_key`
3. Pick a language folder and follow its README.

The examples run in **sandbox mode** by default (`X-Sandbox: true`): files up to 1 MB and 100 requests a month per key, so you can try everything on the bundled 10 KB sample image. Set `CIM4_SANDBOX=0` to run them live against your plan.

| Language | Folder | Run |
|---|---|---|
| Node.js 18+ | [nodejs](nodejs) | `npm install && node 01-tasks-and-jobs.mjs` |
| Python 3.9+ | [python](python) | `pip install -r requirements.txt && python 01-tasks-and-jobs.py` |
| Go 1.22+ | [go](go) | `go run . 01` |
| .NET 8 | [dotnet](dotnet) | `dotnet run -- 01` |

## The API in one minute

- **Base URL:** `https://api.convertintomp4.com`, auth header `X-Api-Key: <key>`.
- **Jobs:** `POST /v1/jobs` with named `tasks`. A task reads another task's output through `input` (a name, or a list of names for `merge` and `archive/create`).
- **Operations:** `import/url`, `import/upload`, `import/base64`, `import/raw`, `convert`, `optimize`, `compress`, `merge`, `thumbnail`, `metadata/read`, `capture-website`, `archive/create`, `archive/extract`, `trim`, `crop`, `pdf/*` (ocr, split, merge, encrypt, decrypt, rotate, compress, extract-pages, watermark), `export/url`, `export/upload`, `export/s3`, `export/azure`, `export/gcs`, `export/sftp`. Full list and options: https://convertintomp4.com/api-docs
- **Status:** `GET /v1/jobs/{id}`, or `GET /v1/jobs/{id}/wait?timeout=90` (long poll, call again after a 408), or the WebSocket, or a `webhook_url` on the job (signed with `X-ConvertIntoMP4-Signature: sha256=<hmac>`; secret from `GET /v1/jobs/webhook-secret`).
- **Errors:** `{ "success": false, "error": { "code": "INVALID_INPUT", "message": "...", "details": {} } }` with the HTTP status. Retry 429 and 5xx; fix the request on other 4xx.
- **Results:** an `export/url` task returns `result.url` (a signed link, valid 24 hours) and `result.file` (`name`, `size`).

## SDKs

If you would rather not call HTTP yourself: [Node.js SDK](https://github.com/reach-up/convertintomp4-node) (`npm install convertintomp4`), [Python SDK](https://github.com/reach-up/convertintomp4-python), [n8n node](https://github.com/reach-up/n8n-nodes-convertintomp4).

## License

MIT, see [LICENSE](LICENSE).
