# Purpose
Covers the `/tx/signing/options` handler: validates happy path parity with golden vectors and exercises all major error branches (auth required, malformed bundle, nonce conflict, missing credentials, session id collision, TTL pruning).

# Key Logic
- Spins up an Express app with in-memory SQLite, seeding accounts/credentials using the shared golden vector, and injects deterministic `TxSessionStore`, `now`, and `idFactory` dependencies.
- Verifies positive responses include the derived challenge/tx_id, WebAuthn options, and persisted store payload; asserts store buffers match canonical bytes from the golden bundle.
- Drives negative scenarios: missing session (401), invalid JSON/base64 (400), nonce conflict (409 via preseeded transaction), missing credentials (409), id collision (500), and TTL pruning removing expired entries.

# Interactions
Uses `createTxOptionsRoutes`, the new `TxSessionStore`, bundle helper (via handler), and SQLite migrations. Tests rely on Node’s native fetch for HTTP calls.

# Refs
Refs: goal server-derived-challenge-and-txid; requirement R-FLOW-SIGN; requirement R-SEC-UV; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations
