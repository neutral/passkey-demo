# Purpose
Exercise the login options handler to ensure it returns SimpleWebAuthn-compatible JSON, persists session state with a 5 minute TTL, retries session-id collisions, and maps generator failures to JSON error envelopes.

# Key Logic
- Spins up the real Express router with deterministic dependencies for challenge/ID generation, verifying `login_session_id`, `expires_at`, and stored session contents.
- Covers TTL pruning semantics, collision retries, and the error path when `generateAuthenticationOptions` throws.
- Uses `AbortSignal.timeout` and closes servers after each run to avoid hanging sockets.

# Interactions
Targets `src/webauthn/login.js` directly; mirrors the registration options suite so Step 8 can reuse the exported store during finish verification.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-SEC-UV; spec R-FLOW-LOGIN/spec.md; decision webauthn-corrections-and-standardizations; decision http-error-envelope
