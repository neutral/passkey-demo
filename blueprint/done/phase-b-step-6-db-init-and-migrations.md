### Step 6 — DB init & migrations (Done: 2025-08-31)

   - Context
     - Initialize SQLite persistence to support account, credential, session, and transaction storage per R-PLAT-3. Apply schema at startup with safe `CREATE TABLE IF NOT EXISTS` statements.

   - Structure
     - Add `server/internal/storage` package (or `server/internal/db`): `storage.go` with `Open(cfg *config.Config) (*sql.DB, error)` and `Migrate(db *sql.DB) error`.
     - PRAGMAs on open: `foreign_keys=ON`, journal_mode=WAL, synchronous=NORMAL (demo-grade), busy_timeout=5000.

   - Source to add (instructions only)
     - `server/internal/storage/storage.go`:
       - Open database at `cfg.DBPath` using `github.com/mattn/go-sqlite3` via `database/sql`.
       - Set connection pool (e.g., `SetMaxOpenConns(1)` for SQLite; `SetConnMaxIdleTime` reasonable).
       - Execute PRAGMAs and call `Migrate` with the schema:
         - Tables from R-PLAT-3 (accounts, credentials, sessions, transactions) with `IF NOT EXISTS` and `FOREIGN KEY` constraints.
       - Provide `Close()` responsibility to caller (main) later; for now, return `db`.
     - Integration note: wire into `cmd/api/main.go` later (when handlers need DB), keeping this step focused on package creation and migrations.

   - Description files to add (instructions only)
     - `server/internal/storage/storage.go.desc.md`: Purpose (SQLite open + migrate), Key Logic (PRAGMAs, schema), Interactions (used by main/handlers), Refs.
       - Refs: goal simple-ui-and-storage; requirement R-PLAT-3; requirement R-PLAT-2; requirement R-ERR.

   - Blueprint updates
     - Refs to include upon implementation: requirement R-PLAT-3; goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-NO-BROKER.

   - Verification (to run after implementation)
     - Build: `cd server && go build ./...` (expect exit 0).
     - Run a tiny snippet (temporary or via main if already integrated) to call `storage.Open(cfg)` then `storage.Migrate(db)`.
     - Confirm DB file exists: `test -f server/demo.db`.
     - Optional (if `sqlite3` CLI available): `sqlite3 server/demo.db '.schema'` shows the four tables with expected columns.

   - User verification commands (copy/paste)

     ```bash
     # Build server with storage package present
     cd server && go build ./... && cd -

     # Quick migration runner (inline Go) — does not modify app code
     cd server
     cat > /tmp/migrate.go <<'EOF'
     package main
     import (
       "log"
       cfgpkg "github.com/neutral/passkey-demo/internal/config"
       store "github.com/neutral/passkey-demo/internal/storage"
     )
     func main(){
       cfg, err := cfgpkg.Load(); if err!=nil{ log.Fatal(err) }
       db, err := store.Open(cfg); if err!=nil{ log.Fatal(err) }
       defer db.Close()
       if err := store.Migrate(db); err!=nil { log.Fatal(err) }
       log.Println("migrated ok")
     }
     EOF
     PORT=0 go run /tmp/migrate.go
     test -f server/demo.db && echo OK:db-exists
     # Optional schema view
     command -v sqlite3 >/dev/null && sqlite3 server/demo.db '.schema' | sed -n '1,60p'
     cd -
     ```

   - Notes
     - Keep PRAGMAs demo-grade; for production, review durability/performance trade-offs.
     - Foreign keys must be enabled for relational integrity; use `ON DELETE CASCADE` as specified.

