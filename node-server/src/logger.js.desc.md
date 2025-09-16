# Purpose
Provide Pino logger, HTTP logging middleware, and helper functions for structured events such as `server_start` and transaction flows.

# Key Logic
- `logger`: Pino instance (JSON logs).
- `httpLogger`: `pino-http` middleware with compact serializers (method, url, id; statusCode).
- `buildServerStartEvent` / `logServerStart`: emit boot payload (rp_id, origin, port, db path).
- `logTxOptionsSuccess` / `logTxOptionsError`: wrap `logger.info`/`logger.error` for `tx_options` events.
- `logTxFinishSuccess` / `logTxFinishError`: structured logging for `/tx/signing/finish` outcomes (include hashes, counters, reasons).

# Interactions
- Imported by `server.js` to attach HTTP logging and bootstrap logs.
- Transaction routes call the tx log helpers to capture correlation ids, anchors, nonce/sign-count metrics, and error reasons.

# Refs
Refs: decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines; requirement R-FLOW-SIGN
