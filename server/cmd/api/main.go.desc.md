# Purpose
Entrypoint for the Go HTTP server used in the demo. It starts an `http.Server` on the configured port and exposes a `/health` endpoint that returns `200 OK` with body `ok`. Route wiring is delegated to `internal/app.BuildRouter`, which assembles handlers and applies group middlewares.

# Key Logic
- Initializes `log/slog` default logger via `internal/logging`: reads `LOG_LEVEL` and `LOG_FORMAT` (`json` default, `text` for dev), optional `DEV_ADD_SOURCE=1`.
- Loads `config.Config`, opens and migrates SQLite via `internal/storage`.
- Constructs in‑memory stores for registration, login, and tx sessions.
- Calls `app.BuildRouter(cfg, db, deps)` which returns a fully wrapped handler: `CORS → Session → group limits → mux`.
- Emits `server_start` structured log with `rp_id`, `origin`, `port`, `db_path` and starts the HTTP server; logs `server_error` on failure.

# Interactions
- Loads configuration via `internal/config.Load()` and uses `cfg.Port` for the listen address; logs `server_start` with `rp_id`, `origin`, and `db_path` at startup.
- Opens SQLite via `internal/storage.Open` + `Migrate` and passes DB handle to the router builder.
- Creates in‑memory stores for registration/login sessions and a shared `TxSessionStore` for `/tx/signing/*` (demo‑grade).
- Called directly by `go run ./cmd/api` during development; will be the main binary for the server.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-FLOW-LOGIN; spec cors-usage-explainer; decision router-builder-wiring; decision structured-logging-with-slog-guidelines; decision request-id-and-slog-json; decision http-error-envelope
