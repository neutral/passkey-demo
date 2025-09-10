# Overview
Holds deterministic vector generation used to produce and verify golden reference data for the transaction bundle and anchors.

# Relations
- Used by `server/cmd/vectors` CLI to emit goldens.
- Imported in tests to recompute vectors and compare against committed JSON under `specs/goldens/`.
- Depends on `internal/encoding` (canonical CBOR, base64url) and `internal/types` (COSE, Bundle).

# Interfaces & Models
- `Generate() (Vectors, error)`: returns a structured JSON-ready payload containing inputs, bundle encodings, and anchors.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

