# Purpose
Exercises the bundle validation helper in isolation to guarantee parity with Go golden vectors and enforce policy/error coverage before integrating HTTP routes.

# Key Logic
- Loads the shared golden vector (`specs/goldens/tx-bundle-v1.json`) to assert canonical CBOR bytes, challenge, and tx_id hashes match historical outputs.
- Mutates canonical bundles to probe policy failures (sender key mismatch, nonce monotonicity, message length, nonce range) and ensures each failure surfaces the expected `BundleValidationError.kind`.
- Verifies fallback decoding tolerates alternative sender key shapes (object fields) so the helper remains resilient to client encoder differences.

# Interactions
Spins up in-memory SQLite via `better-sqlite3`, applying migrations for nonce lookups; no persistent IO. Depends on `cbor-x` to mutate CBOR maps for targeted scenarios.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails
