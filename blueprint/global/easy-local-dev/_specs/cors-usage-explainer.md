# R-OPS-DEV — CORS Usage Explainer (Non‑Developer)

## Purpose & Audience
- For product, QA, and operations collaborators who need to understand how the web app talks to the backend during local development.
- Covers why CORS is needed, how it’s configured in this project, what headers you should expect, and how to verify behavior.

## What CORS Is
- A browser security feature that restricts web pages from making requests to a different origin unless the server explicitly allows it.
- It uses HTTP response headers and a pre‑flight check to grant or deny access for a given origin, method, and headers.

## Why We Use It Here
- Our UI runs on `http://localhost:5173` (Vite dev server) while the API runs on `http://localhost:8080` — different origins.
- CORS allows the browser to call the API from the UI and include the session cookie when permitted.

## What We Store / Handle
- No data is stored by CORS itself; it is purely request/response headers.
- We allow credentials (cookies) on approved origins to carry the user session (`sid`).

## How It Works in This App
- The backend includes a CORS middleware that checks the request `Origin` header against an allowlist: `{ORIGIN} ∪ ORIGIN_ALLOWLIST`.
- If the origin matches exactly:
  - For preflight (`OPTIONS`) requests: server returns 204 with headers:
    - `Access-Control-Allow-Origin: <exact origin>`
    - `Access-Control-Allow-Methods: GET, POST, OPTIONS`
    - `Access-Control-Allow-Headers: Content-Type`
    - `Access-Control-Allow-Credentials: true`
    - `Access-Control-Max-Age: 600`
    - `Vary: Origin`
  - For normal requests: server adds:
    - `Access-Control-Allow-Origin: <exact origin>`
    - `Access-Control-Allow-Credentials: true`
    - `Vary: Origin`
- If the origin does not match:
  - Preflight returns 403 with no `Access-Control-*` headers.
  - Normal responses omit CORS headers; browsers will block access from the page.
- The middleware is applied outermost, so preflights are answered early and are not rate‑limited.

## Performance & Safety Settings
- Exact origin match only; wildcards (`*`) are never used when credentials are involved.
- Allowed methods: `GET, POST, OPTIONS`. Allowed header: `Content-Type` (JSON).
- Preflight caching: `Access-Control-Max-Age: 600` seconds reduces preflight chatter.
- `Vary: Origin` prevents cache poisoning across different origins.

## Security & Privacy
- Credentials are only allowed for approved origins; responses never use `*` with credentials.
- No sensitive data is added to headers; cookies remain `HttpOnly` and are not readable by script.
- Origin allowlists are explicit; no pattern matching or wildcards.

## Operating It Day‑to‑Day
- Configure allowed UI origin via environment:
  - `ORIGIN=http://localhost:5173`
  - Optionally add more via `ORIGIN_ALLOWLIST=http://127.0.0.1:5173`
- Verify with curl or browser dev tools:
  - Preflight: send `OPTIONS` with `Origin` and `Access-Control-Request-Method: POST`; expect 204 with headers above.
  - Actual: `GET /health` with `Origin: http://localhost:5173`; expect `Access-Control-Allow-Origin` echo and `Allow-Credentials: true`.
- Frontend must use `fetch(..., { mode: 'cors', credentials: 'include' })` so the browser sends/receives the session cookie.

## Limitations & When to Upgrade
- Designed for a single SPA origin in local dev; not a production multi‑tenant CORS gateway.
- For multiple environments/domains, move to a centrally managed allowlist and audit logging for CORS decisions.
- Always use HTTPS with `Secure` cookies outside of localhost.

## Errors & Observability
- Disallowed origin preflight: HTTP 403 with no `Access-Control-*` headers.
- Normal requests from disallowed origins are served but blocked by the browser; check the browser console for CORS errors.
- Server logs startup with configured `origin` and may log CORS decisions in debug.

## Glossary
- Origin: scheme + host + port (e.g., `http://localhost:5173`).
- Preflight: an `OPTIONS` request the browser sends to ask permission before the actual request.
- Credentials: cookies or HTTP auth that the browser can include when permitted.

## Refs
- Refs: requirement R-OPS-DEV; requirement R-PLAT-2; requirement R-PLAT-1; goal simple-ui-and-storage; decision webauthn-corrections-and-standardizations

