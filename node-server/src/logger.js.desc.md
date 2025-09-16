# Purpose
Provide the Node server with a centralized Pino logger, HTTP logging middleware, and helper functions that emit structured JSON events with correlation ids for every major ceremony and verification failure.

# Key Logic
- `logger`: Pino instance configured from `LOG_LEVEL` (default `info`).
- `httpLogger`: `pino-http` middleware emitting `{ method, url, id }` and status code per request.
- Helper primitives `logInfo`/`logError` (internal) enrich payloads with `correlation_id`, `rp_id`, `origin`; exported wrappers cover:
  - `logServerStart`, `logRegOptionsSuccess/Error`, `logRegFinishSuccess/Error`, `logLoginOptionsSuccess/Error`, `logLoginFinishSuccess/Error`.
  - Transaction helpers (`logTxOptions*`, `logTxFinish*`, `logTxList*`) and account key helpers (`logMeAccountKey*`).
  - `logWebauthnVerifyFailure` emits the `webauthn_assert_verify` event with `error_kind`, hashed identifiers, and policy flags.

# Interactions
- `server.js` uses `logServerStart` during boot; all route handlers call the helpers with hashed identifiers, nonce/tx metadata, and correlation ids captured by middleware.
- Tests stub `logger.info/error` to assert payloads without writing to stdout.

# Refs
Refs: requirement R-ERR; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines
