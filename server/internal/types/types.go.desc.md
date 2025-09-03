# Purpose
Shared data models used across registration, login, and transaction signing. Keeps COSE key fields, the signing Bundle structure, and minimal WebAuthn payload shapes in one place to avoid duplication.

# Key Types
- `CoseEC2`: COSE EC2 public key fields (kty, alg, crv, x, y) with CBOR map index tags.
- `Bundle`: canonical CBOR map for signing content — sender key, nonce, message, optional valid_until.
- `RegOptions` / `RegFinish`: minimal registration request/response JSON shapes (binary fields as base64url strings).
- `LoginOptions` / `LoginFinish`: minimal assertion request/response JSON shapes (binary fields as base64url strings). `LoginOptions` includes an optional `allow_credentials` list for discoverable vs filtered flows.

# Interactions
- Used by handlers to marshal/unmarshal JSON and by helpers to encode/decode canonical CBOR (with Step 10 codec).
- `CoseEC2` feeds the crypto conversion helper (Step 8) to obtain an `ecdsa.PublicKey` for verification.

# Refs
Refs: requirement R-ID-KEY; requirement R-SCHEMA-LITE; requirement R-PLAT-2; requirement R-FLOW-LOGIN; goal key-first-identity-cose; goal minimal-cbor-bundle
