# Purpose
Verify WebAuthn assertion signatures (ES256) over `SHA256(authenticatorData || SHA256(clientDataJSON))` with strict DER parsing (no trailing bytes) and low‑S enforcement. Also provides a variant that accepts high‑S by normalizing `S`.

# Key Logic
- Digest construction as per WebAuthn: `hCDJ = SHA256(cdj)` then `digest = SHA256(ad || hCDJ)`.
- Parse ASN.1 DER signature into (r,s); enforce curve P‑256 and low‑S (s ≤ N/2) before calling `ecdsa.Verify`.
- Exported sentinel errors for observability: `ErrUnsupportedCurve`, `ErrMalformedDER`, `ErrHighS`, `ErrBadSignature`.
- Variant: `VerifyAssertionAllowHighS` normalizes `s` via `s' = N − s` when high‑S and verifies with `(r, s')`. Use for login compatibility; keep low‑S strictness for transaction signing.

# Interactions
- Used in login/sign finish handlers after parsing AD/CDJ and retrieving the account public key.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; decision webauthn-accept-high-s-login-only
