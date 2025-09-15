# Purpose
Validate the authenticated `/me/account_key` handler returns the canonical account key payload, enforces session context, and maps missing authentication/account rows to the shared JSON error envelope.

# Key Logic
- Spins up the Express router with in-memory SQLite, seeding an account row and injecting a fake `req.session` to mimic the cookie middleware.
- Asserts the response includes canonical base64url CBOR, derived thumb hex, hex/base64url coordinates, and preserves the stored `created_at` timestamp.
- Covers unauthorized access (no session) and the edge case where a session exists but the account row is absent.

# Interactions
Targets `src/me.js` directly; uses the real DB schema via `applyMigrations` and mirrors other WebAuthn tests that rely on deterministic CBOR helpers.

# Refs
Refs: goal passkey-registration-login-uv; requirement R-UI-2BTN; spec passkey-first-identity/spec.md; decision http-error-envelope
