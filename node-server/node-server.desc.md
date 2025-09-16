# Overview
Express-based Node backend scaffold for the Passkey Demo. Provides bootstrap, environment configuration, JSON logging (Pino), request-id middleware, SQLite handle open (PRAGMAs only), and a basic `/health` endpoint. Subsequent steps layer CORS, sessions, WebAuthn, and schema logic.

# Relations
- Primary API consumed by the web app during local dev and test. Starts on `PORT` (default 8080) and logs `server_start` with config fields.
- Depends on: `express`, `pino(+pino-http)`, `dotenv`, `better-sqlite3`, `cbor-x`. Future steps will wire `cors`, `cookie`, and `@simplewebauthn/server`.

# Interfaces & Models
- Endpoint: `GET /health` → `{ status: 'ok' }`.
- Event: `server_start` (JSON log) with `{ rp_id, origin, port, db_path }`.

# Refs
Refs: requirement R-PLAT-3; requirement R-OPS-DEV; decision router-builder-wiring; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines
