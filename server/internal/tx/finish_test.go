package tx

import (
    "context"
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "path/filepath"
    "testing"
    "time"

    _ "github.com/mattn/go-sqlite3"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    storage "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
)

func openFinishDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func mkCoseKey(t *testing.T, priv *ecdsa.PrivateKey) (types.CoseEC2, []byte) {
    t.Helper()
    px := priv.X.Bytes(); py := priv.Y.Bytes()
    x := make([]byte, 32); y := make([]byte, 32)
    copy(x[32-len(px):], px); copy(y[32-len(py):], py)
    k := types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: x, Y: y}
    cbor, err := enc.EncodeCanonical(k)
    if err != nil { t.Fatalf("cbor: %v", err) }
    return k, cbor
}

func insertAcctCred(t *testing.T, db *sql.DB, acctCBOR, credID []byte, signCount int64) {
    t.Helper()
    now := time.Now().Unix()
    th := sha256.Sum256(append([]byte("ACCTK1"), acctCBOR...))
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, th[:], now); err != nil { t.Fatalf("ins acct: %v", err) }
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, signCount, []byte(nil), now); err != nil { t.Fatalf("ins cred: %v", err) }
}

func insertSessionRow(t *testing.T, db *sql.DB, sid string, acctCBOR []byte, exp time.Time) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acctCBOR, exp.Unix()); err != nil { t.Fatalf("ins sess: %v", err) }
}

func TestTxFinish_Happy(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "db")}
    db := openFinishDB(t)
    defer db.Close()
    store := NewTxSessionStore(0)

    // Keys and account
    priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    k, acctCBOR := mkCoseKey(t, priv)
    credID := []byte("cred-tx-1")
    insertAcctCred(t, db, acctCBOR, credID, 7)

    // Auth session cookie
    sid := b64.Encode([]byte("sid-tx-1234567890123456789012"))
    insertSessionRow(t, db, sid, acctCBOR, time.Now().Add(time.Hour))

    // Build bundle and options (tx session)
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "hello"}
    B, _ := enc.EncodeCanonical(bun)
    inOpts := b64.Encode(B)
    // Build options to create tx session and get challenge
    opts, err := BuildTxOptions(context.Background(), cfg, store, db, acctCBOR, inOpts, time.Now)
    if err != nil { t.Fatalf("build opts: %v", err) }

    // Build CDJ/AD and sign
    cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
    cdjBytes, _ := json.Marshal(cdj)
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := make([]byte, 37)
    copy(ad[:32], rpHash[:]); ad[32] = 0x05 /* UV|UP */; ad[33] = 0; ad[34] = 0; ad[35] = 0; ad[36] = byte(8)
    hcdj := sha256.Sum256(cdjBytes)
    d := sha256.Sum256(append(ad, hcdj[:]...))
    sig, _ := ecdsa.SignASN1(rand.Reader, priv, d[:])

    // Build finish body
    body := TxFinishInbound{TxSessionID: opts.TxSessionID, ID: b64.Encode(credID), RawID: b64.Encode(credID), Type: "public-key"}
    body.Response.ClientDataJSON = b64.Encode(cdjBytes)
    body.Response.AuthenticatorData = b64.Encode(ad)
    body.Response.Signature = b64.Encode(sig)
    buf, _ := json.Marshal(body)

    // Call handler
    h := TxFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader(buf))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String()) }
}
