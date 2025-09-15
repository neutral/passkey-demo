### Step 4 — Node: CORS, cookies, and session middleware

Scope

- Implement credentialed CORS with an exact origin allowlist: `ORIGIN ∪ ORIGIN_ALLOWLIST`. Handle preflight correctly and set `Vary: Origin`.
- Add session middleware that reads the `sid` cookie, loads the session from SQLite, validates expiry, optionally performs a rolling refresh, and attaches session context `{ acct_cbor, credential_ids }` to `req`.
- Ensure cookie attributes align with policy: `HttpOnly; Path=/; SameSite=Lax; Secure` when origin is https; omit `Secure` for `http://localhost` dev.

Source to add/modify

- Add `node-server/src/cors.js` — Express middleware implementing:
  - Preflight (OPTIONS): if `Origin` is allowed and method/headers are acceptable, return 204 with headers: `Access-Control-Allow-Origin: <origin>`, `Access-Control-Allow-Methods: GET, POST, OPTIONS`, `Access-Control-Allow-Headers: Content-Type`, `Access-Control-Allow-Credentials: true`, `Access-Control-Max-Age: 600`, `Vary: Origin`.
  - Non-preflight: if `Origin` allowed, set `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials: true`, and `Vary: Origin`.
  - Disallowed origin: preflight → 403 with no `Access-Control-*`; non-preflight → do not set CORS headers.
  - No `Origin` header (same-origin or tools): pass through with no CORS headers changed.
- Add `node-server/src/session.js` — Express middleware implementing:
  - Parse `sid` from `Cookie` header; if absent, continue (unauthenticated).
  - Query `sessions` for row by `session_id`; ensure `expires_at > now`. If expired, treat as unauthenticated.
  - Attach `req.session = { sid, acct_cbor, expires_at, credential_ids?: Buffer[] }` and optionally refresh expiry (rolling TTL) when within a configurable window.
  - Optionally pre-load credential ids (`SELECT credential_id FROM credentials WHERE acct_cbor_fk=?`) to support `allowCredentials` construction in later steps.
- Modify `node-server/src/server.js` — wire `cors` middleware globally and `session` middleware before protected routes (authn/tx groups in later steps). Keep `/health` public.

Description files

- Add `node-server/src/cors.js.desc.md` — detail headers, allowed origins, preflight behavior, and `Vary` policy.
- Add `node-server/src/session.js.desc.md` — describe cookie parsing, DB lookup, expiry/rolling refresh, and attached request context.
- Update `node-server/src/server.js.desc.md` — note middleware order: request-id → http logger → json body → CORS → session → routes.

Blueprint updates (requirements/specs/ADRs/user-flows to add/update)

- Specs (OPS/Dev): update `blueprint/global/easy-local-dev/_specs/cors-usage-explainer.md` to include Node implementation notes and verify headers for preflight vs actual requests.
- Specs (OPS/Dev): update `blueprint/global/easy-local-dev/_specs/dev-servers-and-origins.md` with cookie attribute behavior in Node (same policy as Go).
- Specs (NFR): update `blueprint/global/single-go-backend/_specs/session-cookies-usage-explainer.md` to apply language-agnostically, or add a Node-focused peer under `blueprint/global/single-node-backend/_specs/` with identical semantics.
- Requirements: confirm `blueprint/global/easy-local-dev/requirement.md` Acceptance Criteria include credentialed CORS (exact allowlist) and session cookie handling for Node.
- ADRs: no new ADR required; ensure `http-error-envelope` remains referenced for negative CORS outcomes (403 preflight) and session failures (401 in protected routes later).

Request/response shape

- Preflight request: `OPTIONS <path>` with `Origin`, `Access-Control-Request-Method`, `Access-Control-Request-Headers`.
- Preflight response (allowed): 204 No Content with headers listed above.
- Cookie: `sid=<opaque>`; not set in this step (issued in login finish later) but read by session middleware.

Algorithm

- CORS:
  - Determine `allowed = origin === ORIGIN || ORIGIN_ALLOWLIST.has(origin)`.
  - If `OPTIONS`: if not allowed → 403; if allowed and requested method in {GET,POST} and headers ⊆ {Content-Type} → 204 with allow headers; else 403.
  - If non-OPTIONS and allowed: set `Access-Control-Allow-Origin` echo and `Access-Control-Allow-Credentials: true`; always set `Vary: Origin` when `Origin` present.
- Session:
  - Parse cookies; read `sid`; if missing → next().
  - `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?`; if no row or expired → next().
  - Optionally `SELECT credential_id FROM credentials WHERE acct_cbor_fk=?` to populate `credential_ids`.
  - Attach context; if rolling refresh configured and near expiry, `UPDATE sessions SET expires_at=?`.

Database interactions

- Reads from `sessions` and optionally `credentials`; optional `UPDATE sessions` for rolling TTL.
- No schema changes.

Policies & limits

- CORS: exact-match origin allowlist; allowed methods {GET, POST, OPTIONS}; allowed header `Content-Type`; `Max-Age=600`; `Vary: Origin` on responses that include `Origin`.
- Cookies: `sid` is `HttpOnly; SameSite=Lax; Path=/; Secure` on https; in dev (`http://localhost`) omit `Secure`.
- Session: treat missing/expired/unknown session as unauthenticated; do not 401 at middleware; protected routes will enforce.

Sequencing

- Place CORS before session middleware; keep `/health` publicly accessible.
- Protected route enforcement (401) will be implemented in authn/tx steps; session middleware only annotates the request.

Tests

- Use `node:test` against ephemeral port where needed.
- Files to add:
  - `node-server/test/cors.test.js`:
    - Preflight allowed: `OPTIONS /authn/passkey/login/options` with allowed `Origin` → 204 with correct headers.
    - Preflight disallowed: same with disallowed `Origin` → 403 and no `Access-Control-*` headers.
    - Actual allowed: `GET /health` with allowed origin → includes `Access-Control-Allow-Origin` and `Allow-Credentials`.
  - `node-server/test/session.test.js`:
    - Setup temp DB; run migrations; insert a valid `sessions` row; request with `Cookie: sid=...` → test route echoes `hasSession: true`.
    - Expired/unknown sid → `hasSession: false`.
- Commands: `node --test node-server/test/cors.test.js node-server/test/session.test.js`.

Verification

- Manual curl checks:
  - Preflight allowed:
    - `curl -i -X OPTIONS :8080/authn/passkey/login/options -H 'Origin: http://localhost:5173' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type'`
  - Preflight disallowed:
    - `curl -i -X OPTIONS :8080/authn/passkey/login/options -H 'Origin: https://evil.example' -H 'Access-Control-Request-Method: POST' -H 'Access-Control-Request-Headers: Content-Type'`
  - Actual:
    - `curl -i :8080/health -H 'Origin: http://localhost:5173'`

Acceptance criteria

- CORS middleware returns 204 with correct headers for allowed preflight and 403 with no CORS headers for disallowed origins; actual responses include `Access-Control-Allow-Origin` and `Allow-Credentials` for allowed origins.
- Session middleware parses `sid`, loads session, validates expiry, and attaches session context; no 401s are thrown in middleware.
- Description files added and updated; tests planned for both CORS and session.

Refs: requirement R-OPS-DEV; spec cors-usage-explainer; spec session-cookies-usage-explainer; decision http-error-envelope

---

