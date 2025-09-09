# Purpose
Provide thin wrappers for CBOR encode/decode using `cbor-x` with deterministic behavior suitable for canonical bundle encoding.

# Key Logic
- `encodeCanonical(value)`: encodes JS values to CBOR bytes. Use JS `Map` with numeric keys for integer-key maps (COSE and Bundle).
- `decodeCBOR(data)`: decodes CBOR bytes back to JS values when needed.

# Interactions
- Used by `web/src/lib/bundle.ts` to encode the Bundle (B) deterministically (canonical) for server hashing and equality checks.

# Refs
Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails; spec bundle-shape-and-client-production-explainer

