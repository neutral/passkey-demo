# R-FLOW-SIGN — Message/Transaction Signing Flow

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: fr

## Description
- Post-login, client constructs a canonical CBOR bundle B and sends it to the server to obtain assertion options with `challenge = SHA-256("CHALv1" || B)`. Client performs WebAuthn get and submits assertion; server verifies and stores transaction with `tx_id = SHA-256("TXIDv1" || B)`.

## Depends On
- R-PLAT-1, R-PLAT-2, R-PLAT-3
- R-SCHEMA-LITE, R-ID-KEY, R-SEC-UV

## Scope
- In-scope: options + finish endpoints; canonical CBOR construction on client; nonce monotonicity enforcement; storage of transaction record including raw materials.
- Out-of-scope: multi-sig, complex transaction models.

## Acceptance Criteria
- `/tx/signing/options` recomputes canonical B, enforces per-account nonce strictly increasing, and returns options with derived challenge and allowCredentials.
- `/tx/signing/finish` verifies assertion and stores transaction fields including `tx_id`, `bundle_cbor`, AD, CDJ, signature.

## Flows
- Transaction signing: blueprint/_user-flows/transaction-signing.md

## Interfaces
- POST /tx/signing/options
- POST /tx/signing/finish

## Risks
- CBOR determinism; nonce monotonicity race conditions; large payload limits; replay attempts.

Refs: goal transaction-content-signing; goal server-derived-challenge-and-txid; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-FLOW-SIGN
