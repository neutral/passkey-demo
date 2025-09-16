# Overview
Transaction helper modules for the Node server. Hosts bundle validation/anchoring logic, the signing options handler, the signing finish handler that persists transactions, and the authenticated list endpoint that reads back stored transactions.

# Relations
- Options handler validates bundles and seeds `TxSessionStore` entries consumed by the finish handler.
- Finish handler verifies WebAuthn assertions using the stored session data, updates credential counters, and writes to the `transactions` table.
- List handler reads `transactions` via SQLite for the logged-in account and serializes `{ tx_id_hex, nonce, message, created_at }` for the dashboard with `ERR_UNAUTHORIZED`/`ERR_INTERNAL` envelopes on session/DB failures, logging success/error via shared helpers.
- Depends on SQLite for credential lookups/inserts and shared logging/error utilities for envelope mapping.

# Interfaces & Models
- `validateAndAnchorBundle(db, acctCbor, bundleB64)` → `{ bundle, canonical, challenge, txId }` with typed `BundleValidationError` kinds.
- `createTxOptionsRoutes(config, deps)` → Express router + `TxSessionStore` (stores `{ canonical, challenge, txId, credentialIds, acctCbor, expiresAt }`).
- `createTxFinishRoutes(config, deps)` → Express router consuming the same store, verifying assertions, persisting transactions, and emitting `tx_finish`/`webauthn_assert_verify` logs.
- `createTxListRoutes(config, deps)` → Express router requiring authenticated sessions and returning stored transactions sorted desc along with `tx_list`/`tx_list_error` logs.
- `createTxListRoutes(config, deps)` → Express router requiring authenticated sessions and returning stored transactions sorted desc.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SEC-UV; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails
