# Step 7 — Node: Login options (Done)

### Step 7 — Node: Login options (`/authn/passkey/login/options`)

Scope

- Implement the login options endpoint using `generateAuthenticationOptions` so browsers receive SimpleWebAuthn-native JSON that matches the Node registration flow.
- Maintain a login session store (challenge, rpID, origin, expires_at) with 5-minute TTL, shared with the finish handler.
- Emit `login_options` structured logs, mirroring the Go server fields.
- Ensure the web app updates any adapters/fetch logic to the new JSON shape and verifies `login_session_id` handling.

Source to add/modify

- Add `node-server/src/webauthn/login.js` housing `LoginSessionStore`, options handler, and router creation (paralleling registration but reusing exported helpers for logs + hashing where applicable).
- Update `node-server/src/server.js` to mount the login router (`/authn/passkey/login`) and expose the store via `app.locals` for finish logic.
- Extend `node-server/test` with `login-options.test.js` covering happy path, TTL pruning, error scenarios, and entropy checks.
- Update the web client (`web/src/lib/webauthn.ts`, `web/src/pages/Login.tsx`, and any API helpers) to consume the Node JSON directly (no legacy Go reshaping) and ensure `startAuthentication` receives SimpleWebAuthn-compatible options.

Description files

- Add `node-server/src/webauthn/login.js.desc.md` describing session storage, options handler, logs, and dependencies (Refs: goals/requirements/specs for login).
- Update `node-server/src/server.js.desc.md` to reflect login router wiring and shared context.
- Update `node-server/src/webauthn/webauthn.desc.md` for shared helpers extended to login sessions.
- Update `web/src/pages/Login.tsx.desc.md` (create if missing) to reflect Node server JSON expectations.
- Update `web/src/lib/webauthn.ts.desc.md` documenting new mapping logic.

Blueprint updates

- Update `blueprint/features/login-flow/_specs/spec.md` (if exists) to capture login session storage, policies, and SimpleWebAuthn JSON shape; if absent, create/extend spec to match Node behavior.
- Update `blueprint/features/login-flow/requirement.md` acceptance criteria with TTL, session reuse, and logging expectations.
- If necessary, create a brief note in `blueprint/features/login-flow/_specs/frontend-mapping.md` describing frontend adjustments.

Request/response shape

- Request: POST with empty body.
- Response: JSON from `generateAuthenticationOptions` (challenge, allowCredentials, timeout, rpId etc.) plus additional metadata:
  - `login_session_id`: base64url(24 bytes) with ≥128 bits entropy.
  - `expires_at`: epoch seconds (now + 300).
  - Challenge is base64url; allowCredentials omitted for discoverable credentials.

Algorithm

- `LoginSessionStore`: same TTL logic as registration; ensure `pruneExpired` and `delete` exist.
- Handler steps:
  1. Prune expired sessions before issuing new one.
  2. Generate unique session id (retry collisions up to 3 times).
  3. Invoke `generateAuthenticationOptions` with rpId (from config), `userVerification: 'required'`, `timeout` parity, `allowCredentials` possibly empty.
  4. Persist session `{ challenge: options.challenge, rpID, origin, expiresAt }`.
  5. Return merged JSON + metadata.
  6. Log `login_options` with correlation id, rp_id, origin, expires_at, session_id_len and optional allowCredentials count.
- Error handling: wrap generation/persist errors; respond with `writeError(res, 500, 'internal_error', ...)`.

Database interactions

- None (in-memory session store only). Ensure we avoid DB writes in this step.

Policies & limits

- Session TTL 300s; store is in-memory; Step 8 will consume entries.
- `userVerification: 'required'`; `generateAuthenticationOptions` should include `timeout` (match Go value, e.g., 60000ms).
- `allowCredentials` left empty for resident keys; future enhancements may pass hints.
- Session id should not be exposed in logs outside length metrics.

Sequencing

- Requires Step 5/6 (registration and finish) complete to provide accounts/credentials for Step 8.
- Export router/store so Step 8 finish handler can reuse them.
- Ensure logging structure matches Step 15 expectations.
- Web app updates should be coordinated so tests don’t break (update `web/tests` once Node server is wired).

Tests

- `node-server/test/login-options.test.js` (new):
  - *Happy path*: deterministic `now`/`idFactory`, stub `generateAuthenticationOptions`, assert 200 JSON, store entry persists challenge + TTL.
  - *TTL pruning/expiry*: set expired session and ensure pruned.
  - *Entropy/regeneration*: collisions handled, unique session ids.
  - *Error path*: stub generator to throw, expect 500 `internal_error`.
- Update `web` tests (unit/e2e as applicable) to cover new JSON (e.g., `web/tests/login.spec.ts` expect `login_session_id`).
- Run `npm -C node-server test` and relevant web lint/tests (`npm -C web test` if present).

Verification

- Automated: `npm -C node-server test` (includes new login options tests).
- Manual: start Node server, hit `/authn/passkey/login/options` via curl, confirm JSON fields and log.
- End-to-end: ensure web login button obtains options and triggers `startAuthentication` successfully (once finish step is implemented).

User verification commands

```bash
# Run unit tests
npm -C node-server test

# Manual curl of login options
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=node-server/demo.db   node node-server/src/server.js &
API_PID=$!
sleep 1
curl -sS -X POST http://127.0.0.1:8080/authn/passkey/login/options | jq
kill $API_PID || true

# Frontend (after updates)
npm -C web test  # if login unit tests exist
```

Acceptance criteria

- Node returns SimpleWebAuthn-compatible login options with session metadata; store tracks TTL; logs emit `login_options`.
- Web client consumes new JSON shape with no runtime errors; login UX accepts `login_session_id` and forwards during finish (when implemented).
- Tests and manual verification confirm expected behavior; blueprint/specs updated accordingly.

Notes

- Mirror registration options implementation to minimize divergence.
- Document session store reuse between options/finish; consider extracting common helpers if needed but avoid premature refactor.
- Track any follow-up tasks in blueprint if additional adapters or tests are required.

Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope

