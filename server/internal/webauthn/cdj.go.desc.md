# Purpose
Parse WebAuthn ClientDataJSON to extract type, challenge, and origin; decode the base64url challenge with tolerant rules.

# Key Logic
- JSON unmarshal into a raw struct; require `type`, `origin`, `challenge`.
- Accept only `webauthn.get` or `webauthn.create`.
- Decode `challenge` using internal base64url utilities (unpadded encode policy; tolerant decode).

# Interactions
- Used by assertion/attestation verification alongside authenticatorData parsing and signature checks.
- Policy checks for origin/RP binding are performed elsewhere (security step), not here.

# Refs
Refs: requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
