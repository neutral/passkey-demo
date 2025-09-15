# Purpose
Bootstrap the Express app, attach request-id and HTTP logging, mount WebAuthn registration routes, expose a basic `/health` endpoint, and start listening on the configured port while emitting structured startup logs.

# Key Logic
- `createApp(config, db)`: sets up middlewares (request-id, logging, JSON parser, CORS, session loader), mounts `/authn/passkey/registration` using `createRegistrationRoutes`, and returns an Express instance with shared `app.locals.registration` state.
- `start()`: loads config, opens DB (PRAGMAs), applies migrations from `migrations.sql`, logs `db_init` (PRAGMAs), builds app, starts the server, and logs `server_start` plus `listening` events.

# Interactions
- Depends on `config.js`, `logger.js`, `reqid.js`, `db.js`, `session.js`, and the new `webauthn/reg.js`. Used directly by `npm start`; tests import `createApp` to bind ephemeral ports while registration routes emit `reg_options` logs.

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV; requirement R-FLOW-REG; decision router-builder-wiring; decision request-id-and-slog-json
