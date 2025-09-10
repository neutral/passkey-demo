package http

import (
    "database/sql"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "crypto/sha256"
    storage "github.com/neutral/passkey-demo/internal/storage"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
)

func openDBRefresh(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func insertAcctRefresh(t *testing.T, db *sql.DB, acct []byte) {
    t.Helper()
    th := sha256.Sum256(append([]byte("ACCTK1"), acct...))
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, 0)`, acct, th[:]); err != nil {
        t.Fatalf("ins acct: %v", err)
    }
}

func insertSessRefresh(t *testing.T, db *sql.DB, sid string, acct []byte, exp time.Time) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acct, exp.Unix()); err != nil {
        t.Fatalf("ins sess: %v", err)
    }
}

func TestSessionMiddleware_RefreshExtendsExpiry(t *testing.T) {
    db := openDBRefresh(t); defer db.Close()
    acct := []byte("acct-refresh")
    insertAcctRefresh(t, db, acct)
    sid := b64.Encode([]byte("sid-refresh-123456789012345678901"))
    // Set expiry to 10s from now, TTL=1h so less than half TTL remains
    exp := time.Now().Add(10 * time.Second)
    insertSessRefresh(t, db, sid, acct, exp)

    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    mw := SessionMiddleware(db, true, time.Hour)
    h := mw(next)

    rr := httptest.NewRecorder(); req := httptest.NewRequest("GET", "/", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("unexpected status: %d", rr.Code) }

    // Expiry should be extended roughly by TTL; assert it moved forward significantly
    var expAfter int64
    if err := db.QueryRow(`SELECT expires_at FROM sessions WHERE session_id = ?`, sid).Scan(&expAfter); err != nil {
        t.Fatalf("query exp: %v", err)
    }
    if expAfter <= exp.Unix()+int64(30*time.Minute/time.Second) {
        t.Fatalf("expected expiry extended, got %d <= baseline", expAfter)
    }
}

