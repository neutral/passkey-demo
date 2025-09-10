# Purpose
Serve `POST /tx/signing/options` for authenticated users: validate the client-provided Bundle, derive `challenge`/`tx_id`, build WebAuthn assertion options, and create a short-lived tx session.

# Key Logic
- Auth: read `sid` cookie; query DB `sessions` to load `{acct_cbor, expires_at}` and verify not expired.
- Validate: `ValidateAndAnchorBundle(ctx, db, acct_cbor, bundle_cbor_b64)` → canonical `B`, `challenge`, `tx_id`.
- Credentials: query all `credential_id` for the account; base64url for `allow_credentials`.
- Session: generate `tx_session_id` (24B random), TTL=5 minutes; store `{B, challenge, acct_cbor, expected credential IDs}` in-memory.
- Response: JSON `{ tx_session_id, challenge, options{rp_id, origin, uv_required, allow_credentials}, tx_id_hex, expires_at }`.

# Observability
- Emits lightweight logs for 401/409 cases to aid triage without exposing raw materials:
  - missing/expired session vs sender_key mismatch vs nonce conflict vs no credentials.
  - Includes `acct_hash=hex(SHA-256(acct_cbor))` or `sid_hash` for correlation.

# Interactions
- Depends on `internal/tx/bundle.go` for bundle validation and anchors.
- Reads `sessions` and `credentials` tables in SQLite; writes to in-memory `TxSessionStore`.
- Consumed by the frontend to initiate `navigator.credentials.get(...)` using returned options.

# Refs
Refs: goal server-derived-challenge-and-txid; goal transaction-content-signing; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations
