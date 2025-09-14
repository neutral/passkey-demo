# Purpose
Open a SQLite database using `better-sqlite3` and apply base PRAGMAs. Schema migrations are deferred to Step 3.

# Key Logic
- `openDB(path)`: returns a single writer connection; sets `WAL`, `synchronous=NORMAL`, and `foreign_keys=ON`.

# Interactions
- Used by `src/server.js` at startup; tests may simulate open failure by pointing to an unwritable path.

# Refs
Refs: requirement R-PLAT-3; decision encoding-and-ceremony-guardrails

