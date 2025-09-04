package tx

import (
    "context"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "time"
    httpctx "github.com/neutral/passkey-demo/internal/http"
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
    return buildTxListByAcct(ctx, db, acctCBOR)
}

func buildTxListByAcct(ctx context.Context, db *sql.DB, acctCBOR []byte) (TxListResponse, error) {
    rows, err := db.QueryContext(ctx, `SELECT tx_id, nonce, message, created_at FROM transactions WHERE acct_cbor = ? ORDER BY created_at DESC`, acctCBOR)
    if err != nil {
        return TxListResponse{}, err
    }
    defer rows.Close()
    var out TxListResponse
    for rows.Next() {
        var txID []byte
        var nonce int64
        var message string
        var created int64
        if err := rows.Scan(&txID, &nonce, &message, &created); err != nil {
            return TxListResponse{}, err
        }
        out.Items = append(out.Items, TxListItem{
            TxIDHex:   hex.EncodeToString(txID),
            Nonce:     uint64(nonce),
            Message:   message,
            CreatedAt: created,
        })
    }
    if err := rows.Err(); err != nil {
        return TxListResponse{}, err
    }
    return out, nil
}

// TxListHandler serves GET /tx/list.
func TxListHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        // Prefer session from middleware; otherwise fallback to cookie lookup.
        if s, ok := httpctx.FromSession(r.Context()); ok {
            resp, err := buildTxListByAcct(r.Context(), db, s.AcctCBOR)
            if err != nil {
                http.Error(w, "internal error", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "application/json")
            _ = json.NewEncoder(w).Encode(resp)
            return
        }
        c, err := r.Cookie("sid")
        if err != nil || c.Value == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        resp, err := BuildTxList(r.Context(), db, c.Value)
        if err != nil {
            if errors.Is(err, ErrAuthSession) {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(resp)
    }
}
