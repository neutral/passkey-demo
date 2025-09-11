# 37B — Refactoring Plan

## Purpose

- Translate 37b-refactor.md findings into an actionable, low‑risk sequence.
- Improve separation of concerns, consistency, and testability without changing external behavior.
- Define concrete steps, acceptance criteria, and verification for each change.

## Drivers & Pain Points (from 37b)

- Mixed concerns: handlers blend HTTP transport, domain policy, and persistence.
- Inconsistent orchestration: partial use of “BuildX” pattern across flows.
- Scattered SQL: queries duplicated in handlers; no centralized repos or prepared statements.
- Session stores duplicated: three near‑identical in‑memory stores, no generic TTL/GC.
- Error/Logging inconsistencies: ad‑hoc `http.Error`, mixed logging, no request IDs.
- Middleware underused at router: body/rate limits not consistently applied per route group.

## Target Architecture (high‑level)

- Application layer (pure orchestrators) with services per use‑case: RegistrationService, LoginService, SigningService, TxService.
- Repositories with prepared statements: SessionsRepo, CredentialsRepo, TransactionsRepo, AccountsRepo.
- Router assembler: `buildRouter(cfg, db, deps)` centralizes routes and applies middleware per group.
- Error envelope module: `internal/httpx/errors` for `{code, error, correlation_id}` and status mappers.
- Observability: standardize on `slog` JSON logger; add request ID middleware; hash identifiers in logs via existing `HashID`.
- Utilities: shared `randutil.Bytes`; generic TTL store with GC and typed wrappers for reg/login/tx.

## Proposed Package Layout

- `internal/app/`
  - `router.go` — builds mux; wires middlewares and routes; groups by `/authn/*`, `/tx/*`.
  - `services/` — orchestration services (registration, login, signing, tx).
- `internal/repos/`
  - `sessions.go`, `credentials.go`, `transactions.go`, `accounts.go` (prepare/close + methods).
- `internal/httpx/`
  - `errors/` — JSON envelope + central mappers; helpers for writing errors.
  - `middleware/` — `request-id` (optional wrappers for rate/body limit if colocated).
- `internal/util/`
  - `randutil/` — `Bytes(n int) ([]byte, error)`.
  - `ttlstore/` — generic capacity‑bounded TTL store with GC; typed wrappers for reg/login/tx.

## Incremental Migration Plan

### Step 1 — Router builder and middleware wiring

- Outcome: One place wires routes and wraps group middlewares; `main.go` delegates to `buildRouter`.
- Actions:
  - Add `internal/app/router.go` with `func BuildRouter(cfg *config.Config, db *sql.DB, deps Deps) http.Handler`.
  - Move all `mux.Handle` calls from `server/cmd/api/main.go` to `BuildRouter`.
- Apply `BodyLimitMiddleware` and `RateLimitMiddleware` per group (`/authn/*`, `/tx/*`).
- Keep CORS outermost, session middleware immediately inside CORS.
- Remove handler‑level cookie/DB fallbacks for auth; protected endpoints must rely on session middleware only.
- Acceptance:
  - Server boots; all routes reachable; CORS/session behavior unchanged.
- Body limit and rate limit enforced on `/authn/*` and `/tx/*`.
- Protected handlers no longer perform cookie→DB auth lookups; session middleware is the sole auth source.
- Verification:
  - Build: `go build ./server/...`
  - Run: `RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./server/cmd/api`
  - Curl health: `curl -i :8080/health` → 200.
  - Oversized request to `/tx/signing/finish` returns 413.

### Step 2 — Standard error envelope and mappers

- Outcome: Standard JSON error shape and centralized mappers; adopt in one handler first (e.g., `tx/options`).
- Actions:
  - Add `internal/httpx/errors` with type `{ code, error, correlation_id? }` and helpers: `Write(w, status, code, msg)`.
  - Consolidate/bridge existing `MapPolicyError` and `MapVerifyError` under `httpx/errors`.
  - Switch `server/internal/tx/options.go` to use envelope; keep status codes aligned with R-ERR table.
