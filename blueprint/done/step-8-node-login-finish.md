# Step 8 — Node: Login finish (Done)

### Step 8 — Node: Login finish (`/authn/passkey/login/finish`)

Scope

- Implement the login finish handler using `verifyAuthenticationResponse`, consuming the login session issued in Step 7 and establishing persistent sessions (DB row + sid cookie) with policy parity to the Go backend.
- Enforce challenge binding, RP/origin allowlists, user verification, and monotonic `signCount`; handle duplicate credential/session conflicts with proper JSON envelopes.
- Emit structured `login_finish` logs with hashed identifiers and verification outcomes.
- Update the web app to post the SimpleWebAuthn response JSON (with `login_session_id`) and consume the Node server’s 200 payload.

Source to add/modify

- Extend `node-server/src/webauthn/login.js` with the finish handler, DB fetch/update logic, helper functions for credential hashing, and error mapping.
- Update `node-server/src/server.js` to pass required dependencies (DB, session store) into login routes if not already (reuse shared store via `app.locals`).
- Add `node-server/test/login-finish.test.js` covering success, session expiry, credential mismatch, UV failures, signCount regression, duplicate session usage, and malformed input.
- Update frontend (`web/src/pages/Login.tsx`, related tests) as needed to align with Node finish behavior (payload shape already mostly in place; verify any adjustments).

Description files

- Update `node-server/src/webauthn/login.js.desc.md` to document finish handler, DB interactions, logging fields, and envelope mapping.
- Update `node-server/src/webauthn/webauthn.desc.md` to note shared helper reuse (credential hashing/logging) across login finish.
- Update `node-server/src/server.js.desc.md` if dependency injection changes (e.g., DB passed into login routes).
- Update `node-server/test/login-options.test.js.desc.md` (if cross-reference needed) and add `node-server/test/login-finish.test.js.desc.md` summarizing coverage.
- Update `web/src/pages/Login.tsx.desc.md` with finish handler details (JSON payload, session cookie expectations).

Blueprint updates

- Update `blueprint/features/login-flow/_specs/spec.md` with finish algorithm details: session lookup, verification steps, DB updates, cookie attributes, logging, and error mapping (401/403/409).
- Update `blueprint/features/login-flow/requirement.md` acceptance criteria to mention session creation (`sid` cookie), signCount enforcement, duplicate handling, and structured logs.
- If necessary, update `blueprint/features/login-flow/_specs/frontend-mapping` to note Node finish request/response changes.
- Add references to any ADRs governing session cookies, logging, or verification policies if not already noted.

Request/response shape

- Request: JSON with `login_session_id`, plus SimpleWebAuthn `AuthenticationResponseJSON` fields (`id`, `rawId`, `type`, `response`: `{ authenticatorData, clientDataJSON, signature, userHandle? }`).
- Response: 200 JSON `{ account_thumb_hex, credential_id_b64 }`, `Set-Cookie: sid=<id>; Path=/; HttpOnly; Secure?; SameSite=None` (per Step 4 policies).
- Error envelopes: 400 (malformed/missing fields), 401 (invalid session, signature failures), 403 (policy violations: origin, UV), 409 (signCount regression, session reuse), 500 (unexpected errors).

Algorithm

- Validate request body types; ensure `login_session_id` is a non-empty string and `response` contains SimpleWebAuthn data.
- Look up session from `LoginSessionStore`; if missing/expired, delete stale entry and return 401.
- Retrieve credential from DB by `credential_id`; if not found, return 401.
- Build verification options for `verifyAuthenticationResponse`:
  - `expectedChallenge = session.challenge`
  - `expectedOrigin = [config.ORIGIN, ...allowlist]`
  - `expectedRPID = [config.RP_ID, ...allowlist]`
  - `requireUserVerification = true`
  - Provide `authenticator` object with stored `credentialPublicKey`, `credentialID`, `counter`.
- On verification failure (return `verified=false` or error), map to 401/403 with envelope and delete session.
- Enforce signCount monotonicity: compare result `newCounter` > stored `sign_count`; on regression, return 409 (code `conflict`) and do not update counter but delete session.
- On success: update credential `sign_count`, create session in DB (`sessions` table) with new expiry (e.g., now + 1 hour), generate cookie payload, and delete login session.
- Hash identifiers for logging (reuse helper e.g., `hashIdentifier` or registration equivalent) and emit `login_finish` log with `account_thumb_hex`, `credential_id_hash`, `sign_count`, etc.

Database interactions

- Read from `credentials` (credential_id, acct_cbor_fk, sign_count); join with `accounts` if needed for thumb hashing.
- Update `credentials.set(sign_count = newCounter)`.
- Insert into `sessions(session_id, acct_cbor, expires_at, created_at)` with TTL from config (e.g., 1 hour).
- Wrap writes in transaction or sequential operations; ensure rollback on failure.

Policies & limits

- Session TTL: login session 5 minutes (already enforced); auth sessions (cookie) follow Step 4 TTL (e.g., 1 hour) and HttpOnly+Secure+SameSite policy.
- UV required: `requireUserVerification: true`; enforce `userVerified` flag.
- Attestation: not relevant for login, but ensure origin/RP allowlists enforced via existing helper or config list.
- Duplicate/failure handling: the login session is single-use; delete after success or terminal failure; align error codes with Go server.

Sequencing

- Depends on Steps 5–7 (registration options/finish, login options) complete; expects existing credentials.
- After this step, Step 9 (account key) will rely on session cookie established here.
- Logging fields must align with Step 15 structured logging requirements.
- Coordinate with frontend to ensure finish payload uses new JSON and handles errors gracefully.

Tests

- `node-server/test/login-finish.test.js` (Node test runner):
  - *Happy path*: stub `verifyAuthenticationResponse` (or use real?) to succeed, ensure DB updates, cookie set, session deleted, log fields present.
  - *Expired/missing session*: expect 401 and session removal.
  - *Unknown credential*: expect 401.
  - *UV missing/verification failure*: expect 403/401 with envelope.
  - *SignCount regression*: expect 409 `conflict`, no counter update, session deleted.
  - *Duplicate session use*: reuse same login_session_id to ensure single-use behavior.
  - *Invalid payload*: missing fields => 400.
- Update or add frontend tests (Playwright/E2E) to ensure login completes with Node server once finish implemented.
- Commands: `npm -C node-server test`; run relevant web tests (`npm -C web test:ui` or targeted scripts).

Verification

- Automated: `npm -C node-server test` (includes new finish tests).
- Manual: start Node server, perform register+login via web app or CLI to verify cookie and DB rows; inspect logs.
- Cookie check: confirm `Set-Cookie` attributes align with Step 4 (e.g., `Secure`, `HttpOnly`, `SameSite=None`, `Max-Age`).

User verification commands

```bash
# Run backend unit tests
npm -C node-server test

# Manual smoke (after registering a credential)
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=node-server/demo.db   node node-server/src/server.js &
API_PID=$!
# Use browser or curl with saved login_session_id + assertion to hit /authn/passkey/login/finish
# (requires captured payload)
kill $API_PID || true

# Frontend tests (optional)
npm -C web test:ui  # Playwright E2E if available
```

Acceptance criteria

- Login finish validates assertions, updates credential `sign_count`, creates server session (cookie + DB row), logs `login_finish`, and returns 200 JSON.
- Error cases handled with correct envelopes/statuses (401/403/409/400) and sessions marked single-use.
- Frontend successfully completes login against Node server using SimpleWebAuthn flows.
- Blueprint/spec/description updates reflect the new behavior.

Notes

- Reuse helpers from registration (hashing, logging) where possible to maintain consistency.
- Consider extracting common session store logic into shared util only if it reduces duplication without complicating Step timelines (otherwise capture as future refactor).
- Ensure new tests stub `verifyAuthenticationResponse` deterministically to avoid requiring cryptographic operations.

Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope; decision request-id-and-slog-json

