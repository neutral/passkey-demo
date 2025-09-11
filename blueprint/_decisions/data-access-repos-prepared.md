Title: Prepared Statement Repositories for Read Paths
Status: Accepted (Initial Read Paths)
Date: 2025-09-11

Context:
- Inline SQL strings were duplicated across handlers; prepared statement reuse and centralization improve clarity and performance.

Decision:
1) Add small repositories with prepared statements, prepared on startup and injected:
   - `CredentialsRepo`: list ids by account; get account/sign_count by credential id.
   - `TransactionsRepo`: list transactions by account (created_at DESC).
2) Use repos in tx options, tx finish (read), and tx list. Keep write paths direct for now.

Consequences:
- Less duplication; better lifecycle control for prepared statements; clearer seams for testing.

Alternatives:
- Keep all inline SQL: harder to maintain and optimize.

References:
- Implementation: `server/internal/repos/*`; wiring in `server/internal/app/router.go` and `server/cmd/api/main.go`.

Refs: requirement sqlite-persistence; goal simple-ui-and-storage

