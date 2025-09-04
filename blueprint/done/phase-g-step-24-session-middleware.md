### Step 24 — Session middleware (Done: 2025-09-04)

Verification notes
- Unit tests executed: `cd server && GOCACHE=$(pwd)/.gocache go test ./internal/http -v` and `go test ./...` → all green.
- Tx handlers (options, finish, list) continue to pass with middleware present; cookie fallback retained for compatibility.

24. **Session middleware**

    Scope
    - Add reusable middleware that authenticates requests via cookie `sid`, looks up the server session in DB, enforces expiry, and attaches account context to the request (`acct_cbor` and credential ids). Optionally refreshes expiry on activity.
    - Refactor existing tx handlers (options, finish, list) to read account context from middleware instead of in‑handler cookie lookups.

    Source to add/modify
    - Add `server/internal/http/session.go` (or `server/internal/middleware/session.go`):
      - `type SessionContext struct { AcctCBOR []byte; CredentialIDs [][]byte }`
      - Context helpers: `WithSession(ctx context.Context, s SessionContext) context.Context`, `FromSession(ctx context.Context) (SessionContext, bool)`.
      - Middleware: `func SessionMiddleware(db *sql.DB, refresh bool, ttl time.Duration) func(http.Handler) http.Handler` — reads cookie `sid`; loads `acct_cbor, expires_at`; when `refresh` true and near expiry, updates `expires_at = now+ttl` in DB; queries credential ids `SELECT credential_id FROM credentials WHERE acct_cbor_fk=?`; attaches to context.
    - Modify handlers to consume context:
      - Update `server/internal/tx/options.go` (handler): remove cookie+session DB lookup; retrieve `acct_cbor` and use it; keep body parsing and call to `BuildTxOptions` with `acctCBOR` from context. Update tests if they stood up handler directly; include middleware in test wiring for integration style checks.
      - Update `server/internal/tx/finish.go` (handler): remove cookie+session DB lookup and rely on context; still validate tx session from store.
      - Update `server/internal/tx/list.go` (handler): remove cookie+session DB lookup and rely on context.
    - Tests:
      - Add `server/internal/http/session_test.go` to validate middleware behavior (see Tests below).
      - Update tx handler tests to use the middleware in `httptest` pipelines for end‑to‑end checks where necessary.

    Description files (to create AND updates for modified sources)
    - Create `server/internal/http/session.go.desc.md`: overview, relations (wraps routes; provides context to tx handlers), invariants, and Refs.
    - Update `server/internal/tx/*.desc.md` (options.go.desc.md, finish.go.desc.md, list.go.desc.md): note dependency on session middleware for `acct_cbor` and credential ids.

    Request/response shape
    - N/A (infrastructure). Inputs: cookie `sid`. Effect: enriches request context; unchanged response envelopes.

    Algorithm
    - Read cookie `sid`; if missing → call `next` with a context lacking session; downstream protected handlers should return 401.
    - DB lookup: `SELECT acct_cbor, expires_at FROM sessions WHERE session_id=?`.
      - If not found or expired (`expires_at <= now`) → do not set context; downstream returns 401.
    - Query account credential ids: `SELECT credential_id FROM credentials WHERE acct_cbor_fk=?` (optional but useful for allowlists).
    - Attach `SessionContext{AcctCBOR, CredentialIDs}` to `r.Context()` and forward to next.
    - Optional refresh: if `refresh==true` and `expires_at - now < ttl/2`, update `expires_at = now + ttl` (DB `UPDATE`). Cookie value remains the same (opaque id); cookie attributes unchanged for this demo.

    Database interactions
    - Read: `sessions` by `session_id`, `credentials` by `acct_cbor_fk`.
    - Optional write: `UPDATE sessions SET expires_at=? WHERE session_id=?` when refresh is enabled.

    Policies & limits
    - TTL: reuse login TTL (e.g., 1 hour). Refresh on activity optional; default enabled during dev to smooth flows.
    - Privacy: do not log raw `acct_cbor`; if logging, hash identifiers with SHA‑256.
    - Scope: only attaches data; authorization left to handlers.

    Sequencing
    - Wire middleware only for `/tx/*` paths for now; keep `/authn/*` flows unchanged.
    - Refactor tx handlers to require session context and return 401 when absent; remove duplicate cookie lookups.
    - Step 25/38 will layer rate limits and standardized error envelopes.

    Tests (happy path required, negative cases, invariants)
    - File: `server/internal/http/session_test.go`
    - Setup: in‑memory SQLite; insert `accounts`, `credentials`, and `sessions` rows.
    - Happy path: request with valid `sid` → wrapped handler observes `FromSession(ctx)` true; `AcctCBOR` matches DB; `CredentialIDs` contains expected ids.
    - Negative: missing cookie → downstream protected handler returns 401.
    - Negative: unknown/expired session → downstream returns 401.
    - Refresh (optional): with `refresh=true` and near expiry, after request, DB `expires_at` advanced by roughly `ttl` (allow delta ≤ few seconds).
    - Handler integration: wrap `TxOptionsHandler`, `TxFinishHandler`, and `TxListHandler` with middleware and assert 401/200 behaviors without duplicating session logic.
    - Commands:
      - `cd server && go test ./internal/http -v`
      - `cd server && go test ./internal/tx -v && go test ./...`

    Verification
    - Unit: middleware tests pass; updated tx handler tests still pass with middleware in place.
    - Manual: start server (once routes are wired), hit `/tx/list` with and without cookie and observe 401 vs 200.
    - User verification commands
      ```bash
      cd server
      go test ./internal/http -v
      go test ./internal/tx -v
      go test ./...              # repo-wide sanity
      ```

    Acceptance criteria
    - Middleware attaches `acct_cbor` + `credential_ids` to context for valid sessions; 401 behavior enforced by tx handlers when absent/expired.
    - Duplicate cookie/session lookups removed from tx handlers; tests updated accordingly.

    Notes
    - Keep middleware minimal and composable; avoid adding CORS/limits here (handled in later steps).
    - Do not change cookie format or attributes in this step.

    Refs
    - Refs: goal passkey-registration-login-uv; goal transaction-content-signing; requirement R-PLAT-2; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations

