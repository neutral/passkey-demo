# Purpose
Enforce WebAuthn relying party binding by checking the RP ID hash from authenticatorData and validating the browser origin (scheme/host/port) against configured policy.

# API
- `CheckRpIdHash(adRp [32]byte, rpID string) error` — Validates `rpID`, computes SHA-256, and compares to `adRp`.
- `CheckRpIdHashAllowed(adRp [32]byte, primary string, allow []string) error` — Accepts any match across primary + allowlist.
- `CheckOrigin(origin, expected string, allow []string, devLocalhostOK bool) error` — Parses and normalizes origin and matches against `{expected} ∪ allow` with scheme policy.

# Normalization & Policy
- RP ID: lowercase → strip trailing dot → IDNA ToASCII (punycode) → reject scheme/path/spaces/IP (except `localhost`).
- Origin: parse URL; host lowercase → strip trailing dot → IDNA ToASCII; default ports (80/443) when port omitted.
- Scheme: require `https`; allow `http` only for `localhost` when `devLocalhostOK` is true.
- Allowlists: exact matches only; no wildcards or eTLD+1.

# Errors
- RP: `ErrRpIdInvalid`, `ErrRpIdHashMismatch`.
- Origin: `ErrOriginMalformed`, `ErrOriginScheme`, `ErrOriginHost`, `ErrOriginPort`, `ErrOriginNotAllowed`.

# Refs
Refs: specs/webauthn-rp-origin-verification-spec.md; specs/rp-origin-test-vectors.md; specs/rp-origin-risks-and-ambiguities.md
