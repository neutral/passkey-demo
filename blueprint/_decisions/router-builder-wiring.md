Title: Central Router Builder and Group Middleware Wiring
Status: Accepted
Date: 2025-09-11

Context:
- Route registration and middleware layering were scattered; handlers sometimes mixed transport and auth concerns.
- We need a single place to apply CORS, session auth, rate/body limits, and future observability consistently.

Decision:
1) Introduce `internal/app.BuildRouter(cfg, db, deps) http.Handler` to own route wiring.
2) Layering order: `RequestID` (outermost) → `CORS` → `SessionMiddleware` → per-group middlewares → handlers.
3) Group middlewares:
   - `/authn/*`: body limit 1 MiB; token-bucket rate limit (burst 20, ~10 rps).
   - `/tx/*`: body limit 1 MiB; token-bucket rate limit (burst 20, ~10 rps).
4) Protected endpoints do not read cookies/DB directly; they rely solely on `SessionMiddleware` to inject auth context.
5) Keep `main` minimal: construct deps (stores, repos) and call `BuildRouter`.

Consequences:
- Consistent cross-cutting behavior, easier audits; fewer transport concerns leaking into handlers.
- Enables incremental adoption of new middlewares without changing handlers.

Alternatives:
- Leave wiring in `cmd/api/main.go` and per-package init: rejected; harder to reason about, higher drift risk.

References:
- Implementation: `server/internal/app/router.go`

Refs: requirement R-PLAT-2; requirement R-ERR; goal simple-ui-and-storage

