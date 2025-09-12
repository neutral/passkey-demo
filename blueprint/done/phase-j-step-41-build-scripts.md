# Step 41 — Build scripts (Done: 2025-09-12)

## Scope
- Standardize local dev commands in the root `Makefile` to build and run the Go server and Vite web app.
- Add a `run` target that starts both processes with basic signal handling to stop them together.

## Source to add/modify
- Modify `Makefile` (root):
  - Ensure `.PHONY` includes `server web build clean run test test-all`.
  - Keep existing `server` and `web` targets (paths confirmed).
  - Add `run` target that:
    - Launches `go run ./cmd/api` in `server/` and `npm run dev` in `web/` concurrently.
    - Traps `INT`/`TERM` to stop both child processes.
    - Surfaces server logs (slog) and Vite logs in the same terminal.
  - Keep `build` and `clean` targets unchanged; ensure they succeed from repo root.
  - Add test-related targets for Go (and optional web):
    - `test`: default Go tests in short mode from `server/` with package and -run filtering support.
      - Usage: `make test` or `make test PKG=./internal/webauthn` or `make test RUN=LoginFinish`.
      - Implementation: `cd server && go test -short $${PKG:-./...} $${RUN:+-run $${RUN}} -v`.
    - `test-all`: full Go test run (no `-short`). Same filtering variables supported.
      - Implementation: `cd server && go test $${PKG:-./...} $${RUN:+-run $${RUN}} -v`.
    - Optional `test-web` (only if package.json has a test script): `cd web && npm test -- --watch=false`.

## Description files
- Create/Update (if tracked by desc-check): `Makefile.desc.md` with a short description of targets and usage; note that `run` starts both and uses a trap to stop them together.

## Request/response shape
- N/A (dev tooling only).

## Algorithm
- Implement `run` with portable shell primitives:
  - Start server in background: `(cd server && RP_ID=$${RP_ID:-localhost} ORIGIN=$${ORIGIN:-http://localhost:5173} PORT=$${PORT:-8080} DB_PATH=$${DB_PATH:-server/demo.db} go run ./cmd/api) & SERVER_PID=$$!`
  - Start web in background: `(cd web && npm run dev) & WEB_PID=$$!`
  - `trap 'kill $$SERVER_PID $$WEB_PID 2>/dev/null || true' INT TERM`
  - `wait` for both; exit when either terminates; trap ensures clean shutdown.
- Do not hardcode env; allow user to supply `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH` through the environment or `.env` in a later step.
- Implement `test`/`test-all` to accept variables:
  - `PKG`: Go package path (default `./...`).
  - `RUN`: regex passed to `-run` to filter tests.
  - Keep verbose (`-v`) for better CI logs; short mode only on `test`.

## Database interactions
- None.

## Policies & limits
- The `run` target is for local development only; production should use a proper process supervisor.
- Avoid introducing new tooling dependencies (no `concurrently`, no `tmux`).

## Sequencing
- Depends on Step 5 (config) and the existing server/web scaffolds.
- Optional: In a later step, add `.env.example` (Step 44) to simplify env export before `make run`.

## Tests
- Minimal smoke verification (manual): ensure both processes start and logs print.
- Commands execute successfully from repo root on macOS/Linux shells (`bash`/`zsh`).

## Verification
- Build-only checks remain green:
  - `make build` succeeds (Go server builds; Vite builds).
- Manual run:
  - Terminal: `make run`
  - Expect: server outputs a `server_start` slog JSON line and begins listening on `:${PORT:-8080}`; web dev server prints Vite banner and local URL.
  - Ctrl-C once: both processes terminate; Makefile exits.

## User verification commands
```bash
# From repo root
make build

# Run both server and web together (Ctrl-C to stop)
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo.db \
  make run

# Run short Go tests (all packages by default)
make test

# Run short Go tests for a specific package
make test PKG=./internal/webauthn

# Run short Go tests filtered by test name regex
make test RUN=LoginFinish

# Run full Go test suite (no -short)
make test-all

# Full tests for a specific package with a filter
make test-all PKG=./internal/tx RUN=Finish
```

## Acceptance criteria
- `make run` starts both server and web from the repo root and stops both on Ctrl-C.
- `make server`, `make web`, `make build`, and `make clean` continue to work.
- `make test` runs short Go tests by default; `make test-all` runs full tests; both support `PKG` and `RUN` filters.
- No new external dependencies introduced.

## Notes
- Keep the `run` shell portable; rely only on POSIX sh features available in bash/zsh.
- If Windows dev is needed later, add a minimal PowerShell script (out of scope here).

## Refs
- Refs: requirement R-OPS-DEV; requirement R-PLAT-2