- Acceptance:
  - Modified endpoint returns envelope on error; status codes match mapping.
  - Unit tests cover mapping and JSON shape.
- Verification:
  - `go test ./server/internal/...`
  - Manual: request without session to `/tx/signing/options` returns `401` body `{"code":"ERR_UNAUTHORIZED",...}`.

### Step 3 — Generic TTL session store + `randutil`

- Outcome: One generic TTL store with GC; typed wrappers for reg/login/tx; shared randomness utility.
- Actions:
  - Add `internal/util/randutil.Bytes`; replace all `randBytes` duplicates.
  - Add `internal/util/ttlstore` with capacity, TTL, and background GC.
  - Replace `RegSessionStore`, `LoginSessionStore`, `TxSessionStore` impls to wrap generic store (preserve public API).
- Acceptance:
  - Existing behavior unchanged; stores expire as before; reduced duplication; unit tests for TTL/GC.
- Verification:
  - `rg "randBytes\(" server` returns none; only `randutil.Bytes` remains.
  - `go test ./server/internal/webauthn ./server/internal/tx` passes.

### Step 4 — Repositories and prepared statements (read paths first)

- Outcome: Centralized SQL; handlers/services call repos; prepared statements created at startup and closed on shutdown.
- Actions:
- Add `internal/repos/{sessions,credentials}` with `Prepare(db) (*Repo, error)` and `Close()`.
- Update `http/session.go` and `tx/options.go` to use repos, not inline SQL.
- Introduce a minimal DB interface (e.g., `DBTX` with `ExecContext`, `QueryRowContext`, `QueryContext`) to enable fakes in unit tests.
- Inject repos via a light `Deps` container from `main.go`/`BuildRouter`.
- Acceptance:
  - Behavior unchanged; prepared statements exercised; unit tests green.
- Verification:
  - `go test ./server/internal/http ./server/internal/tx`.
  - Optional: add small repo tests using SQLite temp DB.

### Step 5 — Unify “BuildX” orchestration for finish flows

- Outcome: Thin handlers, service orchestrators mirror `BuildTxFinish` for login/reg finish.
- Actions:
  - Extract `BuildLoginFinish` and `BuildRegFinish` into `internal/app/services` (or `internal/webauthn` if kept local), matching `tx/finish` pattern.
  - Handlers focus on I/O and error envelope; domain/persistence live in builders/services.
- Acceptance:
  - Behavior unchanged; duplicate logic removed; easier unit testing of builders.
- Verification:
  - `go test ./server/internal/webauthn ./server/internal/tx` with existing tests.

### Step 6 — Logging and request IDs

- Outcome: Structured logs with correlation; error envelope may include `correlation_id`.
- Actions:
  - Initialize `slog` JSON logger in `main`; add `request-id` middleware (ULID/UUID per request) and attach to context.
  - Ensure verification logs include hashed identifiers; plumb `trace_id` into error writer as optional `correlation_id`.
- Acceptance:
  - Logs in JSON with `trace_id`; envelope optionally includes same id.
- Verification:

  - Run server; hit any endpoint; logs include `trace_id` field; error responses include `correlation_id` when available.

### Step 7 — Frontend API client and envelope mapping

- Outcome: UI consumes the standard error envelope uniformly and maps codes to toasts/messages.
- Actions:
  - Add `web/src/lib/api.ts` (or `apiClient.ts`) that wraps `fetch`, sets `credentials: 'include'`, decodes `{code,error,correlation_id}` on non‑2xx, and throws a typed `ApiError`.
  - Update pages (`Register`, `Login`, `Dashboard`) to use the client and centralize toast/error mapping based on `code` (e.g., `ERR_UNAUTHORIZED`, `ERR_FORBIDDEN`, `ERR_RATE_LIMIT`).
- Acceptance:
  - Duplicate error handling logic removed from pages; consistent toasts/messages based on envelope `code`.
  - Regression: functional flows still succeed; errors show friendly messages.
- Verification:
  - `npm -C web run build` succeeds.
  - Manual: trigger a 401 on `/tx/signing/options` from the dashboard; toast shows mapped message; console logs include `correlation_id` (if surfaced).

## Quick Wins (tiny diffs)

- Wire `BodyLimitMiddleware` and `RateLimitMiddleware` around `/authn/*` and `/tx/*` now.
- Replace duplicate `randBytes` with `randutil.Bytes`.
- Use session middleware consistently; remove cookie fallback DB lookups from handlers.
- Standardize `Content-Type: application/json` and error envelope where touched.

## Logging Requirements (consistency)

- All verification logs and IDs must use hashed identifiers (`HashID`) — never log raw credential IDs, session IDs, or keys.
- Error envelope may include `correlation_id` matching request ID middleware.

## Risks & Mitigations

- Behavior regression while moving handlers: mitigate with incremental steps, existing unit tests, and endpoint smoke checks per step.
- Prepared statements lifecycle: ensure `Close()` on shutdown; cover with tests.
- Envelope adoption drift: adopt in one handler first; codify mapping in tests, then roll out.
- GC in TTL store: keep bounds conservative; test under race detector locally.

## Candidate Implementation Tasks

- [ ] backend: R-ERR → Build router and group middlewares
  - Decision: router-builder-wiring
  - Refs: goal simple-ui-and-storage; requirement R-ERR
  - Notes: move mux wiring to `internal/app/router.go`; apply CORS→Session→Group Limits.
  - State: Draft
- [ ] backend: R-ERR → Introduce error envelope and mappers
  - Decision: http-error-envelope
  - Refs: requirement R-ERR; decision encoding-and-ceremony-guardrails
  - Notes: switch `tx/options` first; add tests.
  - State: Draft
- [ ] backend: maintainability → Generic TTL store and randutil
  - Decision: session-store-refactor
  - Refs: goal simple-ui-and-storage
  - Notes: replace all `randBytes`; keep public APIs for stores.
  - State: Draft
- [ ] backend: maintainability → Repositories with prepared statements (read paths)
  - Decision: data-access-repos-prepared
  - Refs: goal simple-ui-and-storage; requirement sqlite-persistence
  - Notes: sessions/credentials first; inject via `Deps`.
  - State: Draft
- [ ] backend: maintainability → Unify BuildX orchestration for finish flows
  - Decision: application-services-layer
  - Refs: goal passkey-registration-login-uv; goal transaction-content-signing
  - Notes: extract builders; handlers use envelope; tests remain green.
  - State: Draft
- [ ] backend: R-ERR → Structured logging + request IDs
  - Decision: request-id-and-slog-json
  - Refs: requirement R-ERR
  - Notes: add `correlation_id` to envelope when present.
  - State: Draft

## Description Files to Add/Update when Implementing

- `server/internal/app/app.desc.md` and `server/internal/app/router.go.desc.md`.
- `server/internal/app/services/*.desc.md` for new orchestrators.
- `server/internal/repos/*.desc.md` for each repo.
- `server/internal/httpx/errors/*.desc.md` and `server/internal/httpx/middleware/request_id.go.desc.md`.
- `server/internal/util/randutil/*.desc.md`, `server/internal/util/ttlstore/*.desc.md`.

## Verification Strategy (overall)

- Unit tests must remain green: `go test ./server/...`.
- Endpoint smoke tests (dev):
  - `curl -i :8080/health` → 200
  - No‑session `POST /tx/signing/options` → 401 envelope (`ERR_UNAUTHORIZED`) after Step 2.
  - Oversized payload to `/tx/signing/finish` → 413 after Step 1.
- Logs: JSON shape with `trace_id` present after Step 6.

## Refs

Refs: goal simple-ui-and-storage; goal passkey-registration-login-uv; goal transaction-content-signing; requirement R-ERR; decision encoding-and-ceremony-guardrails; decision webauthn-signcount-zero-counter-policy; decision cbor-cose-interop-and-decoding-fallbacks
