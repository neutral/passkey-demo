# Overview
Holds WebAuthn-specific helpers for the Node server, including session stores and Express routers that wrap `@simplewebauthn/server` flows. Keeps ceremony policy defaults (UV required, resident keys, attestation none) centralized and shares COSE canonicalization/hash helpers across registration and login flows.

# Relations
Routes are mounted by `src/server.js` under `/authn/passkey/*`. Session stores created here are shared with finish handlers to validate client responses and enforce TTLs. Depends on shared utilities such as `logger.js`, `error.js`, and SQLite access for persistence/logging (registration finish); login options remains in-memory.

# Interfaces & Models
Exposes `RegistrationSessionStore`/`createRegistrationRoutes(config, deps)` for registration and `LoginSessionStore`/`createLoginRoutes(config, deps)` for login. Stores sessions with `{ challenge, rpID, origin, expiresAt }`; registration routes additionally verify responses, canonicalize COSE keys, hash accounts (`SHA-256("ACCTK1"||acct_cbor)`), and persist credentials atomically.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-REG/spec.md; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope
