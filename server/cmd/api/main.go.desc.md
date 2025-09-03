# Purpose
Entrypoint for the Go HTTP server used in the demo. It starts an `http.Server` on the configured port and exposes a `/health` endpoint that returns `200 OK` with body `ok`. It wires WebAuthn registration and login endpoints.

# Key Logic
- Reads `PORT` from env (default `8080`).
- Uses `http.NewServeMux()` and registers:
  - `POST /authn/passkey/registration/options`
  - `POST /authn/passkey/registration/finish`
  - `POST /authn/passkey/login/options`
  - `POST /authn/passkey/login/finish`
- Logs a startup line and calls `http.ListenAndServe`.

# Interactions
- Loads configuration via `internal/config.Load()` and uses `cfg.Port` for the listen address; logs rp_id, origin, and db path at startup.
- Opens SQLite via `internal/storage.Open` + `Migrate` and passes DB handle into handlers.
- Creates in‑memory stores for registration/login sessions (demo‑grade).
- Called directly by `go run ./cmd/api` during development; will be the main binary for the server.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV; requirement R-FLOW-LOGIN
