# Step 40 — Logging (Done: 2025-09-12)

## Verification Notes
- Build succeeded: `cd server && go build ./...`.
- Short tests passed: `cd server && go test -short ./...`.
- Removed all stdlib `log.*` usages from `server/`; only slog remains.
- Description files and specs updated to align with structured logging changes.

## Scope
- Introduce consistent structured logging using `log/slog` with JSON output and event-style messages.
- Emit logs for key flows with stable event names: `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`.
- Include correlation id (`X-Request-ID`), environment/config context, and safe identifiers (thumbprints and IDs) without logging raw materials.

## Source to add/modify
- Add `server/internal/logging/logging.go`: logger initialization helper
  - `GetLevelFromEnv() slog.Level` (reads `LOG_LEVEL`: debug|info|warn|error; default info)
  - `New(options) *slog.Logger` (JSON handler on `stdout`, `LevelVar` for dynamic updates; optional `AddSource` in dev only)
  - `FromContext(ctx context.Context) *slog.Logger` (returns `slog.Default()`; placeholder for future context-aware handler)
- Modify `server/cmd/api/main.go`: initialize default slog logger
  - Read `LOG_LEVEL`/`LOG_FORMAT` (json|text); set `slog.SetDefault(...)`; replace `log.Printf/Fatal` with slog
  - Startup logs with `event="server_start"` and attributes `rp_id`, `origin`, `port`, `db_path`
- Modify handlers to log at success/error boundaries with typed attributes (no raw materials):
  - `server/internal/webauthn/reg_options.go`: on success emit `event="reg_options"`, `correlation_id`, `expires_at`, `rp_id`, `origin`, and `session_id_len`
  - `server/internal/webauthn/reg_finish.go`: on success emit `event="reg_finish"`, `correlation_id`, `account_thumb_hex`, `credential_id_hash`, `sign_count`
  - `server/internal/webauthn/login_options.go`: on success emit `event="login_options"`, `correlation_id`, `expires_at`, `rp_id`, `origin`
  - `server/internal/webauthn/login_finish.go`: replace `log.Printf` failure line with `webauthn.LogAssertion` and log success `event="login_finish"` with `account_thumb_hex`, `credential_id_hash`, `sign_count`
  - `server/internal/tx/options.go`: replace `log.Printf` branches with `event="tx_options"` logs; on success emit `tx_id_hex` and `allow_count`
  - `server/internal/tx/finish.go`: on success emit `event="tx_finish"` with `tx_id_hex`, `latency_ms` (if available), and `account_thumb_hex`; on verify errors rely on `MapVerifyError`+single structured log

## Description files
- Create `server/internal/logging/logging.go.desc.md` (purpose, envs, handler options; Refs below).
- Update for modified files to describe new logging behavior and invariants:
  - `server/cmd/api/main.go.desc.md` (startup logger init, event naming)
  - `server/internal/webauthn/reg_options.go.desc.md` (emits `reg_options`)
  - `server/internal/webauthn/reg_finish.go.desc.md` (emits `reg_finish`, privacy rules)
  - `server/internal/webauthn/login_options.go.desc.md` (emits `login_options`)
  - `server/internal/webauthn/login_finish.go.desc.md` (emits `login_finish`; verification failures use `LogAssertion` with safe fields)
  - `server/internal/tx/options.go.desc.md` (emits `tx_options` with `tx_id_hex` on success; error mapping)
  - `server/internal/tx/finish.go.desc.md` (emits `tx_finish` and maps verify/policy errors)

## Request/response shape
- No API response changes. This step defines log event shapes (not user-facing):
  - Common attrs: `correlation_id` (from context), `rp_id`, `origin`
  - reg_options: `expires_at`
  - reg_finish: `account_thumb_hex`, `credential_id_hash?`, `sign_count`
  - login_options: `expires_at`
  - login_finish: `account_thumb_hex`, `credential_id_hash`, `sign_count`, `outcome`
  - tx_options: `tx_id_hex`, `allow_count`
  - tx_finish: `tx_id_hex`, `stored` (bool), `outcome`

## Algorithm
- Logger initialization
  - In `main.go`, construct a `slog.LevelVar` and set from `LOG_LEVEL`; choose handler by `LOG_FORMAT`.
  - `slog.SetDefault(logger)` early; log one `server_start` event with configuration context (omit secrets).
- Correlation id
  - In each handler, attempt `request_id := middleware.FromContext(r.Context())`; when present, include `slog.String("correlation_id", request_id)`.
- Emission points and invariants
  - Emit exactly one success log per handler call at the end of successful processing with the event name and safe fields.
  - Emit at most one failure log per request path for verify/policy errors using `MapVerifyError`/`MapPolicyError` and do not log raw materials.
  - Use typed attributes only (`slog.String/Int/Bool/Uint64`)—no alternating key,value style.
  - Do not log raw CBOR, signatures, credential IDs; use `webauthn.HashID` for any identifier if needed.

## Database interactions
- None added. Logging does not modify DB state. Ensure logging occurs after DB writes on success to capture final `tx_id_hex` where applicable.

## Policies & limits
- Privacy: never log raw credential IDs, signatures, CBOR, or secrets. Use hashed identifiers only (hex SHA‑256). Maintain this invariant in all logs.
- Performance: avoid `AddSource` in production; guard any expensive attribute construction behind `logger.Enabled(ctx, slog.LevelDebug)`.
- Volume: default level `info`. No sampling in this step; revisit if logs become chatty.

## Sequencing
- Depends on existing `RequestID` middleware and error envelope wiring (already present via router builder).
- Keep current HTTP response behavior unchanged; logs are additive and must not change status codes.
- If later adding `sloglint` in CI, ensure all new logs use typed attributes to satisfy lint.

## Tests
- Strategy: capture logs with a custom in‑memory `slog.Handler` in unit tests to assert event names and attributes; no external I/O.
- Files to add:
  - `server/internal/logging/logging_test.go`: env level parsing matrix (debug|info|warn|error → expected `slog.Level`).
  - `server/internal/webauthn/login_finish_log_test.go`: drive a minimal happy path for `LoginFinishHandler` using test doubles/mocks where possible; attach a test logger via `slog.SetDefault` and assert a `login_finish` record with `account_thumb_hex` and `credential_id_hash` present.
  - `server/internal/tx/options_log_test.go`: call `TxOptionsHandler` with a valid session context; assert a `tx_options` record with `tx_id_hex` and `allow_count`.
  - `server/internal/tx/finish_log_test.go`: happy path yields a `tx_finish` record with `tx_id_hex`, `stored=true`.
  - Negative cases: induce verify errors to assert only one failure log is emitted with `error_kind` mapped via `MapVerifyError`, and no raw materials appear.
- Commands:
  - `go test ./server/internal/... -run Log -v`
  - Expect: tests assert event names and required attributes; no test writes to files/stdout.

## Verification
- Build: `go build ./server/...` succeeds.
- Manual (dev):
  - Terminal A: `make server` (ensure `LOG_LEVEL=info` default, `LOG_FORMAT=json`).
  - Terminal B: run the web app (`make web`) and drive Register → Login → Sign Message flows.
  - Observe server stdout: JSON entries with events `reg_*`, `login_*`, `tx_*`; success logs include `account_thumb_hex` and `tx_id_hex` where applicable; each request carries `correlation_id` when present.
- If any test fails, fix forward and re‑run the test subset; repeat until green.

## User verification commands
```bash
# Build server
go build ./server/cmd/api

# Run server with JSON logs at info level
LOG_LEVEL=info LOG_FORMAT=json RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo.db \
  go run ./server/cmd/api

# In another terminal: run the web app and complete Register → Login → Sign
make web
# Watch server stdout for event entries:
#   {"msg":"reg_options", ...}
#   {"msg":"reg_finish", "account_thumb_hex":"..."}
#   {"msg":"login_options", ...}
#   {"msg":"login_finish", "account_thumb_hex":"..."}
#   {"msg":"tx_options", "tx_id_hex":"..."}
#   {"msg":"tx_finish", "tx_id_hex":"...", "stored":true}
```

## Acceptance criteria
- Startup logs use slog JSON with `server_start` event and config attributes; no secrets.
- Each of the six events is emitted once on success with the specified attributes.
- Failure logs for verify/policy cases use a single structured line with `error_kind` and no raw materials.
- Logs include `correlation_id` when the RequestID middleware is active.
- Unit tests for logging pass (`go test ./server/internal/... -run Log`).

## Notes
- Keep event names stable; prefer lower_snake_case attribute keys.
- Consider adding `sloglint` in a follow‑up to enforce typed attributes (`attr-only: true`).
- Consider `tint` for dev text output and sampling/dedup handlers if volume grows.

## Refs
- Refs: requirement R-ERR; decision http-error-envelope; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines

## Appendix — Repo-wide updates (documentation + cleanup)
- Converted legacy stdlib logging to slog:
  - `server/internal/webauthn/att.go`: replaced `log.Printf` with `slog.Info` debug events (`reg_finish_debug`) carrying only safe metadata.
- Description files updated to reflect structured logging and correlation:
  - `server/server.desc.md`: added Observability section (events, attributes, error_kind, correlation_id).
  - `server/internal/httpx/errors/envelope.go.desc.md`: envelope now widely adopted; correlation_id documented.
  - `server/internal/app/router.go.desc.md`: notes RequestID is outermost to correlate logs/envelopes.
  - `server/internal/webauthn/*.desc.md` and `server/internal/tx/*.desc.md`: success/failure event logs documented per handler.
  - `server/internal/logging/logging.go.desc.md`: LOG_LEVEL/LOG_FORMAT and handler construction documented.
- Specs and decisions updated to prevent drift:
  - `blueprint/global/single-go-backend/_specs/spec.md`: Observability and privacy guardrails for slog; event names listed.
  - `blueprint/_decisions/request-id-and-slog-json.md`: status moved to Accepted for slog JSON; Step 40 implementation noted; Refs added.
- Repo-wide verification performed:
  - Removed all `log.Print*` usages and `import "log"` from `server/`.
  - Build and short tests pass: `cd server && go build ./... && go test -short ./...`.

