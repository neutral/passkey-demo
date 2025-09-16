# Purpose
Define the canonical SQLite schema for the Node server. This file is executed at startup to create tables if they do not exist.

# Key Logic
- Includes four tables: `accounts`, `credentials`, `sessions`, `transactions`.
- Enforces foreign keys with `ON DELETE CASCADE` from `credentials`/`transactions`/`sessions` to `accounts`.
- Uses `BLOB` for binary, `TEXT` for ids, and `INTEGER` for timestamps/counters.

# Interactions
- Read and executed by `applyMigrations` from `src/db.js`. Called by `src/server.js` during boot.

# Refs
Refs: requirement R-PLAT-3; decision encoding-and-ceremony-guardrails
