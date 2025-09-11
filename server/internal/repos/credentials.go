package repos

import (
    "context"
    "database/sql"
)

// CredentialsRepo centralizes credential lookups.
type CredentialsRepo struct {
    db           *sql.DB
    listByAcct   *sql.Stmt
    getByID      *sql.Stmt
}

// NewCredentials prepares statements for read paths.
func NewCredentials(ctx context.Context, db *sql.DB) (*CredentialsRepo, error) {
    listByAcct, err := db.PrepareContext(ctx, `SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?`)
    if err != nil { return nil, err }
    getByID, err := db.PrepareContext(ctx, `SELECT acct_cbor_fk, sign_count FROM credentials WHERE credential_id = ?`)
    if err != nil { listByAcct.Close(); return nil, err }
    return &CredentialsRepo{db: db, listByAcct: listByAcct, getByID: getByID}, nil
}

func (r *CredentialsRepo) Close() error {
    _ = r.listByAcct.Close()
    _ = r.getByID.Close()
    return nil
}

// ListIDsByAccount returns all credential IDs for the given account CBOR.
func (r *CredentialsRepo) ListIDsByAccount(ctx context.Context, acctCBOR []byte) ([][]byte, error) {
    rows, err := r.listByAcct.QueryContext(ctx, acctCBOR)
    if err != nil { return nil, err }
    defer rows.Close()
    var out [][]byte
    for rows.Next() {
        var id []byte
        if err := rows.Scan(&id); err != nil { return nil, err }
        out = append(out, append([]byte(nil), id...))
    }
    return out, rows.Err()
}

// GetAccountAndCount returns the owning account and sign_count for a credential id.
func (r *CredentialsRepo) GetAccountAndCount(ctx context.Context, credID []byte) (acctCBOR []byte, signCount int64, err error) {
    err = r.getByID.QueryRowContext(ctx, credID).Scan(&acctCBOR, &signCount)
    return
}

