# Overview
Transaction helper modules for the Node server. Hosts both the bundle validation/anchoring helper and the `/tx/signing/options` router that issues WebAuthn options + tx sessions; future files (finish/list) will live alongside them.

# Relations
Bundle helpers and the options router are consumed by signing handlers to canonicalize CBOR bundles, derive anchors, enforce nonce/account policies, and stage short-lived tx sessions. Depends on SQLite for account credential lookups and on shared logging/error utilities for envelope mapping.

# Interfaces & Models
- `validateAndAnchorBundle(db, acctCbor, bundleB64)` → `{ bundle, canonical, challenge, txId }` with typed `BundleValidationError` kinds.
- `createTxOptionsRoutes(config, deps)` → Express router + `TxSessionStore` (in-memory TTL map storing `{ canonical B, challenge, txId, credentialIds, acctCbor, expiresAt }`). Responses provide `{ tx_session_id, challenge, options, tx_id_hex, expires_at }` for the web client.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SEC-UV; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails
