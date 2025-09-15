# R-ID-KEY — Passkey‑First Identity Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Identity model (Model 2): the account identifier is the canonical CBOR bytes of the COSE EC2 public key obtained at registration.

## Interfaces
- Registration finish: parse `attestationObject` to obtain COSE EC2 public key; store canonical CBOR into `accounts`.
- Subsequent flows: resolve account from `credential_id` via `credentials.acct_cbor_fk`.
- GET `/me/account_key`: authenticated session-only endpoint that returns `{ acct_cbor_b64, sender_key: { kty, alg, crv, x, y }, account_thumb_hex, pubkey_x_hex, pubkey_y_hex, created_at }`, enabling the dashboard to rebuild bundles without reshaping.

## Data / Models
- accounts: `acct_cbor (PK BLOB)`, `acct_thumb (BLOB)`, `created_at`.
- credentials: `credential_id (PK BLOB)`, `acct_cbor_fk (BLOB)`, `sign_count`, `aaguid`, `created_at`.

## Algorithms
- Canonicalize COSE key CBOR (deterministic). Compute `acct_thumb = SHA-256("ACCTK1" || acct_cbor)`. Use account COSE key to verify ES256 assertions.

## Security / Privacy
- Single source of truth for public key: store only in `accounts` (no duplication). Treat binary values as opaque; base64url in JSON.
- `/me/account_key` must enforce `sid` cookie authentication and should degrade to 401 when sessions expire or account rows disappear.

## Errors / Observability
- Clear failures on malformed COSE, duplicate accounts, or FK issues; log thumbprints only.

## Testing Strategy
- Round-trip COSE parse→canonical CBOR; account lookup from credential; assertion verification using stored key.

## Open Questions
- None currently for demo scope.

Refs: decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; goal key-first-identity-cose; requirement R-ID-KEY
