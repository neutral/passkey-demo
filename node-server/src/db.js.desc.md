# Purpose
Open a SQLite database using `better-sqlite3`, apply base PRAGMAs, and run schema migrations from `migrations.sql`.

# Key Logic
- `openDB(path)`: returns a single writer connection; sets `WAL`, `synchronous=NORMAL`, and `foreign_keys=ON`.
- `applyMigrations(db, sqlPath)`: reads the SQL file and executes within a transaction; idempotent via `CREATE TABLE IF NOT EXISTS`.

# Interactions
- Used by `src/server.js` at startup to initialize schema; tests call it directly to validate tables/PRAGMAs.

# Refs
Refs: requirement R-PLAT-3; decision encoding-and-ceremony-guardrails
