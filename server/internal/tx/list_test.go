package tx

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "crypto/sha256"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    storage "github.com/neutral/passkey-demo/internal/storage"
)

func openListDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func insertAccountList(t *testing.T, db *sql.DB, acct []byte) {
    t.Helper()
    th := sha256.Sum256(append([]byte("ACCTK1"), acct...))
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, 0)`, acct, th[:]); err != nil {
        t.Fatalf("ins acct: %v", err)
    }
}

func insertSessionList(t *testing.T, db *sql.DB, sid string, acct []byte, exp time.Time) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acct, exp.Unix()); err != nil {
        t.Fatalf("ins sess: %v", err)
    }
}

func insertTx(t *testing.T, db *sql.DB, acct []byte, txID []byte, nonce int64, msg string, created int64) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        txID, acct, nonce, msg, []byte("B"), []byte("AD"), []byte("CD"), []byte("SIG"), created); err != nil {
        t.Fatalf("ins tx: %v", err)
    }
}

func TestTxList_HappyAndIsolation(t *testing.T) {
    db := openListDB(t)
    defer db.Close()

    // Two accounts
    acctA := []byte("acct-A")
    acctB := []byte("acct-B")
    insertAccountList(t, db, acctA)
    insertAccountList(t, db, acctB)

    // Sessions
    sidA := b64.Encode([]byte("sid-A-12345678901234567890123"))
    sidB := b64.Encode([]byte("sid-B-12345678901234567890123"))
    insertSessionList(t, db, sidA, acctA, time.Now().Add(time.Hour))
    insertSessionList(t, db, sidB, acctB, time.Now().Add(time.Hour))

    // Transactions for A (created_at descending order expected)
    insertTx(t, db, acctA, []byte("id-1"), 1, "m1", 1000)
    insertTx(t, db, acctA, []byte("id-2"), 2, "m2", 2000)
    insertTx(t, db, acctA, []byte("id-3"), 3, "m3", 1500)
    // Transaction for B
    insertTx(t, db, acctB, []byte("id-b"), 9, "mb", 3000)

    h := TxListHandler(db)

    // List for A
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/tx/list", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sidA})
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status A: %d", rr.Code) }
    var out TxListResponse
    if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil { t.Fatalf("json: %v", err) }
    if len(out.Items) != 3 { t.Fatalf("want 3 items, got %d", len(out.Items)) }
    if out.Items[0].Message != "m2" || out.Items[1].Message != "m3" || out.Items[2].Message != "m1" {
        t.Fatalf("order wrong: %+v", out.Items)
    }

    // List for B
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/tx/list", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sidB})
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status B: %d", rr.Code) }
    out = TxListResponse{}
    if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil { t.Fatalf("json: %v", err) }
    if len(out.Items) != 1 || out.Items[0].Message != "mb" {
        t.Fatalf("isolation wrong: %+v", out.Items)
    }
}

func TestTxList_Negatives(t *testing.T) {
    db := openListDB(t)
    defer db.Close()
    h := TxListHandler(db)

    // Missing cookie
    rr := httptest.NewRecorder(); req := httptest.NewRequest("GET", "/tx/list", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("missing cookie code=%d", rr.Code) }

    // Unknown session
    rr = httptest.NewRecorder(); req = httptest.NewRequest("GET", "/tx/list", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: b64.Encode([]byte("nope"))})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("unknown session code=%d", rr.Code) }

    // Expired session
    acct := []byte("acct-X")
    insertAccountList(t, db, acct)
    sid := b64.Encode([]byte("sid-X-12345678901234567890123"))
    insertSessionList(t, db, sid, acct, time.Now().Add(-time.Minute))
    rr = httptest.NewRecorder(); req = httptest.NewRequest("GET", "/tx/list", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("expired session code=%d", rr.Code) }
}
