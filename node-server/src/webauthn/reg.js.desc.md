# Purpose
Serve `POST /authn/passkey/registration/options` by generating WebAuthn creation options with policy defaults and persisting a short-lived registration session for later verification.

# Key Logic
- `RegistrationSessionStore`: Map-backed TTL store with collision detection and `pruneExpired(now)` to keep in-memory state bounded.
- `createRegistrationRoutes(config, deps)`: builds an Express router that logs `reg_options`, generates options via `@simplewebauthn/server`, enforces resident-key/UV policies, stores `{ challenge, rpID, origin, expiresAt }`, and surfaces `reg_session_id` + `expires_at` in JSON responses while mapping failures to `reg_options_error` + JSON envelope.
- Guards session ID entropy (24-char base64url) with limited regeneration attempts and maps failures through the JSON error envelope.

# Interactions
Mounted by `src/server.js` under `/authn/passkey/registration`. Depends on shared `logger.js` for structured logs and `error.js` for envelope responses. Designed so Step 6 can reuse the exported store when verifying registration finishes.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-REG; requirement R-SEC-UV; spec R-FLOW-REG/spec.md; decision webauthn-corrections-and-standardizations; decision request-id-and-slog-json
