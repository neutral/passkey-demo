# Account Binding (Sender Key Equality) Explainer

## Purpose
Explain why and how the Bundle’s `sender_key` must match the logged-in account’s COSE key to prevent cross-account signing.

## Overview
- The account identity is the COSE EC2 public key (key-first identity).
- The Bundle includes `sender_key` to bind content to that identity.
- Server canonicalizes `sender_key` and compares its CBOR bytes to the account’s `acct_cbor` from session:
  - `EncodeCanonical(bundle.sender_key) == acct_cbor` → pass
  - Otherwise reject with a specific error (mapped to 400/401 at the handler).

## Notes
- Comparison is done on canonical CBOR bytes, not on JSON, to avoid encoding ambiguities.
- This prevents a user from submitting a bundle with a different public key to sign content under another account.

## Refs
Refs: goal key-first-identity-cose; goal transaction-content-signing; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails

