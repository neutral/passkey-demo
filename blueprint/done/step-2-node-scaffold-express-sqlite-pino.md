### Step 2 — Node: scaffold `node-server/` (Express + SQLite + Pino)

Scope

- Add a minimal Node backend skeleton (`node-server/`) to run alongside the web app, without changing any API semantics yet.
- Provide Express bootstrap, env config, JSON logging via Pino, health route, and placeholders for DB, errors, and request-id middleware.
- Keep implementation small and ESM-only JS to match the demo scope; postpone feature parity (CORS/cookies/sessions/WebAuthn) to later steps.

Source to add/modify

- Add `node-server/package.json` (ESM):
  - Dependencies: `express`, `pino`, `pino-http`, `dotenv`, `nanoid`, `better-sqlite3`, `cbor-x`, `cookie`, `cors` (installed now, wired later), `@simplewebauthn/server` (for future steps).
  - Dev/test: use built-in `node:test` (no extra devDeps). Add scripts: `start`, `dev`, `test`.
- Add `node-server/src/config.js`:
  - Read env: `RP_ID`, `ORIGIN`, `PORT` (default `8080`), `DB_PATH` (default `server/demo-node.db`), `RP_ID_ALLOWLIST`, `ORIGIN_ALLOWLIST`.
  - Normalize/validate minimal fields; export immutable config object.
- Add `node-server/src/logger.js`:
  - Export Pino instance + `pino-http` middleware; helper `logServerStart(config)` emits `server_start` with key fields.
- Add `node-server/src/reqid.js` (placeholder):
  - Express middleware to attach/request `X-Request-ID` (UUID/nanoid); store on `req` for logs.
- Add `node-server/src/error.js` (placeholder):
  - Helper to write JSON error envelope `{ code, error, correlation_id? }`.
- Add `node-server/src/db.js` (placeholder):
  - Open `better-sqlite3` at `DB_PATH`; apply PRAGMAs; export handle. Postpone schema creation to Step 3.
- Add `node-server/src/migrations.sql` (placeholder):
  - Empty or comment-only file to be filled in Step 3 with tables identical to Go backend.
- Add `node-server/src/server.js`:
  - Load `.env` (dotenv), build config, init logger + req-id middleware, add `express.json()` body parser.
  - Add `GET /health` that returns `{ status: 'ok' }` with `Content-Type: application/json`.
  - Start `app.listen(PORT)` and log `server_start` (include `rp_id`, `origin`, `port`, `db_path`).

Description files

- Add `node-server/node-server.desc.md` — overview of Express bootstrap, logging/event names, and placeholders for DB/session/CORS; describe how it relates to the web app; Refs updated.
- Add `node-server/src/server.js.desc.md` — responsibilities (bootstrap, health, logging); invariants.
- Add `node-server/src/config.js.desc.md` — env mapping, defaults, validation rules.
- Add `node-server/src/logger.js.desc.md` — event naming, correlation id propagation plan.
- Add `node-server/src/db.js.desc.md` — PRAGMAs, migration strategy; note Step 3 for schema parity.
- Add `node-server/src/error.js.desc.md`, `node-server/src/reqid.js.desc.md` — envelope and headers; relation to later steps.

Blueprint updates (requirements/specs/ADRs/user-flows to add/update)

- Requirements (NFR): add Draft `blueprint/global/single-node-backend/requirement.md` — minimal Node backend parity target and local dev constraints (Express, SQLite, JSON logs, .env config).
- Specs (NFR): add `blueprint/global/single-node-backend/_specs/spec.md` — cover interfaces (`/health`), logging event `server_start`, env validation, DB open only (schema in Step 3), testing strategy using `node:test`.
- Specs (OPS/Dev): update `blueprint/global/easy-local-dev/_specs/spec.md` to mention Node server as an alternative to Go during local runs (ports, commands). Keep CORS content unchanged (handled in Step 4).
- ADR (Draft): add `blueprint/_decisions/structured-logging-with-pino-guidelines.md` — language-agnostic logging parity with Pino (event names, minimal fields), aligning with existing slog ADRs.
- User flows: no changes for UX; local dev flow note may be added to `_user-flows` later when the Node server replaces Go in tests.
- Refs maintenance: ensure each new artifact includes `Refs:` to: requirement R-OPS-DEV, requirement R-PLAT-3, decisions `router-builder-wiring`, `request-id-and-slog-json`, `structured-logging-with-slog-guidelines` (and new Pino ADR).

Request/response shape

- `GET /health` (200): `{ status: 'ok' }`
- Log `server_start` (stdout JSON, via Pino): `{ event: 'server_start', rp_id, origin, port, db_path, correlation_id? }`.

Algorithm

- Initialize dotenv, read env, validate minimal fields; construct config.
- Create Express app; attach `pino-http` and request-id middleware; add `express.json()` for future endpoints.
- Health route: return `{ status: 'ok' }` with appropriate headers.
- Start listening; emit `server_start` with config fields.

Database interactions

- Step 2: open the SQLite database handle and apply PRAGMAs only. Defer schema creation to Step 3.

Policies & limits

- No CORS, cookies, or rate limits in this step (placeholders only). These are added in Steps 4 and 14.
- Logging policy: JSON logs only; no sensitive values; event names must be stable (`server_start`).

Sequencing

- This scaffold intentionally excludes CORS/cookies/sessions (Step 4), WebAuthn (Steps 5–8), DB schema (Step 3), and limits (Step 14).
- Keep file/module names stable to avoid churn across steps.

Tests

- Strategy: Use Node’s built-in `node:test` and `assert` to avoid extra devDeps. Happy path is REQUIRED.
- Files to add (examples):
  - `node-server/test/health.test.js` — start app on ephemeral port, request `/health`, expect 200 and `{ status: 'ok' }` JSON.
  - `node-server/test/logger.test.js` — unit test `logServerStart(config)` returns/prints an object with `event='server_start'` and required keys.
  - `node-server/test/config.test.js` — env parsing: defaults for `PORT` and `DB_PATH`, error on missing `RP_ID`/`ORIGIN` if we require them at boot (else warn-only and pass).
- Negative cases:
  - Invalid `PORT` (non-numeric) → boot aborts with clear error.
  - DB open failure (unwritable path) → process exits non-zero; log includes `error_kind='db_open_failed'`.
- Commands:
  - `npm -C node-server i`
  - `node --test node-server/test/*.test.js`
- Expected outcomes:
  - All tests pass; health returns 200; log helper produces correct fields.

Verification

- After implementation, verify locally:
  - `npm -C node-server i && node node-server/src/server.js`
  - Observe a `server_start` JSON log with expected fields.
  - `curl -sS :8080/health -i` → `HTTP/1.1 200 OK` and body `{"status":"ok"}`.
- User verification commands

```bash
# Install and start the Node server
npm -C node-server i
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo-node.db \
  node node-server/src/server.js &
API_PID=$!

# Verify health and logs
curl -sS :8080/health -i | sed -n '1,5p'

# Clean up
kill $API_PID || true
```

Acceptance criteria

- `node-server/` boots with `node src/server.js`, exposes `GET /health` (200 JSON), and emits a `server_start` JSON log with config fields.
- Placeholders (`config`, `logger`, `reqid`, `error`, `db`, `migrations.sql`) exist and are referenced by `server.js`.
- Tests planned as above; happy path and negative cases identified.

Notes

- Keep the scaffold minimal; defer features to their respective steps to reduce risk and change surface.
- Prefer built-in test runner (`node:test`) for zero-deps; switch to a richer runner later only if needed.

Refs: goal simple-ui-and-storage; requirement R-PLAT-3; requirement R-OPS-DEV; decision router-builder-wiring; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines

