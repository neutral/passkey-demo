# Purpose
Entrypoint for the Go HTTP server used in the demo. It starts an `http.Server` on the configured port and exposes a `/health` endpoint that returns `200 OK` with body `ok`.

# Key Logic
- Reads `PORT` from env (default `8080`).
- Uses `http.NewServeMux()` and registers a simple health handler.
- Logs a startup line and calls `http.ListenAndServe`.

# Interactions
- Loads configuration via `internal/config.Load()` and uses `cfg.Port` for the listen address; logs rp_id, origin, and db path at startup.
- No DB, CORS, cookies, or middleware yet. Those are introduced in later steps (DB, sessions, CORS).
- Called directly by `go run ./cmd/api` during development; will be the main binary for the server.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV
