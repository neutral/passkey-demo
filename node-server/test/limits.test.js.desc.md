# Purpose
Cover the new limits middleware by ensuring oversize bodies return `ERR_PAYLOAD_TOO_LARGE` envelopes and repeated requests trigger `ERR_RATE_LIMIT` with `Retry-After` headers.

# Key Logic
- Wraps test Express apps with `buildBodyLimit` and `buildRateLimiter`, using a controllable clock for deterministic token bucket behaviour.
- Exercises success/failure paths and asserts remaining headers/JSON codes.

# Interactions
- Verifies `src/limits.js` + `src/error.js` integration independent of WebAuthn/tx routers.

# Refs
Refs: requirement R-ERR; decision router-builder-wiring; decision http-error-envelope
