# Overview
Central router builder that wires the HTTP server routes and applies group-level middlewares. It centralizes all `ServeMux` registrations for the backend and enforces consistent layering for cross-cutting concerns.

# Relations
- Outermost wrapper is `internal/httpx/middleware.RequestID`, then `internal/http.CORSMiddleware`, then `internal/http.SessionMiddleware`.
- Group middlewares (`BodyLimitMiddleware`, `RateLimitMiddleware`) applied around `/authn/*` and `/tx/*` handlers.
- Depends on `internal/config` for runtime settings; injects in-memory session stores and prepared-statement repositories through `Deps`.
- Uses handlers from `internal/webauthn`, `internal/tx`, and `internal/me`.

# Interfaces & Models
- `func BuildRouter(cfg *config.Config, db *sql.DB, deps Deps) http.Handler`
- `type Deps` contains typed in-memory stores and repos for registration, login, and tx signing flows.

# Refs
Refs: goal simple-ui-and-storage; requirement R-ERR; decision router-builder-wiring
