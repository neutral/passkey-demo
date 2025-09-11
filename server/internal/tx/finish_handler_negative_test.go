package tx

import (
    "context"
    "bytes"
    "crypto/elliptic"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
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
    httpctx "github.com/neutral/passkey-demo/internal/http"
    repos "github.com/neutral/passkey-demo/internal/repos"
)

func openDBFinishNeg(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    return db
}

func mkAcctDeterministic(t *testing.T) (types.CoseEC2, []byte) {
    t.Helper()
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32)
    py := make([]byte, 32)
    copy(px[32-len(gx):], gx)
    copy(py[32-len(gy):], gy)
    k := types.CoseEC2{Kty:2, Alg:-7, Crv:1, X:px, Y:py}
    acctCBOR, _ := enc.EncodeCanonical(k)
    return k, acctCBOR
}

func insertAuthSession(t *testing.T, db *sql.DB, sid string, acctCBOR []byte, exp time.Time) {
    t.Helper()
    if _, err := db.Exec(`INSERT INTO sessions (session_id, acct_cbor, expires_at, created_at) VALUES (?, ?, ?, 0)`, sid, acctCBOR, exp.Unix()); err != nil {
        t.Fatalf("ins sess: %v", err)
    }
}

func TestTxFinishHandler_ErrorMappings(t *testing.T) {
    db := openDBFinishNeg(t)
    defer db.Close()
    store := NewTxSessionStore(0)
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173"}

    // Missing cookie/session → 401
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader([]byte("{}")))
    credsRepo, err := repos.NewCredentials(context.Background(), db)
    if err != nil { t.Fatalf("repo: %v", err) }
    h := httpctx.SessionMiddleware(db, true, time.Hour)(TxFinishHandler(cfg, store, db, credsRepo))
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("missing cookie: %d", rr.Code) }

    // Prepare account + auth session
    _, acct := mkAcctDeterministic(t)
    sid := b64.Encode([]byte("sid-fin-neg-1234567890123456789012"))
    insertAuthSession(t, db, sid, acct, time.Now().Add(time.Hour))

    // Expired tx session → 401 (early)
    sessID := b64.Encode([]byte("txsess-exp-123456789012345678901"))
    _ = store.Put(sessID, TxSession{AcctCBOR: acct, Challenge: []byte("CH"), ExpiresAt: time.Now().Add(-time.Minute)})
    bodyExp := TxFinishInbound{TxSessionID: sessID}
    bufExp, _ := json.Marshal(bodyExp)
    rr = httptest.NewRecorder(); req = httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader(bufExp))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("expired tx session: %d", rr.Code) }

    // Challenge mismatch → 401
    // Create valid tx session (not expired)
    // Build a bundle to derive canonical B and proper challenge for realism
    k, _ := mkAcctDeterministic(t)
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "m"}
    B, _ := enc.EncodeCanonical(bun)
    sum := sha256.Sum256(append([]byte(anchorChallengePrefix), B...))
    goodCh := b64.Encode(sum[:])
    sessID2 := b64.Encode([]byte("txsess-2-123456789012345678901"))
    chBytes, _ := b64.Decode(goodCh)
    _ = store.Put(sessID2, TxSession{AcctCBOR: acct, Challenge: chBytes, ExpiresAt: time.Now().Add(time.Minute)})
    // But send CDJ with a different challenge
    badCh := b64.Encode([]byte("ZZ"))
    cdj := map[string]any{"type": "webauthn.get", "challenge": badCh, "origin": cfg.Origin}
    cdjBytes, _ := json.Marshal(cdj)
    in := TxFinishInbound{TxSessionID: sessID2, RawID: b64.Encode([]byte("id")), ID: b64.Encode([]byte("id"))}
    in.Response.ClientDataJSON = b64.Encode(cdjBytes)
    buf, _ := json.Marshal(in)
    rr = httptest.NewRecorder(); req = httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader(buf))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("challenge mismatch: %d", rr.Code) }

    // Origin not allowed → 403 via policy mapping
    // Use correct challenge to get to origin check
    cdjOK := map[string]any{"type": "webauthn.get", "challenge": goodCh, "origin": "http://bad.example"}
    cdjOKBytes, _ := json.Marshal(cdjOK)
    in2 := TxFinishInbound{TxSessionID: sessID2, RawID: b64.Encode([]byte("id")), ID: b64.Encode([]byte("id"))}
    in2.Response.ClientDataJSON = b64.Encode(cdjOKBytes)
    buf2, _ := json.Marshal(in2)
    rr = httptest.NewRecorder(); req = httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader(buf2))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 403 { t.Fatalf("origin not allowed: %d", rr.Code) }

    // Bad JSON → 400 (malformed body)
    rr = httptest.NewRecorder(); req = httptest.NewRequest("POST", "/tx/signing/finish", bytes.NewReader([]byte("{")))
    req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
    h.ServeHTTP(rr, req)
    if rr.Code != 400 { t.Fatalf("bad json: %d", rr.Code) }
}
