package tx

import (
    "bytes"
    "crypto/elliptic"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    _ "github.com/mattn/go-sqlite3"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    "context"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    httpctx "github.com/neutral/passkey-demo/internal/http"
    storage "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
    repos "github.com/neutral/passkey-demo/internal/repos"
)

func testCfg() *cfgpkg.Config {
    return &cfgpkg.Config{RP_ID: "example.com", Origin: "https://example.com"}
}

func openTxTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil {
        t.Fatalf("migrate: %v", err)
    }
    return db
}

// mkCoseEC2 returns a deterministic COSE EC2 public key using the curve's base point.
// The impl uses the internal encoding package; kept separate for clarity in tests.
func b64cborEncodeImpl(v any) ([]byte, error) { return b64.EncodeCanonical(v) }

func insertAccount(t *testing.T, db *sql.DB, acctCBOR []byte) {
    t.Helper()
    if _, err := db.Exec(`INSERT OR IGNORE INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, 0)`, acctCBOR, []byte("thumb")); err != nil {
        t.Fatalf("insert account: %v", err)
    }
}

func insertCredential(t *testing.T, db *sql.DB, acctCBOR, credID []byte) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, 0, ?, 0)`, credID, acctCBOR, []byte("aaguid-16-bytes!!")); err != nil {
        t.Fatalf("insert credential: %v", err)
    }
}

func insertSession(t *testing.T, db *sql.DB, sid string, acctCBOR []byte, expiresAt int64) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acctCBOR, expiresAt); err != nil {
        t.Fatalf("insert session: %v", err)
    }
}

// encodeCanonical is a local alias to internal encoding for clarity.
func encodeCanonical(t *testing.T, v any) []byte {
    t.Helper()
    B, err := b64cborEncodeImpl(v)
    if err != nil { t.Fatalf("cbor encode: %v", err) }
    return B
}

func TestTxOptionsHandler_Happy(t *testing.T) {
    if testing.Short() { t.Skip("skipping tx options tests in -short mode") }
    db := openTxTestDB(t)
    defer db.Close()
    store := NewTxSessionStore(0)
    cfg := testCfg()

    // Account + credential
    // Build COSE key
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32)
    py := make([]byte, 32)
    copy(px[32-len(gx):], gx)
    copy(py[32-len(gy):], gy)
    k := types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: px, Y: py}
    acctCBOR, err := b64cborEncodeImpl(k)
    if err != nil { t.Fatalf("acct cbor: %v", err) }
    insertAccount(t, db, acctCBOR)
    credID := []byte("cred-1")
    insertCredential(t, db, acctCBOR, credID)

    // Session cookie
    sid := b64.Encode([]byte("sid-123456789012345678901234"))
    insertSession(t, db, sid, acctCBOR, time.Now().Add(1*time.Hour).Unix())

    // Bundle
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "hello"}
    in := TxOptionsInbound{BundleCBOR: b64.Encode(encodeCanonical(t, bun))}
    body, _ := json.Marshal(in)

    // Request
    credsRepo, err := repos.NewCredentials(context.Background(), db)
    if err != nil { t.Fatalf("repo: %v", err) }
    h := httpctx.SessionMiddleware(db, true, time.Hour)(TxOptionsHandler(cfg, store, credsRepo, db))
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader(body))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status: %d", rr.Code) }
    var out TxOptionsResponse
    if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil { t.Fatalf("json: %v", err) }
    if out.TxSessionID == "" || out.Challenge == "" || out.Options.RP_ID == "" || out.Options.Origin == "" || out.ExpiresAt == 0 || out.TxIDHex == "" {
        t.Fatalf("missing fields: %+v", out)
    }
    if !out.Options.UVRequired { t.Fatalf("uv_required expected true") }
    // allowCredentials contains our credential id
    found := false
    for _, s := range out.Options.AllowCredentials {
        b, _ := b64.Decode(s)
        if bytes.Equal(b, credID) { found = true; break }
    }
    if !found { t.Fatalf("credential id not present in allowCredentials") }
    // Store entry exists
    sess, ok := store.Get(out.TxSessionID)
    if !ok { t.Fatalf("tx store missing session %s", out.TxSessionID) }
    ch, _ := b64.Decode(out.Challenge)
    if string(ch) != string(sess.Challenge) || string(sess.AcctCBOR) != string(acctCBOR) || len(sess.B) == 0 {
        t.Fatalf("stored session mismatch") }
}

func TestTxOptionsHandler_Negatives(t *testing.T) {
    if testing.Short() { t.Skip("skipping tx options tests in -short mode") }
    db := openTxTestDB(t)
    defer db.Close()
    store := NewTxSessionStore(0)
    cfg := testCfg()

    // Account and session setup
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32)
    py := make([]byte, 32)
    copy(px[32-len(gx):], gx)
    copy(py[32-len(gy):], gy)
    k := types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: px, Y: py}
    acctCBOR, _ := b64cborEncodeImpl(k)
    insertAccount(t, db, acctCBOR)
    credID := []byte("cred-1")
    insertCredential(t, db, acctCBOR, credID)
    sid := b64.Encode([]byte("sid-123456789012345678901234"))
    insertSession(t, db, sid, acctCBOR, time.Now().Add(1*time.Hour).Unix())

    credsRepo, err := repos.NewCredentials(context.Background(), db)
    if err != nil { t.Fatalf("repo: %v", err) }
    h := httpctx.SessionMiddleware(db, true, time.Hour)(TxOptionsHandler(cfg, store, credsRepo, db))

    // 1) Missing cookie → 401
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader([]byte(`{"bundle_cbor_b64":"AA"}`)))
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("missing session status=%d", rr.Code) }
    var env struct{ Code string `json:"code"` }
    _ = json.Unmarshal(rr.Body.Bytes(), &env)
    if env.Code == "" { t.Fatalf("expected error envelope code") }

    // 2) Expired session → 401
    sid2 := b64.Encode([]byte("sid-2-12345678901234567890123"))
    insertSession(t, db, sid2, acctCBOR, time.Now().Add(-1*time.Minute).Unix())
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader([]byte(`{"bundle_cbor_b64":"AA"}`)))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid2})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("expired session status=%d", rr.Code) }
    _ = json.Unmarshal(rr.Body.Bytes(), &env)
    if env.Code == "" { t.Fatalf("expected error envelope code") }

    // Build a valid bundle for further tests
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "m"}
    validBody, _ := json.Marshal(TxOptionsInbound{BundleCBOR: b64.Encode(encodeCanonical(t, bun))})

    // 3) Bad base64 → 400
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader([]byte(`{"bundle_cbor_b64":"@@bad@@"}`)))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 400 { t.Fatalf("bad base64 status=%d", rr.Code) }
    _ = json.Unmarshal(rr.Body.Bytes(), &env)
    if env.Code == "" { t.Fatalf("expected error envelope code") }

    // 4) Sender key mismatch → 401
    k2 := k
    k2.X = append([]byte{}, k.X...)
    k2.X[31] ^= 0x01
    bunBad := types.Bundle{SenderKey: k2, Nonce: 2, Message: "m"}
    bodyBad, _ := json.Marshal(TxOptionsInbound{BundleCBOR: b64.Encode(encodeCanonical(t, bunBad))})
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader(bodyBad))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("sender mismatch status=%d", rr.Code) }
    _ = json.Unmarshal(rr.Body.Bytes(), &env)
    if env.Code == "" { t.Fatalf("expected error envelope code") }

    // 5) Nonce not monotonic → 409
    // Preload a transaction with nonce=10
    _, err = db.Exec(`INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, 10, 'x', 'B', 'AD', 'CD', 'SIG', 0)`, []byte("id"), acctCBOR)
    if err != nil { t.Fatalf("insert tx: %v", err) }
    bunLow := types.Bundle{SenderKey: k, Nonce: 10, Message: "m"}
    bodyLow, _ := json.Marshal(TxOptionsInbound{BundleCBOR: b64.Encode(encodeCanonical(t, bunLow))})
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader(bodyLow))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 409 { t.Fatalf("nonce conflict status=%d", rr.Code) }
    _ = json.Unmarshal(rr.Body.Bytes(), &env)
    if env.Code == "" { t.Fatalf("expected error envelope code") }

    // 6) No credentials for account → 409
    // Create a new account without credentials
    k3 := k
    k3.X = append([]byte{}, k.X...)
    k3.X[31] ^= 0x02
    acct2, _ := b64cborEncodeImpl(k3)
    insertAccount(t, db, acct2)
    sid3 := b64.Encode([]byte("sid-3-12345678901234567890123"))
    insertSession(t, db, sid3, acct2, time.Now().Add(1*time.Hour).Unix())
    bun2 := types.Bundle{SenderKey: k3, Nonce: 1, Message: "m"}
    body2, _ := json.Marshal(TxOptionsInbound{BundleCBOR: b64.Encode(encodeCanonical(t, bun2))})
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/tx/signing/options", bytes.NewReader(body2))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid3})
    h.ServeHTTP(rr, req)
    if rr.Code != 409 { t.Fatalf("no credentials status=%d", rr.Code) }
    _ = json.Unmarshal(rr.Body.Bytes(), &env)
    if env.Code == "" { t.Fatalf("expected error envelope code") }

    _ = validBody
}
