# Purpose
Generates deterministic golden vectors for the transaction bundle (B), sender key CBOR, and derived anchors used by tests and tools.

# Key Logic
- Builds a COSE EC2 key from the P-256 base point with 32-byte left-padded X/Y to ensure stable CBOR bytes.
- Constructs a minimal bundle `{sender_key, nonce=1, message="hello"}` and encodes it using canonical CBOR.
- Computes anchors using fixed prefixes:
  - `challenge = SHA-256("CHALv1" || B)`
  - `tx_id     = SHA-256("TXIDv1" || B)`
- Emits JSON fields with both hex and base64url encodings where specified.

# Interactions
- Depends on `internal/encoding` for canonical CBOR and base64url.
- Consumed by the `vectors` CLI and by unit tests that compare against committed goldens.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

