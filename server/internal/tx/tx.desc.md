# Overview
Helpers for transaction signing workflows. Validates client‑provided Bundle CBOR, enforces account/key binding and nonce monotonicity, and derives canonical anchors used by signing: challenge and tx_id. Also issues signing options and maintains a short‑lived in‑memory tx session.

# Relations
- Consumed by `/tx/signing/options` (Step 21) to validate input and create a tx session; session is later used by finish handler (Step 22).
- Reads from SQLite `transactions` to enforce per‑account nonce monotonicity.
- Depends on `internal/encoding` for base64url + canonical CBOR and `internal/types.Bundle` schema.

# Interfaces & Models
- `ValidateAndAnchorBundle(ctx, db, acctCBOR, bundleCBORBase64)` → `AnchoredBundle` (parsed bundle, canonical CBOR `B`, challenge `[32]byte`, tx_id `[32]byte`).
- `TxOptionsHandler(cfg, txStore, db)` → HTTP handler for POST `/tx/signing/options`; `BuildTxOptions(...)` generates response and stores session.
- Sentinel errors: `ErrBundleBase64`, `ErrBundleCBOR`, `ErrSenderKeyMismatch`, `ErrNonceNotMonotonic`.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations
