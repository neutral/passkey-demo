# Purpose
Validate the registration finish handler by simulating success, policy failures, duplicate credentials, and malformed requests using in-memory SQLite and stubbed SimpleWebAuthn verifiers.

# Key Logic
- Spins up Express with the real router, passing deterministic clocks, session store entries, and custom `verifyRegistrationResponse` stubs to assert expected inputs and returned data.
- Verifies persistence (`accounts`, `credentials`), thumb hashing, structured responses, session consumption semantics, and error envelopes across 401/403/409 paths.
- Employs `AbortSignal.timeout` per request to avoid hanging tests and ensures database connections are closed after each run.

# Interactions
Covers `src/webauthn/reg.js` finish logic, `RegistrationSessionStore`, and SQLite writes via `applyMigrations`. Shared COSE helpers produce canonical keys consistent with production hashing behavior.

# Refs
Refs: requirement R-FLOW-REG; requirement R-SEC-UV; spec R-FLOW-REG/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope
