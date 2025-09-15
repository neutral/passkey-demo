# Implementation Plan — simplewebauthn + Node Server

## Purpose

- Migrate the Vite+React web app to use `@simplewebauthn/browser` for WebAuthn flows.
- Add a new `node-server/` that reimplements the current Go backend using `@simplewebauthn/server`, matching functionality and policies from `blueprint/done`.
- Preserve data model, security policies (UV required), cookies/CORS, error envelopes, logging, and transaction signing semantics (bundle anchors: CHALv1/TXIDv1).

## Notes

- Keep API shapes JSON-friendly; the browser lib consumes `PublicKeyCredential*OptionsJSON`.
- Parity targets the endpoints and behaviors implemented by done steps: registration, login, session, CORS, limits, logging, error envelopes, transaction signing (options/finish/list), and account key exposure.
- Environment variables retain names from Go server: `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`, `RP_ID_ALLOWLIST`, `ORIGIN_ALLOWLIST`.

## Done

- Step 8 — Node: Login finish (`/authn/passkey/login/finish`) — see `blueprint/done/step-8-node-login-finish.md`
- Step 1 — Web: adopt `@simplewebauthn/browser` for Register/Login — see `blueprint/done/step-1-web-adopt-simplewebauthn-browser.md`
- Step 2 — Node: scaffold `node-server/` (Express + SQLite + Pino) — see `blueprint/done/step-2-node-scaffold-express-sqlite-pino.md`
- Step 3 — Node: DB schema and config parity — see `blueprint/done/step-3-node-db-schema-and-config-parity.md`
- Step 4 — Node: CORS, cookies, and session middleware — see `blueprint/done/step-4-node-cors-cookies-and-session-middleware.md`
- Step 5 — Node: Registration options (`/authn/passkey/registration/options`) — see `blueprint/done/step-5-node-registration-options.md`
- Step 6 — Node: Registration finish (`/authn/passkey/registration/finish`) — see `blueprint/done/step-6-node-registration-finish.md`
- Step 7 — Node: Login options (`/authn/passkey/login/options`) — see `blueprint/done/step-7-node-login-options.md`

---

### Step 9 — Node: GET `/me/account_key` (authenticated)

Scope

- Return the logged-in account’s COSE EC2 public key in JSON and base64url CBOR.

Source to add

- `node-server/src/me.js`: route using session context; CBOR decode to extract x/y for convenience, plus `acct_cbor_b64`.

Verification

- With valid `sid`, returns 200 and expected JSON; without `sid`, 401.

Acceptance criteria

- Shape matches current web expectations for bundle building.

Refs: requirement R-UI-2BTN; requirement R-SCHEMA-LITE; spec bundle-shape-and-client-production-explainer

---

### Step 10 — Node: Bundle anchors and validation helpers

Scope

- Implement `ValidateAndAnchorBundle(acctCBOR, bundleCBORBase64)` in Node to match Go behavior:
  - Decode base64url → CBOR → typed Bundle `{sender_key, nonce, message[, valid_until]}`.
  - Canonical re-encode via `cbor-x` → `B`.
  - Enforce account binding (COSE fields equality), message length ≤ 1024, nonce > last seen and ≤ 2^53-1.
  - Anchors: `challenge = SHA-256('CHALv1'||B)`, `tx_id = SHA-256('TXIDv1'||B)`.
  - Decoder robustness similar to ADR on CBOR/COSE interop.

Source to add

- `node-server/src/tx/bundle.js`: implementation and error codes.

Verification

- Golden vector parity (read `specs/goldens/tx-bundle-v1.json`); compare `B`, `challenge`, `tx_id` to Go output.

Acceptance criteria

- Anchors and policies exactly match Go; error kinds map to envelope codes.

Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails

---

### Step 11 — Node: Tx signing options (`/tx/signing/options`)

Scope

- Auth required (session middleware). Validate bundle, derive anchors, collect account credentials, store a short-lived tx session.
- Use `generateAuthenticationOptions` with `allowCredentials` for this account and `userVerification: 'required'`; set the derived `challenge`.

Source to add

- `node-server/src/tx/options.js`: handler + in-memory TTL store; uses `bundle.js` and DB reads for allowCredentials list.

Request/response

- Request: `{ bundle_cbor_b64: string }`.
- Response: `{ tx_session_id, challenge, options: PublicKeyCredentialRequestOptionsJSON, tx_id_hex, expires_at }`.

Verification

- 401 when unauthenticated; 400 for invalid bundle/base64/CBOR and bundle limits; 409 when nonce not increasing; 200 on success.

Acceptance criteria

- Options JSON feeds directly into `startAuthentication` on the web; logs include `tx_options`.

Refs: requirement R-FLOW-SIGN; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations

---

### Step 12 — Node: Tx signing finish (`/tx/signing/finish`)

Scope

- Verify assertion with `verifyAuthenticationResponse` against the tx session’s `challenge` and policy; require UV; check rpIdHash/origin.
- Persist transaction: compute `tx_id` from `B`; decode `B` to extract `nonce` and `message`; insert row with AD, CDJ, and signature.

Source to add

- `node-server/src/tx/finish.js`: handler using `bundle.js` for anchors and DB writes; error mapping to envelope.

Request/response

- Request: `AuthenticationResponseJSON` plus `tx_session_id`.
- Response: `201 Created` `{ tx_id_hex }`.

Verification

- Negative paths: 400 malformed, 401 expired/unknown session or signature mismatch, 403 policy, 409 nonce not increasing.

Acceptance criteria

