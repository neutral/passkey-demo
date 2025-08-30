# R-PLAT-2 — Backend Service Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Single Go HTTP service providing WebAuthn ceremonies and transaction signing endpoints; synchronous processing; no brokers.

## Interfaces
- REST endpoints as listed in requirement; cookie-based session.

## Data / Models
- SQLite schema for accounts, credentials, sessions, transactions (see R-PLAT-3 spec/schema).

## Algorithms
- WebAuthn option issuance; assertion verification (UV, rpIdHash, origin, low‑S, signCount); signing challenge derivation from CBOR bundle; tx_id derivation.

## HTTP API Examples
- Registration options `POST /authn/passkey/registration/options`
```json
{
  "rp_id": "localhost",
  "origin": "http://localhost:5173",
  "uv_required": true,
  "attestation": "none"
}
```
- Registration finish `POST /authn/passkey/registration/finish` response `201 Created`
```json
{
  "account_thumb_hex": "ab12...fe",
  "account_cbor_b64": "<b64>",
  "credential_id_b64": "<b64>",
  "sign_count": 0
}
```
- Login options `POST /authn/passkey/login/options` response
```json
{
  "login_session_id": "string",
  "publicKey": {
    "challenge": "<b64url 32B random>",
    "rpId": "localhost",
    "userVerification": "required",
    "allowCredentials": []
  },
  "expires_at": 1735689600
}
```
- Login finish `POST /authn/passkey/login/finish` response `200 OK`
```json
{
  "session_id": "opaque-token",
  "account_thumb_hex": "ab12...fe"
}
```
- Signing options `POST /tx/signing/options` request/response
```json
{
  "bundle_cbor_b64": "<b64>",
  "rp_id": "localhost",
  "origin": "http://localhost:5173"
}
```
```json
{
  "tx_session_id": "string",
  "publicKey": {
    "challenge": "<b64url of SHA-256('CHALv1'||B)>",
    "rpId": "localhost",
    "userVerification": "required",
    "allowCredentials": ["<b64url credentialId>"]
  },
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
- Origin/rpId validation; session binding for options; binary treated opaque; minimal logs.

## WebAuthn Verification Details
- ClientDataJSON: `type` matches ceremony; `challenge` equals server-stored; `origin` allowlisted.
- AuthenticatorData: `rpIdHash = SHA-256(rp_id)`; UV flag set; `signCount` strictly increases.
- Signature: `digest = SHA256(AD || SHA256(CDJ))`; DER parse; enforce low‑S; verify with account COSE key.

## Errors / Observability
- Use 400/401/403/409/413/429/5xx appropriately; structured error responses.

## Testing Strategy
- Local E2E with platform authenticator; unit checks for verification helpers.

## Open Questions
- TLS/offload for non-dev deployments (out of scope for demo).

Refs: decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; requirement R-PLAT-2
