# Purpose
Expose the logged-in account’s COSE EC2 public key (sender_key) so the frontend can build a canonical bundle bound to the correct identity.

# Key Logic
- Authenticated via session: reads `acct_cbor` from context or cookie `sid`.
- Decodes `acct_cbor` (canonical CBOR) into COSE EC2 and returns JSON with base64url `x` and `y`.
- 401 when unauthorized; no writes performed.

# Interactions
- Used by Dashboard when building the bundle (Step 32) to fetch `sender_key` that must match the stored account key.

# Refs
Refs: requirement R-FLOW-SIGN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations

