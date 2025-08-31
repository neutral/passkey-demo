# Purpose
Open the SQLite database with safe demo-grade PRAGMAs and apply schema migrations for accounts, credentials, sessions, and transactions.

# Key Logic
- Open via `database/sql` and `github.com/mattn/go-sqlite3` using DSN with `_busy_timeout=5000` and `_foreign_keys=on`.
- Apply PRAGMAs: `journal_mode=WAL`, `synchronous=NORMAL`, `foreign_keys=ON`.
- Create tables with `CREATE TABLE IF NOT EXISTS` per blueprint R-PLAT-3.

# Interactions
- Called by application startup (or dedicated migration runner) before handlers need persistence.
- Returns an `*sql.DB` for use by repositories/handlers; caller closes the DB.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-3; requirement R-PLAT-2; requirement R-NO-BROKER; requirement R-ERR
