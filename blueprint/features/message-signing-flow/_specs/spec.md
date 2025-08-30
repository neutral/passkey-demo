# R-FLOW-SIGN — Message/Transaction Signing Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Server-orchestrated signing bound to content via `challenge = SHA-256("CHALv1" || B)`; transactions identified by `tx_id = SHA-256("TXIDv1" || B)`.

## Interfaces
- POST /tx/signing/options → derives challenge from canonical B; enforces nonce policy
- POST /tx/signing/finish → verifies and stores

## Data / Models
- transactions: `tx_id (BLOB PK)`, `acct_cbor`, `nonce`, `message`, `bundle_cbor`, `auth_data`, `client_data`, `signature`, `created_at`.

## Algorithms
- Canonical CBOR encode Bundle to B; derive challenge+tx_id; store `B` and `challenge` under `tx_session_id`.
- Verify `sender_key` in `B` equals the logged-in account key (prevent cross-account signing).
- Verify ES256 assertion with account key; enforce low‑S; nonce strictly increasing per account.

## Security / Privacy
- UV required; origin/rpId checks; session binding for options.

## Errors / Observability
- 400 on verify failures; 409 on nonce violations; 413 on payload size; 429 on rate limits.

## Testing Strategy
- E2E post-login signing; deterministic CBOR hash consistency; nonce monotonicity checks.

## Open Questions
- Optional valid_until usage and UI exposure.

Refs: decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; goal server-derived-challenge-and-txid; requirement R-FLOW-SIGN
