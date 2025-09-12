# Purpose
Per-IP token bucket rate limiter and middleware to protect `/authn/*` and `/tx/*` endpoints from abuse in a demo setting.

# Key Logic
- Lazy-create a bucket per IP with capacity `C` and refill rate `R` tokens/sec.
- On each request, refill by elapsed*R, cap to `C`, allow if `tokens >= 1`, else 429.
- Middleware extracts IP via `RemoteIP(r)` and enforces bucket.
- When blocked, returns 429 with a standardized JSON error envelope `{code:"ERR_RATE_LIMIT",error}`.

# Interactions
- Composable middleware layer for handlers and future centralized router.

# Refs
Refs: requirement R-PLAT-2; requirement R-ERR; goal transaction-content-signing; decision webauthn-corrections-and-standardizations
