# Step 6 — Node: Registration finish (Done)

### Step 6 — Node: Registration finish (`/authn/passkey/registration/finish`)

Scope

- Implement the finish endpoint so passkey registrations persist accounts/credentials identically to the Go backend (canonical COSE, thumb hash, signCount) and consume the Step 5 session.
- Enforce ceremony policies: challenge/origin/RP ID validation, UV required, attestation fmt `none`, duplicate credential protection, and single-use session semantics.
- Emit `reg_finish` structured logs mirroring Go fields (`account_thumb_hex`, `credential_id_hash`, `sign_count`).

Source to add/modify

- Extend `node-server/src/webauthn/reg.js` with the finish handler, helper functions (COSE canonicalization, thumb hashing), and store accessors.
- Update `node-server/src/server.js` to pass the shared `RegistrationSessionStore` instance (from Step 5) into the finish handler when wiring routes.
- Add `node-server/test/reg-finish.test.js` covering success, policy failures, duplicate credential conflicts, and session expiry behavior using deterministic stubs.

Description files

- Update `node-server/src/webauthn/reg.js.desc.md` to document the finish handler logic, logging, DB writes, and error mappings.
- Update `node-server/src/webauthn/webauthn.desc.md` with finish-handler responsibilities and shared helpers (COSE canonicalization, thumb hashing).
- Update `node-server/src/server.js.desc.md` to reflect wiring of both registration routes (options + finish) and shared store export through `app.locals`.
- Add `node-server/test/reg-finish.test.js.desc.md` summarizing scenario coverage and stubbing strategy.

Blueprint updates

- Update `blueprint/features/registration-flow/_specs/spec.md` (Draft) with details on finish verification: single-use sessions, canonical COSE storage, thumb hashing (`SHA-256("ACCTK1" || acct_cbor)`), logging fields, and error mappings.
- Update `blueprint/features/registration-flow/_specs/frontend-mapping-and-pitfalls.md` to note the client must POST `reg_session_id` alongside the SimpleWebAuthn `RegistrationResponseJSON` and handle 401/403/409/backend errors.
- Confirm `blueprint/features/registration-flow/requirement.md` already reflects persistence/UV policies; append note clarifying duplicate credential conflict handling and single-use session consumption if missing (Draft state preserved).
- Evaluate if any ADR needs updates; if finish logic deviates (e.g., new hashing), draft an ADR stub, otherwise state no ADR changes required.

Request/response shape

- Request: JSON body containing `reg_session_id` (string) plus the SimpleWebAuthn `RegistrationResponseJSON` fields: `id`, `rawId`, `type`, and `response` `{ attestationObject: base64url string, clientDataJSON: base64url string }`.
- Response: `201 Created` with `Content-Type: application/json` body `{ account_thumb_hex: string (lowercase hex), credential_id_b64: string (base64url) }`.
- Error responses: JSON envelope `{ code, error, correlation_id? }` via `writeError`; map to statuses 400, 401, 403, 409, 500.

Algorithm

- Parse request JSON, requiring `reg_session_id` and `response` fields; reject missing/invalid types with 400.
- Lookup session via `RegistrationSessionStore.get`; if absent or expired (`expiresAt <= now`), delete stale session and return 401 (code `unauthorized`).
- Call `verifyRegistrationResponse` with:
  - `expectedChallenge = session.challenge`,
  - `expectedOrigin = config.ORIGIN` (include allowlist handling),
  - `expectedRPID = config.RP_ID`,
  - `requireUserVerification = true`,
  - `supportedAlgorithmIDs = [-7]`,
  - `attestationType = 'none'` (reject others).
- Validate verification result: `verified === true`; ensure `registrationInfo` includes `credentialPublicKey`, `credentialID`, `counter`, `aaguid`; enforce fmt `'none'` if available.
- Canonicalize `credentialPublicKey` into CBOR (using `cbor-x` encoder with canonical sorts). Compute `acct_thumb = SHA-256('ACCTK1' || acct_cbor)`.
- Perform DB operations within transaction-style try/finally:
  1. `INSERT OR IGNORE` `accounts(acct_cbor, acct_thumb, created_at=nowEpochSeconds)`.
  2. `INSERT` `credentials(credential_id, acct_cbor_fk, sign_count, aaguid, created_at)`; on UNIQUE violation, return 409 (code `conflict`).
- Delete registration session after successful persistence (single-use).
- Log success via `logger.info({ event: 'reg_finish', correlation_id: req.id, account_thumb_hex, credential_id_hash, sign_count })` where `credential_id_hash = sha256('CIDv1'||credentialID)` reused from Go (define helper to match Step 15 expectations).
- Respond with 201 JSON payload.
- Handle errors: map verification failures to 400/403/401; map DB errors to 500 unless duplicate; ensure thrown exceptions log with `logger.error({ event: 'reg_finish_error', ... })` then envelope.

Database interactions

- `INSERT OR IGNORE INTO accounts` followed by `INSERT INTO credentials`; ensure `acct_cbor_fk` matches inserted account key.
- Use `better-sqlite3` prepared statements for atomic writes inside try/catch; optionally wrap in `db.transaction` helper for clarity.
- On duplicate credential, do not delete session (still consumed?); plan to delete session to prevent reuse.

Policies & limits

- UV required: `requireUserVerification: true` plus verifying `registrationInfo.credentialDeviceType` if available; reject if `verified` false or `registrationInfo.userVerified` false via 403.
- Attestation limited to `'none'`; if library surfaces other attestation formats, return 400/403 (policy violation).
- Enforce same RP/origin allowlists as Go (use `config.RP_ID_ALLOWLIST`/`ORIGIN_ALLOWLIST` via custom checks if `verifyRegistrationResponse` doesn’t cover allowlists).
- Session TTL: treat expired sessions as unauthorized and delete entry.
- DB stores raw binary blobs; size/entropy expectations (credential IDs up to 1024 bytes) - optionally enforce max length (reject >1024 with 400).

Sequencing

- Depends on Step 5 (options + session store) complete; finish handler must reuse the same store via `app.locals.registration.store` or dependency injection.
- Future Step 15 logging work expects consistent event fields; keep naming consistent.
- Step 8 (login finish) will reuse DB helper utilities (thumb hashing) – consider exporting helpers while avoiding premature abstraction.

Tests

- `node-server/test/reg-finish.test.js` (Node test runner):
  - _Happy path_: stub `verifyRegistrationResponse` to return successful payload; assert 201, JSON response fields, session deleted, DB rows created with canonical CBOR + thumb, log event emitted (use pino logger stub or `logger.info` spy).
  - _Expired/missing session_: create store entry with past `expiresAt`; expect 401 and that session is removed.
  - _Challenge mismatch / verification failure_: stub verifier to return `{ verified: false }` or throw custom error; expect 401 or 400 mapping.
  - _UV missing_: stub result with `userVerified: false` to ensure 403.
  - _Duplicate credential_: seed DB with existing credential, expect 409 conflict and session consumed.
  - _Invalid request body_: send missing `response` to assert 400.
- Use deterministic `idFactory`/clock from Step 5 store and in-memory SQLite (`:memory:`) for DB assertions. Execute tests via `npm -C node-server test`.

Verification

- `npm -C node-server test` includes new `reg-finish` suite plus existing tests; rerun after fixes until green.
- Manual smoke: run Node server, perform registration flow with web front-end (requires future web Step) or simulate using stored session + cURL by posting captured `RegistrationResponseJSON` (document placeholder until end-to-end wiring) to confirm 201 + DB rows.
- Inspect SQLite file (e.g., `sqlite3 server/demo-node.db 'SELECT COUNT(*) FROM accounts;'`) to verify persistence after manual run.

User verification commands

```bash
# Run unit tests (rerun after fixes until green)
npm -C node-server test

# Manual smoke (requires previously obtained response JSON)
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo-node.db \
  node node-server/src/server.js &
API_PID=$!
# TODO: Replace payload.json with captured browser response
curl -sS -X POST http://127.0.0.1:8080/authn/passkey/registration/options | jq '.reg_session_id' > /tmp/reg_session_id
# ...invoke browser to create response...
# curl -sS -X POST http://127.0.0.1:8080/authn/passkey/registration/finish \
#   -H 'Content-Type: application/json' \
#   --data @payload.json | jq
kill $API_PID || true
```

Acceptance criteria

- Registration finish verifies ceremonies, stores canonical account + credential rows, emits structured logs, and returns 201 with thumb + credential id.
- Duplicate credential attempts surface 409 conflicts; expired or invalid sessions yield 401; UV/origin/RP mismatches map to 403/400; errors use JSON envelope with correlation IDs.
- Session entries are single-use and removed on both success and terminal failure paths.

Notes

- Prefer helper functions for hashing/logging to share with future login/sign flows while keeping Step 6 focused on registration.
- Ensure description docs and specs stay synchronized with any helper exports; capture refactor candidates under `blueprint/_refactor` if non-critical abstractions arise.

Refs: goal passkey-registration-login-uv; goal webauthn-policy-defaults; requirement R-FLOW-REG; requirement R-SEC-UV; spec R-FLOW-REG/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope; decision request-id-and-slog-json

---


