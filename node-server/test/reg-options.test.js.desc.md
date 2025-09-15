# Purpose
Exercise the WebAuthn registration options handler to ensure JSON policy flags, session storage, error envelopes, and collision handling behave as required.

# Key Logic
- Spins up an Express app with deterministic dependencies to assert the handler returns UV/resident-key enforced options and persists `{ challenge, rpID, origin, expiresAt }`.
- Verifies `RegistrationSessionStore.pruneExpired` drops stale entries, option generation failures map to `internal_error`, session-id collisions are retried before failing after three attempts, and HTTP requests use `AbortSignal.timeout` + per-test timeouts to avoid hanging runs.

# Interactions
Covers `src/webauthn/reg.js` through HTTP calls using Node's test runner and global `fetch`; relies on `RegistrationSessionStore` exports for state inspection.

# Refs
Refs: requirement R-FLOW-REG; spec R-FLOW-REG/spec.md; decision webauthn-corrections-and-standardizations
