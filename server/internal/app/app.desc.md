# Overview
Top-level `app` package groups application assembly concerns such as the router builder and, in the future, use-case orchestration services. It is intentionally thin and wires together other packages.

# Relations
- Imports handler packages (`webauthn`, `tx`, `me`) and cross-cutting middleware from `internal/http`.
- Owned by the `server/cmd/api` entrypoint, which constructs dependencies and calls `BuildRouter`.

# Interfaces & Models
- `BuildRouter(cfg, db, deps)` returns the fully-wrapped `http.Handler` for the server.
- `Deps` aggregates runtime stores and, later, repository/service instances.

# Refs
Refs: goal simple-ui-and-storage; requirement R-ERR; decision router-builder-wiring

