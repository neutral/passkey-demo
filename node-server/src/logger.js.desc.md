# Purpose
Provide Pino logger and HTTP logging middleware. Emit a structured `server_start` event at boot with key configuration fields.

# Key Logic
- `logger`: Pino instance (JSON logs).
- `httpLogger`: `pino-http` middleware with compact serializers (method, url, id; statusCode).
- `buildServerStartEvent(config)`: returns stable event payload.
- `logServerStart(config)`: logs `server_start`.

# Interactions
- Imported by `src/server.js` to attach HTTP logging and log boot.
- Tests validate `buildServerStartEvent` shape.

# Refs
Refs: decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines

