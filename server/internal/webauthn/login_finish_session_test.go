package webauthn

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/json"
    "net/http/httptest"
    "path/filepath"
    "testing"
    "time"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
)

// Test that login finish issues a session with ~1h expiry recorded in DB.
func TestLoginFinish_SetsOneHourExpiry(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, err := storepkg.Open(cfg)
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storepkg.Migrate(db); err != nil { t.Fatalf("db migrate: %v", err) }

    // Generate ES256 keypair for account
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("keygen: %v", err) }
    cose := types.CoseEC2{Kty:2, Alg:-7, Crv:1, X: pad32(priv.X.Bytes()), Y: pad32(priv.Y.Bytes())}
    acctCBOR, _ := enc.EncodeCanonical(cose)
    // Insert account and credential with initial sign_count
    credID := []byte("cred-login-expiry-1")
    now := time.Now().Unix()
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, acctThumb(acctCBOR), now); err != nil {
        t.Fatalf("insert acct: %v", err)
    }
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, int64(1), []byte(nil), now); err != nil {
        t.Fatalf("insert cred: %v", err)
    }

    // Create login session via builder
    store := NewLoginSessionStore(0)
    opts, err := BuildLoginOptions(cfg, store, time.Now)
    if err != nil { t.Fatalf("build opts: %v", err) }

    // Build CDJ JSON
    cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
    cdjBytes, _ := json.Marshal(cdj)
    cdjB64 := enc.Encode(cdjBytes)

    // Build AD header with UV and monotonic signCount
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := mkADHdr(rpHash, FlagUV|FlagUP, 2)

    // Compute digest and sign (low-S normalized)
    hcdj := sha256.Sum256(cdjBytes)
    d := sha256.Sum256(append(ad, hcdj[:]...))
    sigDER, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sigDER = lowSify(elliptic.P256(), sigDER)

    // Build request body
    body := map[string]any{
        "login_session_id": opts.LoginSessionID,
        "id": enc.Encode(credID),
        "rawId": enc.Encode(credID),
        "type": "public-key",
        "response": map[string]any{
            "clientDataJSON": cdjB64,
            "authenticatorData": enc.Encode(ad),
            "signature": enc.Encode(sigDER),
            "userHandle": "",
        },
    }
    bodyBytes, _ := json.Marshal(body)

    // Call handler
    h := LoginFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/login/finish", bytes.NewReader(bodyBytes))
    tBefore := time.Now().Unix()
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status: %d, body=%s", rr.Code, rr.Body.String()) }

    // Extract session id from Set-Cookie
    set := rr.Header().Get("Set-Cookie")
    if set == "" { t.Fatalf("missing Set-Cookie") }
    // Best-effort parse sid token
    var sid string
    for _, part := range bytes.Split([]byte(set), []byte(";")) {
        if bytes.HasPrefix(part, []byte("sid=")) {
            sid = string(bytes.TrimPrefix(part, []byte("sid=")))
            break
        }
    }
    if sid == "" { t.Fatalf("sid not found in Set-Cookie: %q", set) }

    // Assert expires_at is ~1 hour in the future (allow tolerance)
    var exp int64
    if err := db.QueryRow(`SELECT expires_at FROM sessions WHERE session_id = ?`, sid).Scan(&exp); err != nil {
        t.Fatalf("select exp: %v", err)
    }
    // Allow generous tolerance to avoid flakes in CI environments
    // Expect roughly 3600 seconds; accept [3500, 3700]
    delta := exp - tBefore
    if delta < 3500 || delta > 3700 {
        t.Fatalf("unexpected expiry delta: got %d (exp=%d, before=%d)", delta, exp, tBefore)
    }
}

