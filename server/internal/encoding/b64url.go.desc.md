# Purpose
Provide consistent base64url encoding/decoding for binary fields transported via JSON and occasionally URLs. Encodes without padding; decodes tolerantly with/without padding.

# Key Logic
- Encode: `base64.RawURLEncoding` (no `=` padding).
- Decode (tolerant): try RawURLEncoding; then URLEncoding; then pad to multiple of 4 and retry; error on invalid alphabet/length.

# Interactions
- Used by handlers and parsers to convert credential IDs, keys, signatures, and CBOR bundles between bytes and strings at API boundaries.

# Refs
Refs: decision encoding-and-ceremony-guardrails; requirement R-PLAT-2; requirement R-ERR
