# Light DB Abstraction & Dependency Injection for Testability

## Purpose
- Improve testability and reduce coupling by abstracting minimal DB operations and injecting time/random/session deps.

## Context
- Handlers call `*sql.DB` directly with string queries; tests create in-memory SQLite and call functions.
- Some logic accepts `now func() time.Time` (good); others could benefit from similar injection.

## Proposal
- Define minimal interfaces for DB usage in hot paths:
  - `type DB interface { QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row; QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error); ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) }` (or wrap commonly used methods).
- Pattern: accept `now func() time.Time`, `rand func(n int)([]byte,error)`, and stores as interfaces in builders.
- Optional: small repository structs for grouped queries (sessions, credentials, transactions) to centralize SQL and prepared statements.

## Migration Plan
1) Introduce interfaces/types alongside existing `*sql.DB` usage.
2) Update builders first (e.g., `BuildLoginOptions`, `BuildTxOptions`) to accept interfaces.
3) Gradually adopt in handlers; keep constructors to wire real `*sql.DB`.

## Risks & Mitigations
- Over-abstraction: keep interface surface minimal and close to current usage.

## Testing Strategy
- Unit: use fakes/mocks implementing the minimal DB interface for edge cases (errors, no rows).
- Integration: preserve current sqlite-backed tests for end-to-end sanity.

## Acceptance Criteria
- Core paths accept interfaces and injected clocks/rand; tests can simulate failures without real DB.

## Refs
- Refs: requirement R-PLAT-2; requirement R-PLAT-3; requirement R-OPS-DEV

