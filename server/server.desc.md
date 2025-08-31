# Overview
Backend service written in Go that implements WebAuthn registration/login and transaction signing. It exposes REST endpoints, verifies ceremonies (UV required), persists accounts/credentials/transactions in SQLite, and serves as the single backend for the demo.

# Relations
- Used by the web SPA via JSON over HTTP; relies on platform authenticators (WebAuthn) in modern browsers.
- Depends on SQLite for persistence; defines request/response types and encoding helpers (base64url, canonical CBOR) shared across handlers.

# Interfaces & Models
- Endpoints: `/authn/passkey/registration/options|finish`, `/authn/passkey/login/options|finish`, `/tx/signing/options|finish`, `/tx/list`.
- Models: account (COSE EC2 public key in canonical CBOR as identity), credential (maps to account; tracks signCount, optional AAGUID), session, transaction (tx_id, bundle CBOR, AD, CDJ, signature).
- Policies: UV required; canonical CBOR for bundle; low-S signatures; monotonic signCount and per-account nonce.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-PLAT-3; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
