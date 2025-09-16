# Purpose
Validate that `loadConfig` sets sane defaults, parses comma-separated allowlists, and rejects invalid numeric values.

# Key Logic
- Defaults: `PORT=8080`, `DB_PATH='demo.db'`.
- Lists: `RP_ID_ALLOWLIST` parsed into trimmed array; empty values → empty array.
- Invalid `PORT` throws.

# Interactions
- Imports `loadConfig` from `src/config.js`.

# Refs
Refs: requirement R-OPS-DEV
