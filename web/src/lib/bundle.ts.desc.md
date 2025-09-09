# Purpose
Build and encode the transaction bundle on the client in a way that matches the server’s CDDL and canonical CBOR requirements.

# Key Logic
- `CoseEC2`: client shape for ES256 key (kty=2, alg=-7, crv=1, x/y 32-byte Uint8Array).
- `buildBundle(sender, nonce, message, validUntil?)`: uses a JS `Map<number,any>` so CBOR keys are integers: `0: sender_key`, `1: nonce`, `2: message`, `3?: valid_until`.
- `encodeBundleCanonical`: calls `encodeCanonical` (cbor-x) to produce `B` (Uint8Array).
- `bundleToB64Hex`: previews `B` as base64url and hex for debugging and Step 33 transport.

# Interactions
- `Dashboard.tsx` collects inputs, fetches `sender_key` via `/me/account_key`, builds the bundle, and previews base64url/hex.

# Refs
Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer

