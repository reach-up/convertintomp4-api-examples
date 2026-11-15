# Node.js examples

Node.js 18 or newer (built-in `fetch` and `FormData`). The only dependency is `ws` for example 03.

```bash
npm install
export CIM4_API_KEY=your_key
node 01-tasks-and-jobs.mjs
```

Shared helpers (auth headers, error class, retry on 429, waiting for a job) are in `common.mjs`.
