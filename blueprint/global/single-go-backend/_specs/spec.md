# R-PLAT-2 — Backend Service Spec

## Metadata
- Status: Approved
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Single Go HTTP service providing WebAuthn ceremonies and transaction signing endpoints; synchronous processing; no brokers.

## Interfaces
- REST endpoints as listed in requirement; cookie-based session with HttpOnly `sid` and session middleware.

## Data / Models
- SQLite schema for accounts, credentials, sessions, transactions (see R-PLAT-3 spec/schema).

## Algorithms
- WebAuthn option issuance; assertion verification (UV, rpIdHash, origin, low‑S, signCount); signing challenge derivation from CBOR bundle; tx_id derivation.

## HTTP API Examples
- Registration options `POST /authn/passkey/registration/options`
```json
{
  "reg_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {
    "rp_id": "localhost",
    "origin": "http://localhost:5173",
    "uv_required": true,
    "attestation": "none"
  },
  "expires_at": 1735689600
}
```
- Registration finish `POST /authn/passkey/registration/finish` response `201 Created`
```json
{ "account_thumb_hex": "ab12...fe", "credential_id_b64": "<b64>" }
```
- Login options `POST /authn/passkey/login/options` response
```json
{
  "login_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {"rp_id": "localhost", "origin": "http://localhost:5173", "uv_required": true, "allow_credentials": []},
  "expires_at": 1735689600
}
```
- Login finish `POST /authn/passkey/login/finish` response `200 OK` (sets `sid` cookie)
```json
{ "account_thumb_hex": "ab12...fe", "credential_id_b64": "<b64>" }
```
- Signing options `POST /tx/signing/options` request/response
```json
{
  "bundle_cbor_b64": "<b64>"
}
```
```json
{
  "tx_session_id": "<b64>",
  "challenge": "<b64>",
  "options": {"rp_id": "localhost", "origin": "http://localhost:5173", "uv_required": true, "allow_credentials": ["<b64>"]},
  "tx_id_hex": "deadbeef...",
  "expires_at": 1735689600
}
```
- Signing finish `POST /tx/signing/finish` response `200 OK`
```json
{
  "tx_id_hex": "deadbeef...",
  "stored": true
}
```
- List transactions `GET /tx/list` response
```json
{
  "items": [
    { "tx_id_hex": "deadbeef...", "nonce": 123, "message": "Hello", "created_at": 1735689600 }
  ]
}
```

## Security / Privacy
- Origin/rpId validation; session binding for options; binary treated opaque.
- Structured logging via `log/slog` with event-style entries and JSON output; privacy guardrails:
  - Do not log raw binary materials (CBOR, signatures, raw credential IDs) or secrets.
  - Use hashed identifiers where correlation is needed (e.g., `credential_id_hash`, `account_thumb_hex`).
  - Include `correlation_id` in logs and error envelopes when Request ID middleware is present.
- Auth: session middleware reads `sid` cookie (HttpOnly, SameSite=Lax, Secure where applicable) and provides account context to protected handlers. Handlers never read cookies directly.

## WebAuthn Verification Details
- ClientDataJSON: `type` matches ceremony; `challenge` equals server-stored; `origin` allowlisted.
- AuthenticatorData: `rpIdHash = SHA-256(rp_id)`; UV flag set; `signCount` strictly increases.
- Signature: `digest = SHA256(AD || SHA256(CDJ))`; DER parse; enforce low‑S; verify with account COSE key.

## Errors / Observability
- Use 400/401/403/409/413/429/5xx appropriately; structured error envelope `{code,error,correlation_id?}`.
- Request ID middleware attaches `X-Request-ID`; envelope includes `correlation_id` when present.
- Emit structured log events for major flows: `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`; verification failures emit `webauthn_assert_verify` with `error_kind`.

## Testing Strategy
- Local E2E with platform authenticator; unit checks for verification helpers.

## Open Questions
- TLS/offload for non-dev deployments (out of scope for demo).

Refs: decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails; decision router-builder-wiring; decision http-error-envelope; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines; requirement R-PLAT-2
