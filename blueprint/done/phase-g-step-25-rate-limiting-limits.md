### Step 25 — Rate limiting & limits (Done: 2025-09-04)

25. **Rate limiting & limits**

    Scope
    - Add lightweight, in-memory protections:
      - Per-IP token bucket rate limiting applied to `/authn/*` and `/tx/*`.
      - Request body size limit (e.g., 64 KB) for JSON endpoints.
    - Keep defaults conservative for a demo; avoid external dependencies.

    Source to add/modify
    - Add `server/internal/http/rate.go`:
      - `type TokenBucket struct { capacity int; tokens float64; rate float64; last time.Time }`
      - `type RateLimiter struct { mu sync.Mutex; buckets map[string]*TokenBucket; capacity int; rate float64 }`
      - `func NewRateLimiter(capacity int, rate float64) *RateLimiter`
      - `func (rl *RateLimiter) Allow(ip string, now time.Time) bool` — refill + check.
      - `func RateLimitMiddleware(rl *RateLimiter, ipfn func(*http.Request) string) func(http.Handler) http.Handler`
      - `func RemoteIP(*http.Request) string` — derive IP from `RemoteAddr` (no XFF trust by default).
    - Add `server/internal/http/bodylimit.go`:
      - `func BodyLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler` — wraps `http.MaxBytesReader`, returns 413 on overflow.
    - Tests: `server/internal/http/rate_test.go`, `server/internal/http/bodylimit_test.go` (see Tests).
    - Optional wiring: note in comments how to wrap `/authn/*` and `/tx/*` routes; final router wiring handled when routes are centralized.

    Description files (to create AND updates for modified sources)
    - Create `server/internal/http/rate.go.desc.md`: overview of token bucket, IP extraction assumptions, invariants; Refs.
    - Create `server/internal/http/bodylimit.go.desc.md`: purpose, behavior, returned status; Refs.

    Request/response shape
    - N/A for inputs; middleware concerns only. Responses:
      - 429 Too Many Requests when rate limit exceeded.
      - 413 Payload Too Large when body exceeds limit.

    Algorithm
    - Token bucket:
      - Each IP has a bucket with capacity `C` and refill rate `R` tokens/second.
      - On each request: `tokens = min(C, tokens + (now-last)*R)`; if `tokens >= 1`, allow and decrement; else 429.
      - Use a mutex and a map keyed by IP; lazy-create buckets.
    - Body limit:
      - For requests with bodies, wrap `r.Body` using `http.MaxBytesReader(w, r.Body, max)`; if read exceeds, return 413 and stop processing.

    Database interactions
    - None.

    Policies & limits
    - Defaults (tunable):
      - `/authn/*`: `capacity=5`, `rate=0.5` tokens/sec (~30/min), body limit 64 KB.
      - `/tx/*`: `capacity=10`, `rate=1` token/sec (~60/min), body limit 64 KB.
    - IP extraction: use `RemoteAddr` (host part); do not trust headers in dev; document proxy setups if needed.

    Sequencing
    - Apply non-auth middlewares to both `/authn/*` and `/tx/*` independently of session middleware.
    - Integrate in the router when centralization arrives; meanwhile, tests wrap handlers directly via middleware chains.

    Tests (happy path required, negative cases, invariants)
    - Files: `server/internal/http/rate_test.go`, `server/internal/http/bodylimit_test.go`
    - Rate limiter:
      - Allow under threshold: make `capacity` requests quickly → all 200.
      - Exceed threshold: one more request → 429; assert subsequent 429 until tokens refill.
      - Refill: sleep to allow partial tokens, assert acceptance after `~1/R`.
      - Per-IP isolation: `IP A` limited does not affect `IP B`.
    - Body limit:
      - Request body within 64 KB → 200; oversized by 1 byte → 413.
      - Ensure handler body read completes without panic when limit exceeded.
    - Commands:
      - `cd server && go test ./internal/http -run TestRate -v`
      - `cd server && go test ./internal/http -run TestBodyLimit -v && go test ./...`

    Verification
    - Unit: rate and body limit tests pass; bursts beyond capacity produce 429; oversized bodies return 413.
    - Manual: wrap any handler with the middlewares and observe expected behaviors.
    - User verification commands
      ```bash
      cd server
      go test ./internal/http -v
      go test ./...              # repo-wide sanity
      ```

    Acceptance criteria
    - Rate limiting returns 429 when bucket empty; tokens refill over time.
    - Body limit returns 413 for payloads > 64 KB; normal requests unaffected.
    - No shared state across IP buckets; performance acceptable for demo.

    Notes
    - Keep logic simple and documented; no third-party deps.
    - Future: header-based IP detection behind trusted proxies, per-route configs, and error envelopes (Step 38).

    Refs
    - Refs: requirement R-ERR; requirement R-PLAT-2; goal transaction-content-signing; decision webauthn-corrections-and-standardizations

