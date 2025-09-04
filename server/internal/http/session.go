package http

import (
    "context"
    "database/sql"
    "net/http"
    "time"
)

// SessionContext holds authenticated account identity and related data
// derived from the server session.
type SessionContext struct {
    AcctCBOR      []byte
    CredentialIDs [][]byte
}

type sessionKey struct{}

// WithSession attaches a SessionContext to ctx.
func WithSession(ctx context.Context, s SessionContext) context.Context {
    return context.WithValue(ctx, sessionKey{}, s)
}

// FromSession extracts a SessionContext from ctx.
func FromSession(ctx context.Context) (SessionContext, bool) {
    v := ctx.Value(sessionKey{})
    if v == nil {
        return SessionContext{}, false
    }
    s, ok := v.(SessionContext)
    return s, ok
}

// SessionMiddleware returns a middleware that reads the `sid` cookie, looks up the
// session in the database, optionally refreshes its expiry, loads credential ids
// for the account, and attaches a SessionContext to the request context.
//
// If the cookie is missing or invalid/expired, the middleware passes the request
// through without attaching context; downstream protected handlers should return 401.
func SessionMiddleware(db *sql.DB, refresh bool, ttl time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c, err := r.Cookie("sid")
            if err != nil || c.Value == "" {
                next.ServeHTTP(w, r)
                return
            }
            // Lookup session
            var acctCBOR []byte
            var exp int64
            row := db.QueryRow(`SELECT acct_cbor, expires_at FROM sessions WHERE session_id = ?`, c.Value)
            if err := row.Scan(&acctCBOR, &exp); err != nil {
                next.ServeHTTP(w, r)
                return
            }
            now := time.Now().Unix()
            if now >= exp {
                next.ServeHTTP(w, r)
                return
            }
            // Optionally refresh expiry if half TTL elapsed
            if refresh {
                // Use a simple heuristic: if less than half TTL remains, extend.
                if exp-now < int64(ttl/time.Second)/2 {
                    newExp := time.Now().Add(ttl).Unix()
                    _, _ = db.Exec(`UPDATE sessions SET expires_at=? WHERE session_id=?`, newExp, c.Value)
                    exp = newExp
                }
            }
            // Load credential ids for convenience (allowlists, etc.)
            rows, err := db.Query(`SELECT credential_id FROM credentials WHERE acct_cbor_fk = ?`, acctCBOR)
            var creds [][]byte
            if err == nil {
                defer rows.Close()
                for rows.Next() {
                    var id []byte
                    if err := rows.Scan(&id); err == nil {
                        creds = append(creds, append([]byte(nil), id...))
                    }
                }
            }
            s := SessionContext{AcctCBOR: append([]byte(nil), acctCBOR...), CredentialIDs: creds}
            r2 := r.WithContext(WithSession(r.Context(), s))
            next.ServeHTTP(w, r2)
        })
    }
}

