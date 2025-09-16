# Step 15 — Node: Structured logging and correlation (Done)

Completed: 2025-09-16
Verification notes:
- `cd node-server && node --test test/logger-events.test.js`
- `cd node-server && node --test test/login-finish.test.js`
- `cd node-server && node --test test/tx-finish.test.js`
- `cd node-server && node --test test/reg-options.test.js`
- `cd node-server && node --test test/tx-options.test.js`
- `cd node-server && npm test`
- `bash tools/desc-check.sh HEAD~1 HEAD`

### Step 15 — Node: Structured logging and correlation

Scope
- Emit structured JSON logs for every major ceremony (`server_start`, `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`) and verification failures (`webauthn_assert_verify`) with correlation ids, rp/origin metadata, and hashed identifiers.
- Ensure success and failure logs include consistent fields (e.g., `expires_at`, `allow_credentials_count`, `account_thumb_hex`, `credential_id_hash`, `tx_id_hex`, `nonce`, `error_kind`, `uv`, `up`) without exposing raw secrets.

Source to add/modify
- `node-server/src/logger.js`: expand helper surface (`logServerStart`, `logRegOptions`, `logRegFinish`, `logLoginOptions`, `logLoginFinishSuccess`, `logLoginFinishFailure`, `logTxOptions`, `logTxFinishSuccess`, `logTxFinishFailure`, `logWebauthnVerifyFailure`) that automatically attach correlation ids and normalize payloads (including hashing identifiers).
- `node-server/src/server.js`: emit a single `server_start` log after configuration load, document/env-wire `LOG_LEVEL` (and optional `LOG_FORMAT`), and confirm request-id middleware precedes logging so helpers observe correlation ids.
- `node-server/src/webauthn/reg.js` & `node-server/src/webauthn/login.js`: replace inline `logger.info/error` usage with helper calls in options and finish handlers; compute hashed identifiers once; invoke `logWebauthnVerifyFailure` when SimpleWebAuthn verification fails with mapped `error_kind`, `uv`, `rp_id`, `origin`.
- `node-server/src/tx/options.js`, `node-server/src/tx/finish.js`, `node-server/src/tx/list.js`, `node-server/src/me.js`: adopt helpers for success/error logging, ensuring verification failures emit `webauthn_assert_verify` exactly once and list/me logs include correlation ids.
- Tests: add `node-server/test/logger-events.test.js` covering helper output; update existing suites (`reg-options.test.js`, `reg-finish.test.js`, `login-options.test.js`, `login-finish.test.js`, `tx-options.test.js`, `tx-finish.test.js`) to spy on helper calls for happy paths and failure scenarios; optionally add `node-server/test/helpers/logger-spy.js` to share spy/reset logic.

Description files
- `node-server/src/logger.js.desc.md` — document expanded helper API, event schema, env controls, and privacy guarantees.
- `node-server/src/server.js.desc.md` — note `server_start` logging and correlation-id propagation.
- Update flow docs (`node-server/src/webauthn/reg.js.desc.md`, `node-server/src/webauthn/login.js.desc.md`, `node-server/src/tx/options.js.desc.md`, `node-server/src/tx/finish.js.desc.md`, `node-server/src/tx/list.js.desc.md`, `node-server/src/me.js.desc.md`, `node-server/src/tx/tx.desc.md`) to enumerate emitted events, hashed fields, and guarantees about single success/failure logs.

Blueprint updates
- Review `blueprint/global/single-go-backend/_specs/spec.md` (Observability section) and `server/server.desc.md` for parity; append clarifications if Node logging introduces env knobs or attribute differences.
- No ADR changes expected; ensure updated docs retain `Refs` to existing logging decisions.

Request/response shape
- No API response changes. Log entries must be JSON objects containing at minimum `event`, `correlation_id`, `rp_id`, `origin`, plus: 
  - Success fields: `expires_at`, `session_id_len`, `allow_credentials_count`, `account_thumb_hex`, `credential_id_hash`, `sign_count`, `tx_id_hex`, `nonce` (as applicable).
  - Failure fields: `error_kind`, `reason`, `uv`, `up`, hashed identifiers (`credential_id_hash`, `account_thumb_hex`, `tx_id_hex`).

Algorithm
- Augment helpers to inject correlation id/rp/origin defaults, hash identifiers, and choose log level (`info` for success, `error` for failure) before calling `logger.info/error`.
- Refactor handlers so each success path invokes the appropriate helper exactly once after DB persistence; failure branches call a helper before returning an error response.
- Map validation/verification errors to canonical `error_kind` strings reused in logs and HTTP envelopes to keep parity with the Go backend.
- Honour `LOG_LEVEL` (and optional `LOG_FORMAT`) when initializing the logger so local text output remains possible without breaking structured parity.

Database interactions
- None beyond existing persistence; logging wrappers run around existing DB operations and do not introduce new queries.

Policies & limits
- Privacy: never log raw CBOR, credential IDs, signatures, or session secrets; rely on `hashIdentifier`, `computeAccountThumbHex`, and derived fields (`tx_id_hex`).
- Emit at most one success log and one failure log per request to avoid duplication; guard debug output behind level checks to control volume.

Sequencing
- Expand helper module first, then migrate WebAuthn handlers followed by transaction/list/me handlers.
- Update tests immediately after each handler migration to keep assertions aligned.
- Refresh description/spec docs after code and tests stabilize.

Tests
- `node-server/test/logger-events.test.js`: verify helper outputs (correlation id injection, hashed identifiers, event names, error_kind mapping).
- Update existing suites to spy on helper exports for happy-path and failure scenarios (verification failure, policy violation, conflict) ensuring logs fire exactly once.
- Commands: `cd node-server && node --test test/logger-events.test.js`; `cd node-server && node --test test/login-finish.test.js`; `cd node-server && node --test test/tx-finish.test.js`; `cd node-server && npm test`.

Verification
- Automated: run the commands above until all tests pass after addressing any fixes.
- Manual: start the server with `LOG_LEVEL=info` and trigger registration/login/tx flows plus a forced verification failure; inspect stdout (e.g., `tail | jq`) to confirm event names, correlation ids, and hashed identifiers.
- Confirm no sensitive data is logged and that failure logs contain appropriate `error_kind` values.

User verification commands
```bash
cd node-server
LOG_LEVEL=info node src/server.js > /tmp/node-logs.jsonl &
API_PID=$!
sleep 1
curl -sS -X POST http://127.0.0.1:8080/authn/passkey/registration/options || true
curl -sS -X POST http://127.0.0.1:8080/authn/passkey/login/finish -H 'Content-Type: application/json' -d '{}' || true
tail -n 40 /tmp/node-logs.jsonl | jq '{event, correlation_id, account_thumb_hex?, credential_id_hash?, tx_id_hex?, error_kind?}'
kill $API_PID || true
```

Acceptance criteria
- Each flow emits the documented log events with correlation ids and hashed identifiers on success.
- Verification failures produce exactly one `webauthn_assert_verify` log containing `error_kind` aligned with HTTP responses.
- Automated suites and manual inspection confirm structured JSON logs without sensitive leakage.

Notes
- Capture any deferred enhancements (context-aware logging, sampling handlers) as follow-up tasks if scope expands; keep helper modules free of heavy dependencies to avoid circular imports.

Refs: requirement R-ERR; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines
