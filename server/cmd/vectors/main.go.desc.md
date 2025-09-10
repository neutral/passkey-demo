# Purpose
Implements the `vectors` CLI entrypoint to emit golden vectors for the canonical CBOR bundle and anchors.

# Key Logic
- Parses `-fmt json|text` to choose output format.
- Invokes `internal/vectors.Generate()` and prints either JSON (indented) or a text summary.
- Exits non‑zero on generation or encoding errors.

# Interactions
- Depends on `internal/vectors` for computation. No external I/O beyond stdout/stderr.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

