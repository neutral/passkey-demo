# Purpose
Open a SQLite database using `better-sqlite3`, apply base PRAGMAs, and run schema migrations from `migrations.sql`.

# Key Logic
- `openDB(path)`: detects SQLite in-memory targets (`:memory:`, `file::memory:`, `mode=memory` URIs) and passes them straight to `better-sqlite3`; otherwise resolves the path relative to `process.cwd()`, ensures the parent directory exists (creating it recursively if missing), then opens the DB and sets `WAL`, `synchronous=NORMAL`, and `foreign_keys=ON`.
- `applyMigrations(db, sqlPath)`: reads the SQL file and executes within a transaction; idempotent via `CREATE TABLE IF NOT EXISTS`.

# Interactions
- Used by `src/server.js` at startup to initialize schema; tests call it directly to validate tables/PRAGMAs.

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV; decision encoding-and-ceremony-guardrails
