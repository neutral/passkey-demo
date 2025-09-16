# Step 12 — Node: Tx signing finish (`/tx/signing/finish`) (Done)

Completed: 2025-09-16
Verification notes:
- `cd node-server && node --test test/tx-finish.test.js`
- `cd node-server && npm test`
- `bash tools/desc-check.sh HEAD~1 HEAD`

### Step 12 — Node: Tx signing finish (`/tx/signing/finish`)

Scope
- Add `POST /tx/signing/finish` handler that consumes a tx session created in Step 11, verifies the WebAuthn assertion (`verifyAuthenticationResponse`), enforces UV/origin/rpId policies, and persists the signed transaction (bundle `B`, anchors, authenticator data, client data, signature).
- On success, delete the tx session, update credential `sign_count`, insert into `transactions`, and return `201` with `tx_id_hex`; log `tx_finish`/`tx_finish_error` events and map failures to the standardized error envelope.
- Maintain parity with the Go implementation for nonce/low-S enforcement, signCount monotonicity, and error semantics.

Source to add/modify
- `node-server/src/tx/finish.js` (new) — Express route logic for `/tx/signing/finish`, including session lookup/removal, WebAuthn verification, DB writes, and error mapping.
- `node-server/src/tx/options.js` — export `TxSessionStore` retrieval helpers (e.g., `get`, `delete`) if additional APIs needed for finish step.
- `node-server/src/server.js` — mount finish handler (either extend existing `/tx` router or import new module).
- `node-server/src/logger.js` — add `logTxFinishSuccess` / `logTxFinishError` helpers to emit structured logs.
- `node-server/src/error.js` — ensure envelope codes cover new cases (401 unauthorized session/signature mismatch, 403 policy violations, 409 nonce/signCount issues, 500 internal).

Description files
- `node-server/src/tx/finish.js.desc.md` (new) — describe verification flow, session handling, DB writes, and error mapping.
- `node-server/src/tx/options.js.desc.md` — note shared session store operations consumed by finish handler.
- `node-server/src/server.js.desc.md` — update to reference `/tx/signing/finish` wiring.
- `node-server/src/logger.js.desc.md` — document new logging helpers/events.
- `node-server/src/tx/tx.desc.md` — expand overview to include finish handler responsibilities.

Blueprint updates
- None expected; existing R-FLOW-SIGN spec covers signing finish behavior. Confirm parity while implementing.

Request/response shape
- Request JSON: `{ tx_session_id: string, response: AuthenticationResponseJSON fields... }` (mirrors web payload) — e.g., includes `id`, `rawId`, `response.authenticatorData`, `response.clientDataJSON`, `response.signature`, `response.userHandle`.
- Success (`201 Created`): `{ tx_id_hex: string }`.
- Error envelopes:
  - `401 ERR_UNAUTHORIZED` — unknown/expired session, signature mismatch, credential mismatch.
  - `403 ERR_FORBIDDEN` — UV missing, rpId/origin policy violations.
  - `409 ERR_CONFLICT` — nonce not increasing or signCount regression.
  - `400 ERR_BAD_REQUEST` — malformed JSON/base64/CBOR, missing fields.
  - `500 ERR_INTERNAL` — unexpected DB or verification errors.

Algorithm
- Ensure `req.session`/authenticated account present; reject otherwise (401).
- Parse JSON payload; validate `tx_session_id` string and presence of WebAuthn response fields; base64url-decode `rawId`, `authenticatorData`, `clientDataJSON`, `signature`.
- Look up tx session from `TxSessionStore` (Step 11). If missing or expired → 401; delete on success path to prevent reuse.
- Fetch credential row by `credential_id` ensuring account binding; load account data for verification (COSE key, current sign_count, transports).
- Invoke `verifyAuthenticationResponse` with expected challenge (`session.challenge`), `config.ORIGIN`, `config.RP_ID`, requiring UV and matching rpIdHash/origin.
- Enforce policy from decisions: UV must be true; reject high-S signatures if policy demands; validate signature counter monotonicity (409 on regression).
- Update credential `sign_count` with new counter from verification; wrap in transaction when inserting new record to ensure atomicity.
- Decode canonical bundle `B` (from session) back to logical structure to extract `nonce`, `message`, `valid_until`; compute `tx_id` (should match session `txId`) and guard against mismatch.
- Insert transaction row (`transactions`) with canonical bundle bytes, anchors, AD, CDJ, signature, `created_at = nowSeconds`.
- Log `tx_finish` success with correlation id, `tx_id_hex`, `nonce`, `account_thumb_hex`, `credential_id_hash`, `sign_count`.
- Map errors to envelopes: 401 for missing/invalid session or signature mismatch; 403 for UV/origin failures; 409 for nonce/sign_count; 400 for malformed payload; 500 for unexpected issues.

Database interactions
- `SELECT credential_id, sign_count, transports FROM credentials WHERE credential_id = ? AND acct_cbor_fk = ?`.
- `UPDATE credentials SET sign_count = ? WHERE credential_id = ?`.
- `INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`.
- Use transactions to ensure counter update + insert are atomic (BEGIN/COMMIT, rollback on failure).

Policies & limits
- Session must be authenticated and tx session valid; reject reuse by deleting after finish attempts.
- Enforce UV required, rpId/origin checks, and low-S acceptance per ADR (`web-authn-accept-high-s-signing-too`).
- Nonce monotonicity guaranteed by Step 10 helper; double-check before insert (409 on violation).
- Sign count must strictly increase; treat regression as conflict.
- Avoid logging raw payloads; only emit hashes/thumbs.

Sequencing
- Depends on Step 11’s router/store; finish handler must share the same `TxSessionStore` instance.
- Prepares data for Step 13 (tx list) by inserting rows with proper schema.
- Error envelopes align with Step 14 refactor; keep `writeError` usage centralized.

Tests
- Add `node-server/test/tx-finish.test.js` covering:
  - Happy path: seeded session + credential + successful verification stub → 201, DB row inserted, session removed, counter updated.
  - Missing session id or expired session → 401 envelope.
  - Malformed payload/base64 → 400.
  - Signature mismatch (stub `verifyAuthenticationResponse` to throw) → 401.
  - UV missing/origin mismatch (stub returns `verified:false` or missing UV) → 403.
  - Sign count regression (stub returns lower counter) → 409 and no DB insert.
  - Nonce conflict (pre-existing transaction with same nonce) → 409.
  - Session reuse attempt (second POST with same id) → 401/409 and store cleared.
- Tests use in-memory SQLite + migrations, deterministic session store, and mocked `verifyAuthenticationResponse` to control behavior.
- Commands: `cd node-server && node --test test/tx-finish.test.js`; run full `npm test` afterward.

Verification
- Execute targeted `node --test test/tx-finish.test.js` until green.
- Run `npm test` in `node-server` to ensure full suite passes.
- Optionally exercise manual flow (options → finish) via curl/postman once implemented.
- Maintain description policy with `bash tools/desc-check.sh HEAD~1 HEAD`.

User verification commands
```bash
cd node-server
node --test test/tx-finish.test.js
npm test
cd ..
bash tools/desc-check.sh HEAD~1 HEAD
```

Acceptance criteria
- Finish endpoint verifies assertions against stored tx sessions, persists transaction rows, updates credential counters, and returns 201 with `tx_id_hex`.
- Error responses/logs align with Go parity for all negative flows.
- Automated tests (unit + full suite) pass and cover happy and edge cases.

Notes
- Ensure tx sessions are removed after success or terminal errors to prevent replay.
- Reuse shared helpers for base64/canonical bundle decoding to avoid duplication.
- When inserting DB rows, store binary fields as buffers (no JSON string conversions).

Refs: requirement R-FLOW-SIGN; decision webauthn-accept-high-s-signing-too; decision http-error-envelope; goal server-derived-challenge-and-txid; decision encoding-and-ceremony-guardrails
