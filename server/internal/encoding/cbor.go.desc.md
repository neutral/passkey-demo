# Purpose
Provide deterministic (canonical) CBOR encoding/decoding for data that must hash/sign the same way across runs and platforms.

# Key Logic
- Encoder: uses fxamacker/cbor canonical options (deterministic map key ordering, shortest lengths, no indefinite forms).
- Decoder: standard safe defaults; roundtrips canonical output back into Go types via `DecodeCanonical`.

# Interactions
- Used by transaction bundle hashing and challenge derivation to ensure stable bytes for signatures.
- Lives alongside base64url utilities in `internal/encoding` as shared transport helpers.

# Refs
Refs: requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; requirement R-PLAT-2
