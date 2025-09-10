# Overview
CLI tool to generate deterministic golden vectors for bundle CBOR and anchors. Useful for review and for tests to compare against a committed JSON file.

# Relations
- Calls `internal/vectors.Generate()` to produce all values.
- Outputs either JSON (default) or a human-readable text block.

# Interfaces & Models
- Flags: `-fmt json|text` (default `json`).
- Writes to stdout.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

