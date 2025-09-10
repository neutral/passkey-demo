# Purpose
Verifies that programmatically generated vectors (bundle CBOR and anchors) match the committed golden JSON, ensuring stability across platforms.

# Key Logic
- Calls `internal/vectors.Generate()` to recompute the vectors.
- Reads `specs/goldens/tx-bundle-v1.json` from the repo and unmarshals into the same struct.
- Compares fields for inputs, bundle encodings, and anchors; fails on any mismatch.

# Interactions
- Depends on `internal/vectors` and file I/O. Uses `runtime.Caller` to resolve the golden file path relative to the repository root.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

