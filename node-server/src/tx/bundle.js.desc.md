# Purpose
Implements `validateAndAnchorBundle`, the Node analogue to the Go server helper. It decodes client-supplied transaction bundles (base64url CBOR), enforces schema/policy invariants, and derives canonical bytes plus challenge/tx_id anchors.

# Key Logic
- Uses `cbor-x` decoders to parse the bundle and recover nested COSE sender keys, including fallback paths when the nested map arrives in alternate shapes (string keys, tagged bytes).
- Enforces bundle constraints: message UTF-8 length ≤ 1024, nonce ∈ [1, 2^53−1] and strictly increasing per account (queried via `SELECT MAX(nonce)`), and sender key equality with the authenticated account COSE public key.
- Re-encodes the logical bundle into canonical CBOR (`B`) and derives SHA-256 anchors `challenge = SHA256("CHALv1"||B)` and `tx_id = SHA256("TXIDv1"||B)`.
- Surfaces failures as `BundleValidationError` with stable `kind` codes (base64, CBOR, sender key, nonce, message) for HTTP mapping.

# Interactions
Reads from SQLite to enforce nonce monotonicity; no writes performed. Returns Buffers for anchors/canonical bytes that downstream signing routes persist or include in WebAuthn options. Relies on `better-sqlite3` for prepared statements cached per-connection.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails
