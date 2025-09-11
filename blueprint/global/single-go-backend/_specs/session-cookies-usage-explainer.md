# R-PLAT-2 — Session Cookies Usage Explainer (Non‑Developer)

## Purpose & Audience
- For collaborators who need to understand how login state is maintained between browser and backend without reading code.
- Explains what the session cookie is, how it’s issued and used, and what protections apply in local dev vs. HTTPS.

## What Session Cookies Are
- A small identifier (`sid`) the server sets in the browser after a successful login so subsequent API calls know which account is active.
- Stored by the browser as a cookie and sent automatically on requests that meet cookie policy.

## Why We Use It Here
- Keeps the demo simple: no tokens or client storage; the browser carries a single cookie.
- Works across pages and refreshes; the server stays authoritative for session validation and expiry.

## What We Store / Handle
- In the database `sessions` table: `session_id` (the cookie value), `acct_cbor` (account identity), `expires_at` (Unix seconds).
- In the browser: only the `sid` cookie; it is a session cookie (no explicit `Max-Age`/`Expires`) and is cleared when the browser session ends.

## How It Works in This App
- On login success, the server creates a session row (TTL ~1 hour) and sets cookie `sid=<session_id>` with attributes:
  - `HttpOnly` (not readable by JavaScript)
  - `SameSite=Lax` (limits cross‑site sends to safer cases)
  - `Secure` set to `true` only when the configured `ORIGIN` is HTTPS; for localhost http dev, `Secure` is omitted.
- Subsequent API calls from the SPA include the cookie automatically when using `credentials: 'include'` and matching SameSite rules.
- Protected endpoints (e.g., transaction APIs) require a valid, unexpired session; otherwise they return 401 Unauthorized.
- Session middleware looks up the session by `sid`, checks expiry, and attaches the account identity to the request context. Handlers do not read cookies directly.

## Performance & Safety Settings
- Session TTL: ~1 hour; may refresh on activity (middleware can extend expiry when half TTL remains).
- Cookie is session‑scoped (ends with browser session) even if the server’s DB expiry is longer; both must be valid for access.
- Same‑site semantics: `http://localhost:5173` (SPA) → `http://localhost:8080` (API) is considered same‑site, so `SameSite=Lax` allows cookie on typical SPA requests.

## Security & Privacy
- `HttpOnly` prevents script access; reduces risk from XSS reading the cookie.
- `Secure` required on HTTPS so cookies are only sent over encrypted transport; permitted to be off for `http://localhost` during dev.
- `SameSite=Lax` reduces cross‑site request risks; combined with CORS, the API only accepts credentialed requests from allowed origins.
- The cookie contains no PII or secrets; it’s an opaque ID; server validates it against the DB.

## Operating It Day‑to‑Day
- Start the app, log in once to establish the session; verify a `sid` cookie appears in the browser for the API domain.
- To reset: log out (if provided) or clear browser cookies for `localhost`; the server also expires/cleans sessions over time.
- Local dev env vars: `ORIGIN=http://localhost:5173` ensures cookie `Secure` is omitted so the browser can send it over HTTP.
- Frontend requests must set `credentials: 'include'` to carry the cookie.

## Limitations & When to Upgrade
- Not a production SSO/session management system; demo‑grade session TTL and policies.
- For production: enforce HTTPS everywhere; consider `SameSite=Strict` or CSRF tokens for form posts; add logout endpoints and server‑side revocation.

## Errors & Observability
- Missing/invalid/expired session → HTTP 401 with a clear error code per error policy.
- Check server logs for login events and session creation; inspect the `sessions` table for active rows.

## Glossary
- `sid`: session identifier stored as a cookie and used to look up the server session.
- `SameSite`: cookie attribute controlling when browsers send cookies on cross‑site requests.
- `Secure`: cookie attribute requiring HTTPS for transmission.

## Refs
- Refs: requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-ERR; goal simple-ui-and-storage; decision webauthn-corrections-and-standardizations
