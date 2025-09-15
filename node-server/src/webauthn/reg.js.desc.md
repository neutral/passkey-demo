# Purpose
Serve registration flows: `/authn/passkey/registration/options` issues WebAuthn creation options with policy defaults and stores a short-lived session; `/authn/passkey/registration/finish` verifies the authenticator response, persists account + credential rows, and emits structured logs.

# Key Logic
- `RegistrationSessionStore`: Map-backed TTL store with collision detection and `pruneExpired(now)` to keep in-memory state bounded.
- `createRegistrationRoutes(config, deps)`: builds an Express router that logs `reg_options`, generates options via `@simplewebauthn/server`, enforces resident-key/UV policies, stores `{ challenge, rpID, origin, expiresAt }`, and surfaces `reg_session_id` + `expires_at` in JSON responses while mapping failures through `reg_options_error`.
- Finish handler: calls `verifyRegistrationResponse`, enforces fmt `none` + UV, canonicalizes COSE keys via `cbor-x`, hashes accounts (`SHA-256("ACCTK1"||acct_cbor)`), writes `accounts`/`credentials` in a transaction, logs `reg_finish`, and maps duplicate credentials to 409 envelopes.

# Interactions
Mounted by `src/server.js` under `/authn/passkey/registration`. Depends on shared `logger.js` for structured logs, `error.js` for envelopes, and SQLite via injected `db` for persistence. Exports a shared session store reused across options/finish handlers. Login flow in step 7 reuses account hashing/logging helpers from this module.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-REG; requirement R-SEC-UV; spec R-FLOW-REG/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope; decision request-id-and-slog-json
