# Purpose
Bootstrap the Express app, attach request-id and HTTP logging, mount WebAuthn registration/login routes plus authenticated account/transaction endpoints, expose `/health`, and start listening on the configured port while emitting structured startup logs.

# Key Logic
- `createApp(config, db)`: sets up middlewares (request-id, logging, CORS, session loader), wraps `/authn/*` and `/tx/*` groups with `buildRateLimiter` + `buildBodyLimit` (1 MiB JSON parser) before mounting registration/login/tx routers, attaches `/me` routes, and installs the shared error handler for parser overflows.
- `start()`: loads config, opens DB (PRAGMAs), applies migrations from `migrations.sql`, logs `db_init` (PRAGMAs), builds app, starts the server, and logs `server_start` plus `listening` events.

# Interactions
- Depends on `config.js`, `logger.js`, `reqid.js`, `db.js`, `session.js`, WebAuthn routers, `me.js`, and transaction routes (`tx/options.js`, `tx/finish.js`, `tx/list.js`) plus `limits.js`/`error.js` middleware. Used directly by `npm start`; tests import `createApp` to bind ephemeral ports while downstream routers emit structured events (`reg_options`, `login_options`, `tx_options`, `tx_finish`).

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV; requirement R-FLOW-REG; requirement R-FLOW-SIGN; decision router-builder-wiring; decision request-id-and-slog-json
