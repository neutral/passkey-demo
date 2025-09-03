# Bundle Anchors (CHALv1/TXIDv1) Explainer

## Purpose
Explain how the server derives deterministic anchors from the transaction Bundle: the WebAuthn `challenge` and the transaction identity `tx_id`, both bound to the bundle’s canonical CBOR bytes `B`.

## Overview
- The client submits a Bundle as canonical CBOR bytes `B` (base64url in JSON).
- The server re-validates and canonicalizes the bundle (ensures `B` determinism), then derives:
  - `challenge = SHA-256("CHALv1" || B)` — used in WebAuthn assertion options.
  - `tx_id     = SHA-256("TXIDv1" || B)` — stable identifier for storage and lookups.
- Prefixes provide versioned domain separation so different anchors never collide.

## Algorithms
```text
B        = cbor_canonical(Bundle)
challenge = SHA-256("CHALv1" || B)
tx_id     = SHA-256("TXIDv1" || B)
```

## Notes
- Canonical CBOR is required to keep hashes and signatures stable across platforms.
- The same logical Bundle always yields the same `B`, challenge, and `tx_id`.
- Any change to fields (e.g., `nonce`, `message`) changes both anchors.

## Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations

