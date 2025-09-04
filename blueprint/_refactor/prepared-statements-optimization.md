# Optimize Prepared Statements & Reduce Duplication

## Purpose
- Reduce repeated SQL string usage and improve performance by preparing frequently used statements once and reusing them.

## Context
- Queries like session lookups, credential fetches, and transaction inserts appear across handlers/builders.

## Proposal
- Introduce small repos (e.g., `SessionsRepo`, `CredentialsRepo`, `TransactionsRepo`) holding prepared statements:
  - `Prepare(db *sql.DB) (*Repo, error)`; fields like `ByID *sql.Stmt`, `Insert *sql.Stmt`, etc.
  - Lifecycle: prepare on startup; close on shutdown.
- Replace inline queries in hot paths with repo calls.

## Migration Plan
1) Identify hottest queries (lookup session, list credential IDs, tx insert).
2) Prepare statements in app startup and pass repos to handlers/builders.
3) Replace inline queries; keep SQL text centralized.

## Risks & Mitigations
- Statement management complexity: encapsulate in repos with `Close()`; add tests for prepare/close.

## Testing Strategy
- Unit: repo prepare/close happy path and error path; methods execute against in-memory SQLite.
- Integration: existing handler tests remain green.

## Acceptance Criteria
- Common queries use prepared statements via repos; duplication reduced; tests pass.

## Refs
- Refs: requirement R-PLAT-2; requirement R-PLAT-3; requirement R-OPS-DEV

