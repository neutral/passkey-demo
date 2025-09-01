# Purpose
Parse the fixed header of WebAuthn authenticatorData for assertions to extract the RP ID hash, flags, and counter.

# Key Logic
- Length check ≥ 37 bytes.
- Copy first 32 bytes as `RpIDHash`, read `flags` (bitfield), read `signCount` as big‑endian uint32.
- Expose helpers for flags: `HasUP`, `HasUV`.

# Interactions
- Used by login and signing verification to enforce UV and counter policies and to validate rpIdHash.
- Remainder (beyond 37 bytes) is returned for flows that need attested credential or extension data.

# Refs
Refs: requirement R-PLAT-2; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations
