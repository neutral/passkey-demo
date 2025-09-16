# Step 14 — Node: Error envelopes and limits (Done)

Completed: 2025-09-16
Verification notes:
- `cd node-server && node --test test/error-envelope.test.js`
- `cd node-server && node --test test/limits.test.js`
- `cd node-server && npm test`
- `bash tools/desc-check.sh HEAD~1 HEAD`

### Step 14 — Node: Error envelopes and limits

Scope
− Standardize JSON error envelope to `{ code: 'ERR_*', error: string, correlation_id?: string }` for all API responses.
− Align HTTP status ↔ code mapping with Go backend parity (400/401/403/409/413/429/500) and ensure every handler uses centralized helpers.
− Enforce request body size limits (exact bytes per flow) and shared token-bucket rate limiting on `/authn/*` and `/tx/*` routers.

Source to add/modify
− `node-server/src/error.js` — replace simple helper with typed error factory exposing constants (e.g., `ERR_BAD_REQUEST`) and convenience functions (`respondBadRequest(res, correlationId, message?, opts?)`). Ensure backwards-compatible API for tests via re-export.
− `node-server/src/limits.js` — new module encapsulating `express.json` size middleware (configurable per route) and a configurable token-bucket limiter (probably `rate-limiter-flexible` or custom minimal implementation) with helpers `buildBodyLimit({ bytes })` and `buildRateLimiter({ points, duration })`.
− `node-server/src/server.js` — wire body limit + rate limit middleware around `/authn/passkey/*` and `/tx/*` routers; ensure order respects session middleware and logging. Adjust JSON parser limit (currently 1mb) to delegate to new body limit wrappers.
− Existing route handlers (`webauthn/reg.js`, `webauthn/login.js`, `tx/options.js`, `tx/finish.js`, `tx/list.js`, `me.js`) — update to use new error helper signatures and codes; ensure catch blocks log `error_kind` or code where appropriate.
− Test suites touching error codes (`node-server/test/*`) — update expectations to match new envelope codes/status; add new cases for 413 and 429.
− Optional shared test utils (if helpful) to DRY checking error bodies (`test/helpers/response.js`).

Description files
− `node-server/src/error.js.desc.md` — describe standardized envelope, helper APIs, and mapping table.
− `node-server/src/limits.js.desc.md` (new) — document body limit + rate limiter behavior, defaults, and interactions with routers.
− Update route-specific `.desc.md` files impacted by policy changes (`node-server/src/webauthn/reg.js.desc.md`, `login.js.desc.md`, `tx/options.js.desc.md`, `tx/finish.js.desc.md`, `tx/list.js.desc.md`, `me.js.desc.md`) to reflect new error codes and limits.
− `node-server/src/server.js.desc.md` — note rate/body limit middleware layering.

Blueprint updates
− Review `blueprint/global/error-responses-and-limits/requirement.md` and accompanying spec to confirm Node implementation satisfies documented limits; update if numeric thresholds differ (likely note actual byte limits chosen). If aligning values already covered, append verification note only.
− No new ADRs expected; ensure `Refs` in updated docs continue to cite `decision http-error-envelope` and `decision router-builder-wiring`.

Request/response shape
− Error envelope (all endpoints):
  ```json
  { "code": "ERR_BAD_REQUEST", "error": "Invalid bundle", "correlation_id": "uuid-123" }
  ```
− Rate limit response: status 429 with `ERR_RATE_LIMIT` and `Retry-After` header containing seconds until reset.
− Payload too large: status 413 with `ERR_PAYLOAD_TOO_LARGE` and optional `max_bytes` metadata in body if helpful (string or number field in `details`).

Algorithm
− Error helper should accept `statusCode`, canonical `code`, optional `details`, and default user-facing `error` message; centralize mapping to avoid drift.
− For body size limits, wrap JSON parser in route-specific middleware (e.g., registration/login finish allowed 16KB, options maybe 32KB). Use `express.json({ limit })` per router rather than global 1MB.
− Implement rate limiter using per-IP buckets (keyed by `req.ip` or forwarded header per config), with budgets defined per requirement (e.g., 5 req/sec burst 10 for ceremony routes). On exceed, call error helper with 429.
− Ensure middleware attaches limiter before handlers but after request-id/logging for correlation; always call `next()` with error handled via express (or respond directly) to maintain consistent envelope.
− Update handlers to remove hardcoded lowercase codes; leverage helper for clarity (e.g., `return respondUnauthorized(res, correlationId, 'Unauthorized')`).
− Guarantee existing catch blocks propagate `correlation_id` and, where available, include `error_kind` metadata by extending logger payload.

Database interactions
− None beyond existing handler queries; rate limiter should use in-memory token bucket (no DB) for simplicity per demo scope.

Policies & limits
− Map statuses to codes exactly:
  - 400 → `ERR_BAD_REQUEST`
  - 401 → `ERR_UNAUTHORIZED`
  - 403 → `ERR_FORBIDDEN`
  - 409 → `ERR_CONFLICT`
  - 413 → `ERR_PAYLOAD_TOO_LARGE`
  - 429 → `ERR_RATE_LIMIT`
  - 500 → `ERR_INTERNAL`
− Recommended body size caps (confirm spec): registration/login options ≤ 32KB, finish ≤ 16KB, tx bundle/options ≤ 64KB; list/me remain GET (no body).
− Rate limit: e.g., 10 req/min burst 20 per IP for `/authn/*`, 30 req/min for `/tx/*`; align with requirement or Go defaults.
− Document how limits surface to clients (headers such as `Retry-After`).

Sequencing
− Introduce new error helper + migrate routes first to keep tests compiling.
− Layer new middleware in `server.js` once helpers ready; ensure global JSON parser limit doesn’t conflict (remove/adjust existing `express.json({ limit: '1mb' })` or scope to `/health`).
− Update tests after code changes to avoid transient failures; ensure Step 15 logging adjustments will hook into new error helpers (expose `error_kind` fields for future use).

Tests
− Update existing unit tests that assert specific error codes/messages:
  - `test/login-options.test.js`, `test/login-finish.test.js`, `test/reg-options.test.js`, `test/reg-finish.test.js`, `test/tx-options.test.js`, `test/tx-finish.test.js`, `test/tx-list.test.js`, `test/me.test.js` (if present).
− Add new tests:
  - `test/error-envelope.test.js` covering helper behavior (e.g., `respondBadRequest` sets status + body and includes correlation id).
  - `test/limits.test.js` verifying body limit middleware returns 413 with `ERR_PAYLOAD_TOO_LARGE` and includes optional metadata.
  - Rate limiting test (using fake clock or sequential calls) ensuring 429 after threshold and `Retry-After` header.
− Ensure tests cover correlation id propagation, absence of `details` when not provided, and HEAD/OPTIONS unaffected.
− Commands: `cd node-server && node --test test/error-envelope.test.js`, `node --test test/limits.test.js`, full `npm test`.

Verification
− Run targeted new tests (`node --test test/error-envelope.test.js`, `node --test test/limits.test.js`).
− Re-run suites impacted by changes (`npm test`).
− Manual: start server, use curl to trigger 400 (malformed JSON), 401 (without session), 413 (body over limit), 429 (rapid repeated calls) and confirm envelopes/header values.
− `bash tools/desc-check.sh HEAD~1 HEAD` after edits to ensure description coverage.

User verification commands
```bash
cd node-server
npm test
curl -sS -X POST http://localhost:8080/authn/passkey/login/options -H 'Content-Type: application/json' -d '{"broken":true}' || true
python - <<'PY'
import requests
url = 'http://localhost:8080/tx/signing/options'
payload = {'bundle_cbor_b64': 'A' * 200000}
for _ in range(6):
    r = requests.post(url, json=payload)
    print(r.status_code, r.json())
PY
```

Acceptance criteria
− All handlers return standardized envelopes with upper-case codes mapped to documented HTTP statuses.
− Body size overages yield 413 with clear messaging; rate limiter enforces thresholds and responds with 429 + `Retry-After`.
− Updated tests reflect new envelopes/limits and all suites pass.

Notes
− Confirm global error handling (Express default) cannot leak stack traces; ensure helper covers unexpected errors.
− Coordinate numeric limits with frontend expectations (web app may display message strings; keep them user-friendly).
− Rate limiter should be resilient in dev (allow env override to disable for local testing if needed).

Refs: requirement R-ERR; decision http-error-envelope; decision router-builder-wiring
