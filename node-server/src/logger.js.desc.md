# Purpose
Provide Pino logger, HTTP logging middleware, and helper functions for structured events such as `server_start` and transaction option flows.

# Key Logic
- `logger`: Pino instance (JSON logs).
- `httpLogger`: `pino-http` middleware with compact serializers (method, url, id; statusCode).
- `buildServerStartEvent` / `logServerStart`: emit boot payload (rp_id, origin, port, db path).
- `logTxOptionsSuccess` / `logTxOptionsError`: wrap `logger.info`/`logger.error` with consistent payloads for `tx_options` and `tx_options_error` events.

# Interactions
- Imported by `server.js` to attach HTTP logging and bootstrap logs.
- Transaction routes call `logTxOptionsSuccess/Error` to record anchor metadata and correlation ids.

# Refs
Refs: decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines; requirement R-FLOW-SIGN
