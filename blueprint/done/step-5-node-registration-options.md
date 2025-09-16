# Step 5 — Node: Registration options (Done)

### Step 5 — Node: Registration options (`/authn/passkey/registration/options`)

Scope

- Implement `POST /authn/passkey/registration/options` with `@simplewebauthn/server` so the browser receives native JSON that honors demo policies.
- Back the endpoint with an in-memory registration-session store containing `{ challenge, rpID, origin, expires_at }` and a fixed 5 minute TTL, exposing helpers for Step 6 to reuse.
- Emit a `reg_options` structured log on success containing correlation id, rp/origin, expiry, and session id length.

Source to add/modify

- Add `node-server/src/webauthn/reg.js` defining the registration session store, dependency-injected handler, and router mounting `POST /registration/options`.
- Modify `node-server/src/server.js` to import and mount the registration router under `/authn/passkey/registration` prior to other route placeholders.
- Add `node-server/test/reg-options.test.js` covering happy path shape policies, TTL pruning, error envelope handling, and session id randomness.

Description files

- Add `node-server/src/webauthn/webauthn.desc.md` describing shared WebAuthn helpers and session flow (Refs: goal passkey-registration-login-uv; requirement R-FLOW-REG; decision webauthn-corrections-and-standardizations).
- Add `node-server/src/webauthn/reg.js.desc.md` summarizing the handler, store semantics, logging fields, and dependencies (Refs: same as above).
- Update `node-server/src/server.js.desc.md` to record the new router wiring and emitted log events (Refs: requirement R-FLOW-REG; decision request-id-and-slog-json).
- Add `node-server/test/reg-options.test.js.desc.md` outlining coverage for happy/negative paths (Refs: requirement R-FLOW-REG; spec R-FLOW-REG/spec.md).

Blueprint updates

- Update `blueprint/features/registration-flow/_specs/spec.md` (still Draft) to include Node-specific details: `reg_session_id`, TTL 300s, logging expectations, and reliance on `@simplewebauthn/server` JSON helpers.
- Update `blueprint/features/registration-flow/_specs/frontend-mapping-and-pitfalls.md` to explicitly call out the response fields `reg_session_id` and `expires_at` the web app must echo during finish.
- Amend `blueprint/features/registration-flow/requirement.md` Acceptance Criteria if needed with an explicit 5 minute registration session expiry note (remains Draft but clarifies TTL policy).

Request/response shape

- Request: `POST /authn/passkey/registration/options` with empty JSON body and `Accept: application/json`.
- Response: 200 JSON with `PublicKeyCredentialCreationOptionsJSON` merged with:
  - `reg_session_id`: string, 24 char base64url from `nanoid` (>=128 bits entropy).
  - `expires_at`: integer epoch seconds `Math.floor(now/1000) + 300`.
  - `rp`: `{ id: config.RP_ID, name: 'Passkey Demo' }`.
  - `user`: ephemeral placeholder `{ id: base64url(32 bytes), name: 'passkey-user', displayName: 'Passkey User' }`.
  - `challenge`: base64url (32 bytes entropy).
  - `pubKeyCredParams`: `[ { type: 'public-key', alg: -7 } ]`.
  - `attestation`: `'none'`.
  - `authenticatorSelection`: `{ residentKey: 'required', requireResidentKey: true, userVerification: 'required' }`.
  - `timeout`: 60000 (ms) mirroring Go implementation for parity.

Algorithm

- Build a `RegistrationSessionStore` (Map keyed by session id) with helpers `createSession`, `getSession`, `deleteSession`, and `pruneExpired(nowEpochSeconds)`; inject `now` for tests.
- Handle POST `/registration/options`:
  1. Call `store.pruneExpired(nowEpochSeconds)` before generating a new session.
  2. Generate `sessionId` via injected `idFactory()` (default `nanoid` returning base64url) and compute `expiresAt = nowEpochSeconds + 300`.
  3. Call `generateRegistrationOptions` with config rp/origin, constant `rpName = 'Passkey Demo'`, `authenticatorSelection` enforcing UV + resident key, `attestationType: 'none'`, `supportedAlgorithmIDs: [-7]`, and ephemeral user placeholders.
  4. Persist `{ challenge: options.challenge, rpID: config.RP_ID, origin: config.ORIGIN, expiresAt }` in the store; if an id collision occurs, regenerate up to 3 attempts before returning 500.
  5. Return merged JSON via `res.status(200).json({ ...options, reg_session_id: sessionId, expires_at: expiresAt })`.
  6. Log `logger.info({ event: 'reg_options', correlation_id: req.id, rp_id: config.RP_ID, origin: config.ORIGIN, expires_at: expiresAt, session_id_len: sessionId.length })`.
- On errors from option generation or store writes, respond with `writeError(res, 500, 'internal_error', 'Internal server error', req.id)`.

Database interactions

- None; ensure implementation avoids touching SQLite, leaving persistence for Step 6.

Policies & limits

- Session TTL fixed at 300 seconds; store must reject/clean expired entries and expose pruning for periodic cleanup.
- Enforce `residentKey: 'required'`, `requireResidentKey: true`, `userVerification: 'required'`, and `attestation: 'none'` on every response.
- Restrict `pubKeyCredParams` to ES256 (`alg: -7`).
- Session id must be >=128 bits entropy, not leaked beyond body (logs include only length) and stored server-side.

Sequencing

- Relies on middleware from Steps 2–4 (request id, JSON body parser, CORS, session loader) already mounted in `createApp`.
- Step 6 will reuse the session store and logging interface; design exports (`createRegistrationRoutes`, `createRegistrationSessionStore`) so finish handler can import without refactor.
- Structured logging emitted now should align with Step 15’s logging consolidation; keep field names stable.

Tests

- `node-server/test/reg-options.test.js` (Node test runner):
  - *Happy path*: stub `now` and `idFactory` for determinism, POST endpoint, assert 200, JSON policy fields, and store entry matching TTL + challenge.
  - *TTL pruning*: create session, advance `now` beyond expiry, call `pruneExpired`, assert session removed and new request yields fresh session id.
  - *Error handling*: inject handler with stub `generateRegistrationOptions` throwing to assert 500 with `{ code: 'internal_error', correlation_id: <req.id> }`.
  - *Entropy/regeneration*: two sequential POSTs with default id generator produce distinct 24 char base64url ids.
- Execute via `npm -C node-server test`; rerun after any fixes until passing.

Verification

- Run `npm -C node-server test` and confirm all tests succeed; rerun after addressing failures (fix-forward loop).
- Manual smoke: start Node server with demo env vars, curl the endpoint, check JSON for policy flags (`residentKey`, `userVerification`, `attestation`) and presence of `reg_session_id`/`expires_at`.
- Confirm server log includes a `reg_options` event with correlation id matching `X-Request-ID` header.

User verification commands

```bash
# Run unit tests (rerun after fixes until green)
npm -C node-server test

# Manual curl against dev server
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=node-server/demo.db \
  node node-server/src/server.js &
API_PID=$!
sleep 1
curl -sS -X POST http://127.0.0.1:8080/authn/passkey/registration/options | jq
kill $API_PID || true
```

Acceptance criteria

- Endpoint returns library-ready JSON enforcing UV/resident key/attestation policies and includes session metadata for diagnostics.
- Sessions persist in-memory with 5 minute TTL, logged via `reg_options`, and are retrievable for Step 6.
- Failure paths surface `code: internal_error` with correlation id via envelope helper; no DB writes occur.

Notes

- Keep session store exported and pure so Step 6 can add finish logic without refactoring; inject dependencies (clock/generator) for deterministic tests.
- Align manual logs/error strings with planned Step 15 logging polish to minimize churn.

Refs: goal passkey-registration-login-uv; goal webauthn-policy-defaults; requirement R-FLOW-REG; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations; decision request-id-and-slog-json

---


