package http

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "crypto/sha256"
    storage "github.com/neutral/passkey-demo/internal/storage"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
)

func openDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func insertAcct(t *testing.T, db *sql.DB, acct []byte) {
    t.Helper()
    th := sha256.Sum256(append([]byte("ACCTK1"), acct...))
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, 0)`, acct, th[:]); err != nil {
        t.Fatalf("ins acct: %v", err)
    }
}

func insertCred(t *testing.T, db *sql.DB, acct []byte, credID []byte) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, 0, ?, 0)`, credID, acct, []byte(nil)); err != nil {
        t.Fatalf("ins cred: %v", err)
    }
}

func insertSess(t *testing.T, db *sql.DB, sid string, acct []byte, exp time.Time) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acct, exp.Unix()); err != nil {
        t.Fatalf("ins sess: %v", err)
    }
}

func TestSessionMiddleware_Happy(t *testing.T) {
    db := openDB(t); defer db.Close()
    acct := []byte("acct-test")
    insertAcct(t, db, acct)
    cred := []byte("cred-1")
    insertCred(t, db, acct, cred)
    sid := b64.Encode([]byte("sid-ctx-12345678901234567890123"))
    insertSess(t, db, sid, acct, time.Now().Add(time.Hour))

    // Next handler checks context
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        s, ok := FromSession(r.Context())
        if !ok || string(s.AcctCBOR) != string(acct) || len(s.CredentialIDs) != 1 {
            http.Error(w, "no session", http.StatusUnauthorized)
            return
        }
        w.WriteHeader(http.StatusOK)
    })

    mw := SessionMiddleware(db, true, time.Hour)
    h := mw(next)

    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("code=%d", rr.Code) }
}

func TestSessionMiddleware_MissingOrExpired(t *testing.T) {
    db := openDB(t); defer db.Close()
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if _, ok := FromSession(r.Context()); ok {
            w.WriteHeader(http.StatusOK); return
        }
        http.Error(w, "no session", http.StatusUnauthorized)
    })
    mw := SessionMiddleware(db, false, time.Hour)
    h := mw(next)

    // Missing cookie
    rr := httptest.NewRecorder(); req := httptest.NewRequest("GET", "/", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("missing cookie code=%d", rr.Code) }

    // Expired session
    acct := []byte("acct-x")
    insertAcct(t, db, acct)
    sid := b64.Encode([]byte("sid-x-12345678901234567890123"))
    insertSess(t, db, sid, acct, time.Now().Add(-time.Minute))
    rr = httptest.NewRecorder(); req = httptest.NewRequest("GET", "/", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("expired cookie code=%d", rr.Code) }
}

