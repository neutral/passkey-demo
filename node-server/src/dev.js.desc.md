# Purpose
Development launcher that ensures the Node server starts on an available port.
It prefers `PORT` from the environment (default `8080`), and if that port is
occupied, it selects the next available port and sets `process.env.PORT` before
delegating to the standard `start()` function.

# Key Logic
- Reads `.env` via `dotenv/config` for convenience.
- Uses `get-port` to probe a small set of candidate ports `[PORT, PORT+1, PORT+2]`.
- Sets `process.env.PORT` to the chosen free port and calls `start()` from
  `src/server.js` to boot the Express app.
- Emits a concise warning if the preferred port is busy.

# Interactions
- Depends on `get-port` and the existing server startup in `src/server.js`.
- No changes to request handling; only affects the selected listening port in dev.

# Refs
Refs: goal passkey-demo-stability; requirement node-dev-ergonomics; spec node-server-config; decision logging-and-startup
