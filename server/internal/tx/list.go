package tx

import (
    "context"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "time"
    httpctx "github.com/neutral/passkey-demo/internal/http"
    repos "github.com/neutral/passkey-demo/internal/repos"
)

// TxListItem is a single transaction list entry in the response.
type TxListItem struct {
    TxIDHex   string `json:"tx_id_hex"`
    Nonce     uint64 `json:"nonce"`
    Message   string `json:"message"`
    CreatedAt int64  `json:"created_at"`
}

// TxListResponse is the JSON response envelope for /tx/list.
type TxListResponse struct {
    Items []TxListItem `json:"items"`
}

// BuildTxList resolves the account by session cookie and returns that account's transactions.
func BuildTxList(ctx context.Context, db *sql.DB, sid string) (TxListResponse, error) {
    // Resolve auth session
    var acctCBOR []byte
    var exp int64
    if err := db.QueryRowContext(ctx, `SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`, sid).Scan(&acctCBOR, &exp); err != nil {
        return TxListResponse{}, ErrAuthSession
    }
    if time.Now().Unix() >= exp {
        return TxListResponse{}, ErrAuthSession
    }
    // Query transactions for this account (newest first)
    // Build using direct query (legacy path used only in tests)
    rows, err := db.QueryContext(ctx, `SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC`, acctCBOR)
    if err != nil { return TxListResponse{}, err }
    defer rows.Close()
    var out TxListResponse
    for rows.Next() {
        var txID []byte
        var nonce int64
        var message string
        var created int64
        if err := rows.Scan(&txID, &nonce, &message, &created); err != nil { return TxListResponse{}, err }
        out.Items = append(out.Items, TxListItem{TxIDHex: hex.EncodeToString(txID), Nonce: uint64(nonce), Message: message, CreatedAt: created})
    }
    return out, rows.Err()
}

func buildTxListByAcct(ctx context.Context, repo *repos.TransactionsRepo, acctCBOR []byte) (TxListResponse, error) {
    rows, err := repo.ListByAccount(ctx, acctCBOR)
    if err != nil { return TxListResponse{}, err }
    var out TxListResponse
    for _, r := range rows {
        out.Items = append(out.Items, TxListItem{
            TxIDHex:   hex.EncodeToString(r.TxID),
            Nonce:     uint64(r.Nonce),
            Message:   r.Message,
            CreatedAt: r.Created,
        })
    }
    return out, nil
}

// TxListHandler serves GET /tx/list. Requires session middleware.
func TxListHandler(repo *repos.TransactionsRepo) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        s, ok := httpctx.FromSession(r.Context())
        if !ok {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        resp, err := buildTxListByAcct(r.Context(), repo, s.AcctCBOR)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}
