# Purpose
Convert COSE EC2 public keys (ES256 on P‑256) from WebAuthn into Go `ecdsa.PublicKey` values for signature verification.

# Key Logic
- Validate COSE parameters: `kty=2 (EC2)`, `alg=-7 (ES256)`, `crv=1 (P‑256)`.
- Enforce 32‑byte X/Y coordinates; construct `ecdsa.PublicKey` on `elliptic.P256()`.
- Verify the point lies on the curve; return an error if not.

# Interactions
- Consumes `types.CoseEC2` (from attestation parsing) and returns an `ecdsa.PublicKey` used by the signature verifier.
- Called by Step 13 utilities that perform assertion signature checks.

# Refs
Refs: requirement R-ID-KEY; requirement R-PLAT-2; goal key-first-identity-cose
