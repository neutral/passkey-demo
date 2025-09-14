# Purpose
Bootstrap the Express app, attach request-id and HTTP logging, expose a basic `/health` endpoint, and start listening on the configured port while emitting a `server_start` log.

# Key Logic
- `createApp(config, db)`: sets up middlewares and routes; returns an Express instance.
- `start()`: loads config, opens DB (PRAGMAs), builds app, starts the server, and logs `server_start` and `listening` events.

# Interactions
- Depends on `config.js`, `logger.js`, `reqid.js`, and `db.js`. Used directly by `npm start`; tests import `createApp` to bind ephemeral ports.

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV; decision router-builder-wiring; decision request-id-and-slog-json

