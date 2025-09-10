# Overview
Committed golden vectors for deterministic bundle CBOR and anchor values. Used by unit tests and external tools to validate encoding and hashing logic.

# Relations
- Produced by `server/cmd/vectors` CLI (`-fmt json`).
- Consumed by `server/internal/tx/anchors_golden_test.go` for equality checks.

# Interfaces & Models
- `tx-bundle-v1.json` schema:
  - `version`: fixed string.
  - `inputs`: `message`, `nonce`, `sender_key_cbor_hex`, `sender_key_cbor_b64`.
  - `bundle`: `bundle_cbor_hex`, `bundle_cbor_b64`.
  - `anchors`: `challenge_hex`, `challenge_b64`, `tx_id_hex`.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

