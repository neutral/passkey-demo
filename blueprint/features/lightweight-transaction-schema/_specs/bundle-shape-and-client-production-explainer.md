# Bundle Shape & Client Production Explainer

## Purpose
Clarify the Bundle schema, its canonical CBOR encoding, and how the web client constructs and submits it to the server.

## Overview
- Bundle (CDDL):
  - `0: CoseKeyEC2Pub` (sender_key)
  - `1: uint` (nonce)
  - `2: tstr` (message)
  - `3: uint` (valid_until, optional)
- Go model: `types.Bundle` mirrors this schema with CBOR integer map keys.
- Client responsibilities:
  - Build the logical Bundle from UI inputs and the account’s COSE key.
  - Encode to canonical CBOR bytes `B` (deterministic).
  - Base64url-encode `B` (without padding) and send as `bundle_cbor_b64`.
- Server responsibilities:
  - Decode, validate, re-encode canonically to recompute `B`.
  - Derive anchors (`challenge`, `tx_id`) and enforce nonce/account policies.

## Example (conceptual)
```json
{
  "sender_key": { "kty": 2, "alg": -7, "crv": 1, "x": "<32B>", "y": "<32B>" },
  "nonce": 42,
  "message": "Pay Alice",
  "valid_until": 1735689600
}
```
Wire format: CBOR(Map[int->...]) encoded canonically → bytes `B` → `bundle_cbor_b64` in JSON request.

## Refs
Refs: goal minimal-cbor-bundle; goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

