# Purpose
Serve WebAuthn login flows: `/authn/passkey/login/options` issues SimpleWebAuthn authentication options with demo policies and stores a short-lived login session; `/authn/passkey/login/finish` verifies the browser response, updates credential counters, creates DB-backed sessions, and emits structured logs mirroring the Go backend.

# Key Logic
- `LoginSessionStore`: in-memory TTL store (5 min) with collision detection and `pruneExpired(now)` to bound size.
- `createLoginRoutes(config, deps)`: builds an Express router that generates options via `generateAuthenticationOptions`, enforces `userVerification: 'required'`, returns the library-native JSON plus `{ login_session_id, expires_at }`, logs `login_options`, and maps generator failures to `ERR_INTERNAL`.
- Finish handler verifies assertions with `verifyAuthenticationResponse`, enforces UV and challenge binding, updates credential `sign_count`, inserts a persistent session row, sets the `sid` cookie (HttpOnly, SameSite=Lax, Secure when origin is https), deletes the login session, and logs `login_finish` with hashed identifiers while returning `ERR_BAD_REQUEST`/`ERR_UNAUTHORIZED`/`ERR_FORBIDDEN`/`ERR_CONFLICT` envelopes for policy violations.

# Interactions
Mounted by `src/server.js` at `/authn/passkey/login` after shared rate limiting and 1 MiB JSON body cap. Depends on `logger.js` for structured logs, `error.js` for JSON envelopes, and SQLite for credential lookups/session persistence. The exported store is reused across options/finish handlers.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope; decision request-id-and-slog-json
