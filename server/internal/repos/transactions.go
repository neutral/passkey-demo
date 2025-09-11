package repos

import (
    "context"
    "database/sql"
)

// TransactionsRepo centralizes transaction read queries.
type TransactionsRepo struct {
    db           *sql.DB
    listByAcct   *sql.Stmt
}

func NewTransactions(ctx context.Context, db *sql.DB) (*TransactionsRepo, error) {
    stmt, err := db.PrepareContext(ctx, `SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC`)
    if err != nil { return nil, err }
    return &TransactionsRepo{db: db, listByAcct: stmt}, nil
}

func (r *TransactionsRepo) Close() error { return r.listByAcct.Close() }

type TxRow struct {
    TxID     []byte
    Nonce    int64
    Message  string
    Created  int64
}

// ListByAccount returns transactions ordered by created_at DESC.
func (r *TransactionsRepo) ListByAccount(ctx context.Context, acctCBOR []byte) ([]TxRow, error) {
    rows, err := r.listByAcct.QueryContext(ctx, acctCBOR)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []TxRow
    for rows.Next() {
        var row TxRow
        if err := rows.Scan(&row.TxID, &row.Nonce, &row.Message, &row.Created); err != nil { return nil, err }
        out = append(out, row)
    }
    return out, rows.Err()
}

