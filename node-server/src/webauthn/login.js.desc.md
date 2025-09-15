# Purpose
Serve `POST /authn/passkey/login/options` using SimpleWebAuthn helpers. Issues authentication options with demo policies, persists a short-lived login session (challenge, rpID, origin, expiry), and emits `login_options` logs for parity with the Go backend.

# Key Logic
- `LoginSessionStore`: in-memory TTL store (5 min) with collision detection and `pruneExpired(now)` to bound size.
- `createLoginRoutes(config, deps)`: builds an Express router that generates options via `generateAuthenticationOptions`, enforces `userVerification: 'required'`, returns the library-native JSON plus `{ login_session_id, expires_at }`, and logs `login_options` (or `login_options_error` on failure).
- Session IDs use 24-char base64url strings (≥128 bits entropy); retries collisions up to three times before erroring.

# Interactions
Mounted by `src/server.js` at `/authn/passkey/login`. Depends on `logger.js` for structured logs and `error.js` for JSON envelopes. The exported store is reused by the login finish handler to validate challenges and TTLs.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope
