# Purpose
Verify that a successful login creates a server session with a one-hour expiry recorded in the database and that a `sid` cookie is issued.

# Key Logic
- Build a valid WebAuthn get() flow (UV set, monotonic signCount) and call `LoginFinishHandler`.
- Parse `Set-Cookie` to extract `sid`; query `sessions.expires_at`.
- Assert the expiry delta is approximately 3600 seconds (with tolerance for execution time).

# Interactions
- Uses `internal/storage` to open/migrate SQLite, `internal/encoding` for base64url, and `internal/types` for COSE encoding.
- Exercises the complete login finish path that inserts into `sessions` and sets the cookie.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-2; decision router-builder-wiring; spec session-cookies-usage-explainer

