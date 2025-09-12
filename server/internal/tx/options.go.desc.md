# Purpose
Serve `POST /tx/signing/options` for authenticated users: validate the client-provided Bundle, derive `challenge`/`tx_id`, build WebAuthn assertion options, and create a short-lived tx session.

# Key Logic
- Auth: requires session middleware to populate `{acct_cbor}` in request context (no cookie/DB fallback).
- Validate: `ValidateAndAnchorBundle(ctx, db, acct_cbor, bundle_cbor_b64)` → canonical `B`, `challenge`, `tx_id`.
- Credentials: query all `credential_id` for the account; base64url for `allow_credentials`.
- Session: generate `tx_session_id` (24B random), TTL=5 minutes; store `{B, challenge, acct_cbor, expected credential IDs}` in-memory.
- Response: JSON `{ tx_session_id, challenge, options{rp_id, origin, uv_required, allow_credentials}, tx_id_hex, expires_at }`.

# Observability
- Emits structured logs `tx_options` via `InfoContext` (context handler injects `correlation_id`):
  - Success: `outcome=success`, `tx_id_hex`, `allow_count`.
  - Failures: `outcome=failure` with `reason` in {`invalid_bundle`,`bundle_limits`,`sender_key_mismatch`,`nonce_not_monotonic`,`no_credentials`,`internal_error`} and `account_hash=hex(SHA-256(acct_cbor))`.

- Depends on `internal/tx/bundle.go` for bundle validation and anchors.
- Error mapping: returns JSON error envelope `{code,error,correlation_id?}` with statuses:
  - 400 for invalid bundle/base64/CBOR and bundle limits (`ErrMessageTooLong`, `ErrNonceOutOfRange`).
  - 401 for `ErrSenderKeyMismatch`.
  - 409 for `ErrNonceNotMonotonic` and `ErrNoCredentials`.
- Reads `credentials` in SQLite; writes to in-memory `TxSessionStore`.
- Consumed by the frontend to initiate `navigator.credentials.get(...)` using returned options.

# Refs
Refs: goal server-derived-challenge-and-txid; goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; decision structured-logging-with-slog-guidelines; decision request-id-and-slog-json; decision http-error-envelope
