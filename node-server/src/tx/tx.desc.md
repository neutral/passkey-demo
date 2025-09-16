# Overview
Transaction helper modules for the Node server. Currently hosts bundle validation/anchoring logic shared by upcoming signing routes; subsequent steps will add options/finish HTTP handlers alongside this helper.

# Relations
Consumed by the signing HTTP handlers (options and finish) to canonicalize CBOR bundles, derive anchors, and enforce nonce/account policies before DB writes. Depends on SQLite access for nonce lookups and on `cbor-x` for canonical encoding/decoding.

# Interfaces & Models
Exports helper functions (starting with `validateAndAnchorBundle`) that operate on canonical CBOR bundles and return derived anchors plus parsed logical fields (`senderKey`, `nonce`, `message`, `validUntil`). Error paths surface typed `BundleValidationError` instances so routers can map to HTTP envelopes.

# Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails
