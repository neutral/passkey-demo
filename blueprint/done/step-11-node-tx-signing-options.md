# Step 11 — Node: Tx signing options (`/tx/signing/options`) (Done)

Completed: 2025-09-16
Verification notes:
- `cd node-server && node --test test/tx-options.test.js`
- `cd node-server && npm test`
- `bash tools/desc-check.sh HEAD~1 HEAD`

### Step 11 — Node: Tx signing options (`/tx/signing/options`)

Scope
- Implement `POST /tx/signing/options` router that requires authenticated sessions, validates the submitted bundle via `validateAndAnchorBundle`, derives challenge/tx anchors, and returns WebAuthn assertion options with `userVerification: 'required'`.
- Persist a short-lived transaction session (in-memory TTL store) containing canonical bundle bytes, anchors, account CBOR, and allowCredentials list for Step 12 to consume; expose deterministic error envelopes and structured logging.
- Ensure behavior matches Go parity for bundle policies, credential lookup, and error mapping.

Source to add/modify
- `node-server/src/tx/options.js` (new) — Express router + session store utilities (create, get, prune) using bundle helper, logging, and error helpers.
- `node-server/src/server.js` — mount `/tx` router (if absent) after authentication middleware; wire body parser + limits as needed.
- `node-server/src/logger.js` — add `tx_options` / `tx_options_error` helper to emit structured logs with anchors/counts.
- `node-server/src/error.js` — confirm existing envelope helpers cover 400/401/409/500; extend constants if new codes required.
- (Optional) `node-server/src/tx/bundle.js` — only if additional exports (e.g., `ERROR_KINDS`) needed by the handler; otherwise no change.

Description files
- `node-server/src/tx/options.js.desc.md` (new) — document handler responsibilities, session TTL, DB lookups, error mapping, and logging.
- `node-server/src/server.js.desc.md` — update relations to mention `/tx/signing/options` mounting and dependency on session middleware.
- `node-server/src/logger.js.desc.md` — note new logging events and payload fields.
- `node-server/src/error.js.desc.md` — update if additional error codes/status mappings are introduced.

Blueprint updates
- None expected: existing signing requirement/spec already describe options flow; confirm parity during implementation.

Request/response shape
- Request JSON: `{ "bundle_cbor_b64": string }` (base64url canonical bundle from web client).
- Success (`200 OK`):
  ```json
  {
    "tx_session_id": "<base64url>",
    "challenge": "<base64url>",
    "options": {
      "rpId": "example.com",
      "timeout": 60000,
      "userVerification": "required",
      "allowCredentials": [
        { "type": "public-key", "id": "<base64url>", "transports": ["internal"] }
      ]
    },
    "tx_id_hex": "<64 hex chars>",
    "expires_at": 1700000300
  }
  ```
- Error envelopes:
  - `401 ERR_UNAUTHORIZED` when session absent/expired or account mismatch.
  - `400 ERR_BAD_REQUEST` for malformed JSON/base64/CBOR or bundle size violations.
  - `409 ERR_CONFLICT` for nonce monotonicity violations or missing credentials.
  - `500 ERR_INTERNAL` for unexpected failures (DB/read/store errors).

Algorithm
- Require `req.session`; if missing, write 401 envelope and log `tx_options_error`.
- Parse JSON body (ensure upstream body parser handles size limits); validate presence/type of `bundle_cbor_b64`.
- Call `validateAndAnchorBundle(db, session.acct_cbor, bundle)`; catch `BundleValidationError` and map `kind` → status (base64/CBOR/message length → 400; sender mismatch → 401; nonce conflict → 409; etc.).
- Query `credentials` table for all credential IDs tied to the account; if none, return 409 with code `ERR_CONFLICT` to match Go parity.
- Prune expired tx sessions, then generate new `tx_session_id` using base64url random bytes; retry up to N (e.g., 3) on collision or treat as 500 if exhausted.
- Store session payload `{ B, challenge, txId, acct_cbor, credential_ids, expiresAt }` in TTL map keyed by `tx_session_id`.
- Build WebAuthn options object: use config `RP_ID`, `ORIGIN`, forced `userVerification: 'required'`, `allowCredentials` array with base64url credential IDs and default `transports: ['internal']`, `timeout` consistent with spec (60s).
- Log success with `tx_options` event including `tx_id_hex`, `challenge_hash`, `allow_credentials_count`, `account_thumb_hex`, `expires_at`.
- Respond with JSON body including derived challenge (base64url string), `tx_session_id`, `tx_id_hex`, `expires_at`, and sanitized options.

Database interactions
- Read-only query: `SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?` (prepare once, reuse).
- No writes; rely on SQLite for credential enumeration.

Policies & limits
- Session must be authenticated; rely on Step 4 middleware and ensure TTL store honors 5-minute expiry.
- Nonce monotonicity enforced via bundle helper; return 409 conflict on violation.
- Message length ≤1024 bytes and nonce ≤2^53−1 enforced by helper; map to 400.
- Rate limiting/body size will be centralized in Step 14; ensure this handler respects those middlewares (no duplicate enforcement).
- Logging must include correlation id; avoid leaking bundle contents in logs (hash only).

Sequencing
- Depends on Step 10 helper (complete) and earlier auth/session infrastructure; design session store exports so Step 12 can fetch/remove entries.
- Coordinate with planned Step 14 error envelope refactor—reuse shared helpers to prevent duplication.
- Ensure handler gracefully handles presence/absence of Step 15 logging enhancements (use logger helpers introduced here).

Tests
- Add `node-server/test/tx-options.test.js` covering:
  - Happy path: seeded account credentials + valid bundle; assert 200 response, derived challenge equals helper output, session store entry contains expected fields, and allowCredentials matches DB rows.
  - Unauthorized: no session → 401 envelope with `code: 'unauthorized'` (or existing constant).
  - Invalid JSON/base64/CBOR: send malformed body to trigger 400 and confirm `BundleValidationError.kind` mapping.
  - Nonce conflict: seed previous transaction to force helper `NONCE_NOT_MONOTONIC` → 409.
  - No credentials for account → 409 conflict.
  - Session ID collision exhaustion: stub id generator to produce duplicates and assert 500 envelope.
  - TTL pruning: simulate expired session in store before request to verify pruning and new session creation.
- Tests rely on in-memory SQLite via migrations, stubbed randomness/time, and stub `generateAuthenticationOptions` (if used) to focus on challenge override.
- Commands: `cd node-server && node --test test/tx-options.test.js`; run full suite `npm test` after.

Verification
- After implementation run targeted `node --test test/tx-options.test.js`; fix issues until green.
- Execute `npm test` in `node-server/` to ensure full suite passes.
- Optionally run `bash tools/desc-check.sh HEAD~1 HEAD` to confirm description policy compliance.
- Manual curl optional once Step 12 integrated; not required in this step.

Acceptance criteria
- Endpoint returns WebAuthn options with server-derived challenge and stored tx session for follow-up finish step.
- Error envelopes align with Go parity (401/400/409) and include correlation ids; logs capture `tx_options` events.
- Automated tests (unit + full suite) pass and cover happy/negative paths.

Notes
- Keep session store API minimal but accessible to Step 12 (`get`, `take`, `pruneExpired`).
- Use shared base64 utilities to avoid drift; ensure all IDs are base64url without padding.

Refs: requirement R-FLOW-SIGN; requirement R-SEC-UV; goal server-derived-challenge-and-txid; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
