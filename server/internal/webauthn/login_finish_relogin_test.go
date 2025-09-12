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

// Test that re-login works after expiring the previous session in the DB.
func TestLoginFinish_ReLoginAfterExpiry_Succeeds(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, err := storepkg.Open(cfg)
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storepkg.Migrate(db); err != nil { t.Fatalf("db migrate: %v", err) }

    // Generate ES256 keypair and insert account + credential
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("keygen: %v", err) }
    cose := types.CoseEC2{Kty:2, Alg:-7, Crv:1, X: pad32(priv.X.Bytes()), Y: pad32(priv.Y.Bytes())}
    acctCBOR, _ := enc.EncodeCanonical(cose)
    credID := []byte("cred-relogin-1")
    now := time.Now().Unix()
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, acctThumb(acctCBOR), now); err != nil {
        t.Fatalf("insert acct: %v", err)
    }
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, int64(5), []byte(nil), now); err != nil {
        t.Fatalf("insert cred: %v", err)
    }

    // Helper to perform a login finish and return sid and status
    doLogin := func(signCount uint32) (string, int) {
        store := NewLoginSessionStore(0)
        opts, err := BuildLoginOptions(cfg, store, time.Now)
        if err != nil { t.Fatalf("build opts: %v", err) }
        // Build CDJ JSON
        cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
        cdjBytes, _ := json.Marshal(cdj)
        // Build AD with UV and provided signCount
        rpHash := sha256.Sum256([]byte(cfg.RP_ID))
        ad := mkADHdr(rpHash, FlagUV|FlagUP, signCount)
        // Sign digest
        hcdj := sha256.Sum256(cdjBytes)
        d := sha256.Sum256(append(ad, hcdj[:]...))
        sig, _ := ecdsa.SignASN1(rand.Reader, priv, d[:])
        sig = lowSify(elliptic.P256(), sig)
        // Build request body
        body := map[string]any{
            "login_session_id": opts.LoginSessionID,
            "id": enc.Encode(credID),
            "rawId": enc.Encode(credID),
            "type": "public-key",
            "response": map[string]any{
                "clientDataJSON": enc.Encode(cdjBytes),
                "authenticatorData": enc.Encode(ad),
                "signature": enc.Encode(sig),
                "userHandle": "",
            },
        }
        b, _ := json.Marshal(body)
        rr := httptest.NewRecorder()
        req := httptest.NewRequest("POST", "/authn/passkey/login/finish", bytes.NewReader(b))
        LoginFinishHandler(cfg, store, db).ServeHTTP(rr, req)
        // Extract sid
        set := rr.Header().Get("Set-Cookie")
        sid := ""
        for _, part := range bytes.Split([]byte(set), []byte(";")) {
            if bytes.HasPrefix(part, []byte("sid=")) { sid = string(bytes.TrimPrefix(part, []byte("sid="))); break }
        }
        return sid, rr.Code
    }

    // First login
    sid1, code := doLogin(6)
    if code != 200 { t.Fatalf("first login code=%d", code) }
    if sid1 == "" { t.Fatalf("first login missing sid") }

    // Expire the session
    if _, err := db.Exec(`UPDATE sessions SET expires_at = ? WHERE session_id = ?`, time.Now().Add(-time.Minute).Unix(), sid1); err != nil {
        t.Fatalf("expire session: %v", err)
    }

    // Second login after expiry should succeed and issue a new sid
    sid2, code := doLogin(7)
    if code != 200 { t.Fatalf("second login code=%d", code) }
    if sid2 == "" { t.Fatalf("second login missing sid") }
    if sid2 == sid1 { t.Fatalf("expected new sid; got same value") }
}

