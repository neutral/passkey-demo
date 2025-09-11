# Overview
Prepared-statement repository for transaction queries. Avoids duplicating SQL in handlers when listing transactions for an account.

# Relations
- Used by `/tx/list` handler to read transactions by `acct_cbor`.

# Interfaces & Models
- `NewTransactions(ctx, db)` prepares the query.
- `ListByAccount(ctx, acctCBOR) ([]TxRow, error)` returns rows ordered by `created_at DESC`.

# Refs
Refs: goal simple-ui-and-storage; requirement sqlite-persistence; decision data-access-repos-prepared

