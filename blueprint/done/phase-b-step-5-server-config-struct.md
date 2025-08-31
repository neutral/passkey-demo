### Step 5 — Server config struct (Done: 2025-08-31)

   - Context
     - Centralize runtime configuration (RP ID, origin, port, DB path, allowlists) to support secure WebAuthn checks and local dev.

   - Structure
     - Add a `server/internal/config` package with a single `config.go` file, or a top-level `server/config.go` if keeping it flat (choose one; prefer `internal/config`).
     - Expose `type Config struct { RP_ID string; Origin string; Port string; DBPath string; RPAllowlist []string; OriginAllowlist []string }` and `func Load() (*Config, error)`.

   - Source to add (instructions only)
     - `server/internal/config/config.go`:
       - Read env vars: `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`, `RP_ID_ALLOWLIST` (comma-separated), `ORIGIN_ALLOWLIST` (comma-separated).
       - Defaults: `RP_ID=localhost`, `ORIGIN=http://localhost:5173`, `PORT=8080`, `DB_PATH=server/demo.db`.
       - Normalize: trim spaces; lowercase `RP_ID`; ensure `Origin` has scheme and no trailing slash.
       - Derive `RPAllowlist` (include `RP_ID` if not present) and `OriginAllowlist` (include `Origin` if not present).
       - Validate: `RP_ID` non-empty; `Origin` parses as URL; `PORT` numeric; deny wildcard origins; allowlist entries must be exact matches (no globs).
       - Log (on startup) a concise summary (rp_id, origin, port, db_path) without secrets.
     - Integration note: refactor `cmd/api/main.go` later to call `config.Load()` and bind to `cfg.Port` (tracked in a later step to avoid scope creep here).

   - Description files to add (instructions only)
     - `server/internal/config/config.go.desc.md`: Purpose (central config), Key Logic (env parsing, defaults, validation), Interactions (used by main and handlers), Refs.
       - Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-PORTABLE.

   - Blueprint updates
     - Refs to include upon implementation: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-PORTABLE; requirement R-SEC-UV (policy alignment).

   - Verification (to run after implementation)
     - Build: `cd server && go build ./...` (expect exit 0).
     - Run with defaults: `cd server && RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./cmd/api` and observe startup log contains rp_id/origin/port.
     - Invalid config: `cd server && ORIGIN=bad go run ./cmd/api` should log/return a clear error from `config.Load()`.

   - User verification commands (copy/paste)

     ```bash
     # Build with config package present
     cd server && go build ./... && cd -

     # Run with explicit envs and observe log (Ctrl+C to stop)
     cd server
     RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo.db \
       go run ./cmd/api > /tmp/step5_server.log 2>&1 & echo $! > /tmp/step5_api.pid
     sleep 1
     grep -E "rp_id|origin|listening" -i /tmp/step5_server.log | sed -n '1,5p'
     kill $(cat /tmp/step5_api.pid) && rm -f /tmp/step5_api.pid
     cd -

     # Invalid origin should fail fast
     cd server && ORIGIN=bad go run ./cmd/api || echo "expected failure" && cd -
     ```

   - Notes
     - Keep config minimal and focused; secrets are out of scope for the demo.
     - RP/Origin allowlists backstop later security checks in Step 14 and across ceremonies.