- Transaction recorded; nonce advances; list reflects new entry.

Refs: requirement R-FLOW-SIGN; decision webauthn-accept-high-s-signing-too; decision http-error-envelope

---

### Step 13 — Node: Tx list (`/tx/list`)

Scope

- Auth required; return the account’s transactions ordered by `created_at DESC`.

Source to add

- `node-server/src/tx/list.js`: handler; SQL mirrors Go; may add prepared statements/repo layer later.

Verification

- With `sid`, returns list; without, 401.

Acceptance criteria

- Shape matches web expectations: `{ items: [{ tx_id_hex, nonce, message, created_at }, ...] }`.

Refs: requirement R-UI-2BTN; requirement R-PLAT-3

---

### Step 14 — Node: Error envelopes and limits

Scope

- Standard JSON error envelope `{code,error,correlation_id?}`; consistent mapping:
  - 400: `ERR_BAD_REQUEST` (malformed inputs/base64/JSON/DER)
  - 401: `ERR_UNAUTHORIZED` (no/invalid session; unknown credential; signature mismatch)
  - 403: `ERR_FORBIDDEN` (origin/rp policy; UV missing)
  - 409: `ERR_CONFLICT` (signCount or nonce policy)
  - 413: `ERR_PAYLOAD_TOO_LARGE`
  - 429: `ERR_RATE_LIMIT`
  - 500: `ERR_INTERNAL`
- Apply body size caps and token-bucket rate limiting to `/authn/*` and `/tx/*` routers.

Source to add/modify

- `node-server/src/error.js`, `node-server/src/limits.js` (body limit + rate limit middleware); wrap routers in `server.js`.

Verification

- Curl negative cases return envelopes and statuses; rate limiting blocks over-burst.

Acceptance criteria

- Parity with Go backend mapping and behavior.

Refs: requirement R-ERR; decision http-error-envelope; decision router-builder-wiring

---

### Step 15 — Node: Structured logging and correlation

Scope

- Pino JSON logs with event names: `server_start`, `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`, `webauthn_assert_verify`.
- Include `correlation_id` (from request-id middleware), `rp_id`, `origin`, hashes (`credential_id_hash`, `account_thumb_hex`).

Source to add/modify

- `node-server/src/logger.js` event helpers; integrate in handlers.

Verification

- Logs appear with expected fields; error flows include stable `error_kind` where applicable.

Acceptance criteria

- Logging parity with Go for key flows.

Refs: decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines

---

### Step 16 — Web: adapt to Node server option/finish shapes

Scope

- Switch API base to `ORIGIN` for CORS; ensure fetches include credentials where needed.
- If Node options include only lib JSON without extra `*_session_id`, add an adapter to include/extract session identifiers as needed (prefer explicit ids for parity).
- Update Playwright config to boot Node server during E2E runs.

Source to modify

- `web/src/config.ts` (if present) or callers using `apiUrl()` to point to Node server.
- `web/tests/*.spec.ts` to expect JSON lib shapes.

Verification

- E2E: register → login → dashboard sign → list update works against Node server.

Acceptance criteria

- Web app interoperates with Node API using `@simplewebauthn/browser` JSON.

Refs: requirement R-PLAT-1; requirement R-OPS-DEV

---

### Step 17 — E2E and golden parity

Scope

- Add Node-side test script to validate bundle golden vectors (`specs/goldens/tx-bundle-v1.json`).
- Smoke E2E via Playwright: run Node server and Vite; Virtual Authenticator; assert all flows.

Source to add

- `node-server/scripts/golden-check.js` reading golden JSON and comparing outputs.
- `web/playwright.config.ts` updates to run Node server for E2E project.

Verification

- `node node-server/scripts/golden-check.js` passes; Playwright E2E green.

Acceptance criteria

- Bundle anchors and flows verified end-to-end against goldens.

Refs: requirement R-FLOW-SIGN; spec golden-vectors; spec manual-e2e

---

### Step 18 — Docs, env, and examples

Scope

- Update `.env.example` with Node server variables; update README with Node start commands.
- Provide REST client examples for Node endpoints; ensure Postman or REST Client entries reflect JSON shapes.

Source to add/modify

- `node-server/README.md`, `.env.example` additions; `docs/` updates; API examples under `docs/`.

Verification

- Quick curl scripts work against Node server; README steps reproduce register → login → sign → list.

Acceptance criteria

- Documentation parity and clear local dev instructions.

Refs: requirement R-OPS-DEV; step-41/42/43 analogs; decision http-error-envelope

---

## User Verification Commands (after all steps)

```bash
# 1) Install deps
npm -C web ci || npm -C web i
npm -C node-server ci || npm -C node-server i

# 2) Start Node server (dev)
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo-node.db \
  node node-server/src/server.js &
API_PID=$!

# 3) Start Vite
npm -C web run dev &
WEB_PID=$!

# 4) Manual E2E
# Visit http://localhost:5173 → Register → Login → Dashboard sign → See list item

# 5) Golden check
node node-server/scripts/golden-check.js

# 6) Cleanup
kill $WEB_PID || true
kill $API_PID || true
```

## Refs

Refs: goal passkey-registration-login-uv; goal server-derived-challenge-and-txid; goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-PLAT-3; requirement R-SEC-UV; requirement R-ERR; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-OPS-DEV; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision http-error-envelope; decision request-id-and-slog-json; decision cbor-cose-interop-and-decoding-fallbacks; decision webauthn-corrections-and-standardizations; decision webauthn-accept-high-s-login-only; decision webauthn-accept-high-s-signing-too
