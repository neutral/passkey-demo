### Step 10 — Node: Bundle anchors and validation helpers

Scope

- Implement `ValidateAndAnchorBundle(acctCBOR, bundleCBORBase64)` in Node to match Go behavior:
  - Decode base64url → CBOR → typed Bundle `{sender_key, nonce, message[, valid_until]}`.
  - Canonical re-encode via `cbor-x` → `B`.
  - Enforce account binding (COSE fields equality), message length ≤ 1024, nonce > last seen and ≤ 2^53-1.
  - Anchors: `challenge = SHA-256('CHALv1'||B)`, `tx_id = SHA-256('TXIDv1'||B)`.
  - Decoder robustness similar to ADR on CBOR/COSE interop.

Source to add

- `node-server/src/tx/bundle.js`: implementation and error codes.

Verification

- Golden vector parity (read `specs/goldens/tx-bundle-v1.json`); compare `B`, `challenge`, `tx_id` to Go output.

Acceptance criteria

- Anchors and policies exactly match Go; error kinds map to envelope codes.

Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails

---

