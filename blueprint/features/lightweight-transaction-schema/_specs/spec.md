# R-SCHEMA-LITE — Schema Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Canonical CBOR schema for demo bundle; deterministic encoding rules and anchors for challenge/tx_id.

## Interfaces
- Used internally by client and server; transmitted as base64url in JSON where needed.

## Data / Models
- Bundle and CoseKey structures as per CDDL; binary fields as base64url in API.

## Algorithms
- Deterministic CBOR encoding; hash anchors:
```text
B = cbor_canonical(Bundle)
challenge = SHA-256("CHALv1" || B)
tx_id     = SHA-256("TXIDv1" || B)
```

## Security / Privacy
- Ensure content-binding via derived challenge; avoid mutable fields that break determinism.

## Errors / Observability
- Reject non-deterministic encodings; log size/nonce anomalies.

## Testing Strategy
- Golden vectors for canonical CBOR; stable hashes across runs.

## Open Questions
- If/when to use optional `valid_until`.

Refs: decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; goal minimal-cbor-bundle; requirement R-SCHEMA-LITE
