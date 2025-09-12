# Purpose
Validate a client-provided Bundle (base64url CBOR), recompute canonical CBOR bytes `B`, enforce account binding and nonce monotonicity, and derive anchors used by signing: `challenge` and `tx_id`.

# Key Logic
- Decode `bundle_cbor_b64` (base64url tolerant of padding) → CBOR → `types.Bundle`.
- Re-encode the bundle with canonical CBOR to form `B`.
- Compute anchors:
  - `challenge = SHA-256("CHALv1" || B)`
  - `tx_id = SHA-256("TXIDv1" || B)`
- Account binding: compare logical COSE fields (`kty`, `alg`, `crv`, `x`, `y`) of `sender_key` against `acct_cbor` from the logged-in session (tolerates encoder differences while preserving identity binding).
- Nonce policy: `bundle.nonce` must be strictly greater than `MAX(nonce)` in `transactions` for the account.
- Limits & invariants:
  - `nonce` must be within [1, 2^53-1] to align with JS safe integer range; `ErrNonceOutOfRange` when exceeded.
  - `message` length ≤ 1024 bytes; reject longer with `ErrMessageTooLong`.
- Sentinel errors for mapping by the HTTP layer: `ErrBundleBase64`, `ErrBundleCBOR`, `ErrSenderKeyMismatch`, `ErrNonceNotMonotonic`.

# Interactions
- Called by `/tx/signing/options` (Step 21) after resolving the user session to `acct_cbor`.
- Reads from DB table `transactions` to check monotonic nonce; no writes.
- Consumes `internal/encoding` (base64url + CBOR) and `internal/types.Bundle`.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations
