# Phase I — Step 37c — Integration Fixes and Standardizations

## Purpose
Capture the set of integration fixes and standardizations applied after the 37b refactoring plan to stabilize the system and remove drift across layers.

## Changes

- Central router builder and group middlewares
  - `server/internal/app/router.go` (+ `.desc.md`)
    - Single `BuildRouter(cfg, db, deps)` assembles all routes.
    - Layering: `RequestID` → `CORS` → `SessionMiddleware` → group middlewares (body/rate limits) → handlers.
    - `/authn/*` and `/tx/*` wrapped with 1 MiB body limit and token-bucket rate limiter.
  - ADR: `blueprint/_decisions/router-builder-wiring.md` (Accepted).
  - `server/cmd/api/main.go` delegates to `BuildRouter` and prepares injected deps.

- Session-only auth model in protected handlers
  - Removed handler-level cookie→DB fallbacks; rely exclusively on `SessionMiddleware` for auth context.
  - Updated handlers:
    - `server/internal/me/account_key.go`
    - `server/internal/tx/list.go`
    - `server/internal/tx/options.go`
    - `server/internal/tx/finish.go`
  - Descriptions updated for the above files.

- Error envelope and request IDs
  - `server/internal/httpx/errors/*` — standard JSON envelope `{code,error,correlation_id?}`; tests included.
  - `server/internal/httpx/middleware/request_id.go` — attaches request id and echoes via header.
  - Adopted in `/tx/signing/options` (returns clear codes like `ERR_BAD_REQUEST`, `ERR_CONFLICT`, etc.).
  - ADRs: `blueprint/_decisions/http-error-envelope.md`, `blueprint/_decisions/request-id-and-slog-json.md` (Accepted).

- Generic TTL store and shared randomness
  - `server/internal/util/ttlstore/*` — generic capacity-bounded TTL store with GC; tests included.
  - `server/internal/util/randutil/*` — centralized crypto randomness helpers.
  - Stores migrated (public API preserved):
    - `RegSessionStore`, `LoginSessionStore`, `TxSessionStore`.
  - ADR: `blueprint/_decisions/session-store-refactor.md` (Accepted).

- Prepared-statement repositories (read paths)
  - `server/internal/repos/credentials.go`, `server/internal/repos/transactions.go` (+ `.desc.md`).
  - Wiring: `server/cmd/api/main.go` prepares repos; `router.go` injects into handlers.
  - Adoption:
    - `/tx/signing/options` now lists credentials via repo.
    - `/tx/list` reads transactions via repo.
    - `/tx/signing/finish` reads credential owner/sign_count via repo.
  - ADR: `blueprint/_decisions/data-access-repos-prepared.md` (Accepted).

- Frontend API client and partial adoption
  - `web/src/lib/api.ts` (+ `.desc.md`) — wraps `fetch` with `credentials: 'include'` and decodes the error envelope into a typed `ApiError`.
  - `web/src/pages/Dashboard.tsx` updated to use client for `/tx/signing/options`; errors surfaced with codes.

## Verification

- Build: `go build ./server/...`
- Tests (subset):
  - `go test ./server/internal/httpx/errors -v` → PASS
  - `go test ./server/internal/tx -run TestTxOptionsHandler_Happy -v` → PASS
- Manual smoke:
  - Health: `curl -i :8080/health` → 200
  - No-session `POST /tx/signing/options` → 401 envelope (`ERR_UNAUTHORIZED`).
  - Oversized `POST /tx/signing/finish` body → 413 due to body limit.

## Descriptions Updated

- `server/internal/app/app.desc.md`
- `server/internal/app/router.go.desc.md`
- `server/cmd/api/main.go.desc.md`
- `server/internal/tx/options.go.desc.md`
- `server/internal/tx/finish.go.desc.md`
- `server/internal/tx/list.go.desc.md`
- `server/internal/me/account_key.go.desc.md`
- `web/src/pages/Dashboard.tsx.desc.md`
- `server/internal/httpx/errors/envelope.go.desc.md`
- `server/internal/httpx/middleware/request_id.go.desc.md`

## Refs

Refs: goal simple-ui-and-storage; goal passkey-registration-login-uv; goal transaction-content-signing; requirement R-ERR; requirement R-PLAT-2; requirement sqlite-persistence; decision router-builder-wiring; decision http-error-envelope; decision session-store-refactor; decision data-access-repos-prepared; decision request-id-and-slog-json
