# Purpose
Exercise the login finish handler covering success, session expiry, credential lookup failures, UV enforcement, sign-count conflicts, and malformed payloads. Ensures cookies are issued with correct attributes and login sessions are single-use.

# Key Logic
- Spins up the Express router with deterministic dependencies (clocks, session ID generator, SimpleWebAuthn verifier stubs) and in-memory SQLite, asserting credential updates, session inserts, and logging side effects.
- Validates error envelopes across 401/403/409/400 paths and confirms login sessions are removed on terminal failures.

# Interactions
Targets `src/webauthn/login.js` finish logic along with `sessions`/`credentials` tables. Complements `login-options.test.js` by covering finish-specific behaviors.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope; decision request-id-and-slog-json
