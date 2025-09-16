# Purpose
Expose reusable middleware for enforcing JSON body size caps and per-origin token-bucket rate limiting consistent with the Go backend.

# Key Logic
- `buildBodyLimit({bytes})` checks `Content-Length`, applies an `express.json` parser with the same limit, and calls `respondPayloadTooLarge` when exceeded (including `details.max_bytes`).
- `buildRateLimiter({burst, refillPerSecond, getKey, now})` implements an in-memory token bucket keyed by `req.ip` (overrideable) and emits `ERR_RATE_LIMIT` with `Retry-After` when tokens are exhausted.
- `DEFAULT_LIMITS` mirrors demo policies (1 MiB bodies for `/authn/*` and `/tx/*`, burst 20 / refill 10 rps).

# Interactions
- Applied in `server.js` around `/authn/*` and `/tx/*` routers before handlers so every WebAuthn/transaction request observes the same limits.
- Depends on `error.js` responders to emit standardized envelopes and headers.

# Refs
Refs: requirement R-ERR; decision http-error-envelope; decision router-builder-wiring
