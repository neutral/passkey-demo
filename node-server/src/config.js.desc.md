# Purpose
Define and validate environment configuration for the Node server, mapping `.env`/process variables to a normalized config object.

# Key Logic
- `loadConfig(env)`: reads `RP_ID`, `ORIGIN`, `PORT` (default 8080), `DB_PATH` (default `server/demo-node.db`), and comma-separated allowlists; validates numeric `PORT` and warns if `RP_ID/ORIGIN` are absent during scaffold.
- Returns an immutable object used by `server.js` and other modules.

# Interactions
- Used by `src/server.js` at startup. Tests import `loadConfig` to validate defaults and error cases.

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV

