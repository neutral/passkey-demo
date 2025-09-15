# Overview
Holds WebAuthn-specific helpers for the Node server, including session stores and Express routers that wrap `@simplewebauthn/server` flows. Keeps ceremony policy defaults (UV required, resident keys, attestation none) centralized for reuse across options/finish handlers.

# Relations
Routes are mounted by `src/server.js` under `/authn/passkey/*`. Session stores created here are shared with finish handlers to validate client responses and enforce TTLs. Depends on shared utilities such as `logger.js` and `error.js` for structured logs and envelopes.

# Interfaces & Models
Exposes `RegistrationSessionStore` (in-memory with 300s TTL pruning) and `createRegistrationRoutes(config, deps)` returning `{ router, store }` for registration flows. Stores sessions with `{ challenge, rpID, origin, expiresAt }` ahead of verification logic in subsequent steps.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-REG; spec R-FLOW-REG/spec.md; decision webauthn-corrections-and-standardizations
