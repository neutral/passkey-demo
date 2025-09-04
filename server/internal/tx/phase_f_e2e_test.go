package tx

import (
    "encoding/asn1"
    "bytes"
    "context"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "math/big"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    _ "github.com/mattn/go-sqlite3"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    storage "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
)

func openPhaseDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func mkAcct(t *testing.T) (*ecdsa.PrivateKey, types.CoseEC2, []byte) {
    t.Helper()
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("keygen: %v", err) }
    x := priv.X.Bytes(); y := priv.Y.Bytes()
    px := make([]byte, 32); py := make([]byte, 32)
    copy(px[32-len(x):], x); copy(py[32-len(y):], y)
    k := types.CoseEC2{Kty:2, Alg:-7, Crv:1, X:px, Y:py}
    acctCBOR, _ := enc.EncodeCanonical(k)
    return priv, k, acctCBOR
}

func insertAcctCredSess(t *testing.T, db *sql.DB, acctCBOR, credID []byte, sc int64, sid string) {
    t.Helper()
    th := sha256.Sum256(append([]byte("ACCTK1"), acctCBOR...))
    now := time.Now().Unix()
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, th[:], now); err != nil { t.Fatalf("ins acct: %v", err) }
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, sc, []byte(nil), now); err != nil { t.Fatalf("ins cred: %v", err) }
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acctCBOR, time.Now().Add(time.Hour).Unix()); err != nil { t.Fatalf("ins sess: %v", err) }
}

func TestPhaseF_E2E_OptionsFinishList(t *testing.T) {
    db := openPhaseDB(t)
    defer db.Close()
    store := NewTxSessionStore(0)

    // Account and credential
    priv, k, acctCBOR := mkAcct(t)
    credID := []byte("cred-e2e-1")
    sid := b64.Encode([]byte("sid-e2e-1234567890123456789012"))
    insertAcctCredSess(t, db, acctCBOR, credID, 0, sid)

    // Build bundle
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "phase-f-e2e"}
    B, _ := enc.EncodeCanonical(bun)
    inB64 := b64.Encode(B)

    // Options
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173"}
    opts, err := BuildTxOptions(context.Background(), cfg, store, db, acctCBOR, inB64, time.Now)
    if err != nil { t.Fatalf("options: %v", err) }

    // Finish via handler
    cdj := map[string]any{"type":"webauthn.get","challenge":opts.Challenge,"origin":"http://localhost:5173"}
    cdjBytes, _ := json.Marshal(cdj)
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := make([]byte, 37)
    copy(ad[:32], rpHash[:]); ad[32] = 0x05; ad[36] = byte(1)
    hcdj := sha256.Sum256(cdjBytes)
    d := sha256.Sum256(append(ad, hcdj[:]...))
    sig, _ := ecdsa.SignASN1(rand.Reader, priv, d[:])
    sig = lowSifyDER2(t, elliptic.P256(), sig)

    body := TxFinishInbound{TxSessionID: opts.TxSessionID, ID: b64.Encode(credID), RawID: b64.Encode(credID), Type: "public-key"}
    body.Response.ClientDataJSON = b64.Encode(cdjBytes)
    body.Response.AuthenticatorData = b64.Encode(ad)
    body.Response.Signature = b64.Encode(sig)
    buf, _ := json.Marshal(body)
    rr := httptest.NewRecorder(); req := httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader(buf))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    TxFinishHandler(cfg, store, db).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("finish status=%d body=%s", rr.Code, rr.Body.String()) }

    // List and verify presence
    rr = httptest.NewRecorder(); req = httptest.NewRequest("GET", "/tx/list", nil)
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    TxListHandler(db).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("list status=%d", rr.Code) }
    var out TxListResponse
    if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil { t.Fatalf("json: %v", err) }
    if len(out.Items) != 1 || out.Items[0].Message != "phase-f-e2e" || out.Items[0].Nonce != 1 {
        t.Fatalf("list mismatch: %+v", out.Items)
    }
}

// lowSifyDER2 mirrors test helper in finish_test.go (scoped locally to avoid cross-file dependency).
func lowSifyDER2(t *testing.T, curve elliptic.Curve, sigDER []byte) []byte {
    t.Helper()
    var s struct{ R, S *big.Int }
    if _, err := asn1.Unmarshal(sigDER, &s); err != nil { t.Fatalf("asn1: %v", err) }
    halfN := new(big.Int).Rsh(curve.Params().N, 1)
    if s.S.Cmp(halfN) == 1 {
        s.S.Sub(curve.Params().N, s.S)
    }
    out, err := asn1.Marshal(s)
    if err != nil { t.Fatalf("asn1 marshal: %v", err) }
    return out
}
