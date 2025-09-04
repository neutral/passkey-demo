# Centralize HTTP Route Registration in `server/cmd/api`

## Purpose
- Streamline the HTTP layer by consolidating route registration and middleware wiring in one place.

## Context
- Handlers live in packages (`internal/webauthn`, `internal/tx`) but route assembly can become scattered as features grow.

## Proposal
- Create a single router builder in `server/cmd/api`:
  - `func buildRouter(cfg *config.Config, db *sql.DB, deps ...) http.Handler` that wires all endpoints.
  - Centralize middleware (CORS, body limits, logging) and session injection once Step 24 lands.
- Keep handlers package-local and pure; router does wiring only.

## Migration Plan
1) Introduce `buildRouter` returning a `http.ServeMux` (or chi/gorilla if allowed) with current routes.
2) Move scattered registrations into that function.
3) Keep public surface minimal; CLI `main` calls `buildRouter` and `http.ListenAndServe`.

## Risks & Mitigations
- None significant; improves clarity without changing handler logic.

## Testing Strategy
- Unit: instantiate router with in-memory deps; route table snapshot (names/paths) and a few smoke requests.

## Acceptance Criteria
- One authoritative router builder; handlers unchanged; app still serves all endpoints.

## Refs
- Refs: requirement R-PLAT-2; requirement R-OPS-DEV; decision webauthn-corrections-and-standardizations

