package webauthn

import (
    "bytes"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/binary"
    "encoding/json"
    "net/http/httptest"
    "path/filepath"
    "strings"
    "testing"
    "time"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
)

// mkADHdr builds just the 37-byte header of authenticatorData: rpIdHash, flags, signCount
func mkADHdr(rpHash [32]byte, flags byte, signCount uint32) []byte {
    b := make([]byte, 37)
    copy(b[:32], rpHash[:])
    b[32] = flags
    binary.BigEndian.PutUint32(b[33:37], signCount)
    return b
}

func pad32(b []byte) []byte { p := make([]byte, 32); copy(p[32-len(b):], b); return p }

func TestLoginFinish_Happy(t *testing.T) {
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
    credID := []byte("cred-login-1")
    now := time.Now().Unix()
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, acctThumb(acctCBOR), now); err != nil {
        t.Fatalf("insert acct: %v", err)
    }
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, int64(5), []byte(nil), now); err != nil {
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
    ad := mkADHdr(rpHash, FlagUV|FlagUP, 6)

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

    h := LoginFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/login/finish", bytes.NewReader(bodyBytes))
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status: %d, body=%s", rr.Code, rr.Body.String()) }
    // Cookie present with expected attributes for http origin (no Secure)
    if set := rr.Header().Get("Set-Cookie"); !strings.Contains(set, "sid=") {
        t.Fatalf("expected Set-Cookie with sid, got: %q", set)
    } else {
        if !strings.Contains(set, "HttpOnly") { t.Fatalf("cookie missing HttpOnly: %q", set) }
        if !strings.Contains(set, "SameSite=Lax") { t.Fatalf("cookie SameSite not Lax: %q", set) }
        if strings.Contains(set, "Secure") { t.Fatalf("cookie should not be Secure for http origin: %q", set) }
    }
    // sign_count updated
    var sc int64
    if err := db.QueryRow(`SELECT sign_count FROM credentials WHERE credential_id = ?`, credID).Scan(&sc); err != nil || sc != 6 {
        t.Fatalf("sign_count got=%d err=%v", sc, err)
    }
    // session consumed
    if _, ok := store.Get(opts.LoginSessionID); ok { t.Fatalf("login session should be deleted") }
}

func TestSetSessionCookie_SecureToggle(t *testing.T) {
    // http origin → no Secure
    {
        cfg := &cfgpkg.Config{Origin: "http://localhost:5173"}
        rr := httptest.NewRecorder()
        setSessionCookie(rr, cfg, "abc")
        set := rr.Header().Get("Set-Cookie")
        if strings.Contains(set, "Secure") {
            t.Fatalf("unexpected Secure for http origin: %q", set)
        }
        if !strings.Contains(set, "SameSite=Lax") || !strings.Contains(set, "HttpOnly") {
            t.Fatalf("missing attributes: %q", set)
        }
    }
    // https origin → Secure present
    {
        cfg := &cfgpkg.Config{Origin: "https://example.com"}
        rr := httptest.NewRecorder()
        setSessionCookie(rr, cfg, "abc")
        set := rr.Header().Get("Set-Cookie")
        if !strings.Contains(set, "Secure") {
            t.Fatalf("expected Secure for https origin: %q", set)
        }
        if !strings.Contains(set, "SameSite=Lax") || !strings.Contains(set, "HttpOnly") {
            t.Fatalf("missing attributes: %q", set)
        }
    }
}

func TestLoginFinish_Negatives(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, _ := storepkg.Open(cfg); _ = storepkg.Migrate(db)

    // Expired session
    store := NewLoginSessionStore(0)
    _ = store.Put("expired", LoginSession{Challenge: []byte{1,2,3}, RP_ID: cfg.RP_ID, Origin: cfg.Origin, ExpiresAt: time.Now().Add(-time.Minute)})
    h := LoginFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/login/finish", bytes.NewReader([]byte(`{"login_session_id":"expired","response":{}}`)))
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("expired session code=%d", rr.Code) }

    // Prepare account + credential + valid session
    priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    cose := types.CoseEC2{Kty:2, Alg:-7, Crv:1, X: pad32(priv.X.Bytes()), Y: pad32(priv.Y.Bytes())}
    acctCBOR, _ := enc.EncodeCanonical(cose)
    credID := []byte("cred-neg-1")
    _, _ = db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, acctThumb(acctCBOR), time.Now().Unix())
    _, _ = db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, int64(10), []byte(nil), time.Now().Unix())
    store2 := NewLoginSessionStore(0)
    opts, _ := BuildLoginOptions(cfg, store2, time.Now)

    // Helper to send and get code
    send := func(mod func(m map[string]any)) int {
        rpHash := sha256.Sum256([]byte(cfg.RP_ID))
        ad := mkADHdr(rpHash, FlagUV|FlagUP, 11)
        cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
        cdjBytes, _ := json.Marshal(cdj)
        hcdj := sha256.Sum256(cdjBytes)
        d := sha256.Sum256(append(ad, hcdj[:]...))
        sig, _ := ecdsa.SignASN1(rand.Reader, priv, d[:])
        sig = lowSify(elliptic.P256(), sig)
        m := map[string]any{
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
        if mod != nil { mod(m) }
        b, _ := json.Marshal(m)
        rr := httptest.NewRecorder()
        req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
        LoginFinishHandler(cfg, store2, db).ServeHTTP(rr, req)
        return rr.Code
    }

    // Type mismatch
    if c := send(func(m map[string]any) {
        resp := m["response"].(map[string]any)
        // overwrite clientDataJSON with type=create
        bad := map[string]any{"type":"webauthn.create","challenge": opts.Challenge, "origin": cfg.Origin}
        b, _ := json.Marshal(bad)
        resp["clientDataJSON"] = enc.Encode(b)
    }); c != 400 { t.Fatalf("type mismatch got=%d", c) }

    // Challenge mismatch
    if c := send(func(m map[string]any) {
        resp := m["response"].(map[string]any)
        bad := map[string]any{"type":"webauthn.get","challenge": enc.Encode([]byte("wrong")), "origin": cfg.Origin}
        b, _ := json.Marshal(bad)
        resp["clientDataJSON"] = enc.Encode(b)
    }); c != 401 { t.Fatalf("challenge mismatch got=%d", c) }

    // Origin not allowed
    if c := send(func(m map[string]any) {
        resp := m["response"].(map[string]any)
        bad := map[string]any{"type":"webauthn.get","challenge": opts.Challenge, "origin": "https://example.com"}
        b, _ := json.Marshal(bad)
        resp["clientDataJSON"] = enc.Encode(b)
    }); c != 403 { t.Fatalf("origin not allowed got=%d", c) }

    // rpIdHash mismatch
    if c := send(func(m map[string]any) {
        resp := m["response"].(map[string]any)
        // rebuild AD with wrong rpHash
        ad := mkADHdr(sha256.Sum256([]byte("other.com")), FlagUV|FlagUP, 11)
        resp["authenticatorData"] = enc.Encode(ad)
    }); c != 403 { t.Fatalf("rpIdHash mismatch got=%d", c) }

    // Missing UV
    if c := send(func(m map[string]any) {
        resp := m["response"].(map[string]any)
        ad := mkADHdr(sha256.Sum256([]byte(cfg.RP_ID)), FlagUP /* no UV */, 11)
        resp["authenticatorData"] = enc.Encode(ad)
    }); c != 403 { t.Fatalf("missing UV got=%d", c) }

    // Unknown credential id
    if c := send(func(m map[string]any) {
        m["id"] = enc.Encode([]byte("nope"))
        m["rawId"] = enc.Encode([]byte("nope"))
    }); c != 401 { t.Fatalf("unknown cred got=%d", c) }

    // Non-increasing signCount (equal)
    if c := func() int {
        rpHash := sha256.Sum256([]byte(cfg.RP_ID))
        ad := mkADHdr(rpHash, FlagUV|FlagUP, 10) // equal to stored
        cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
        cdjBytes, _ := json.Marshal(cdj)
        hcdj := sha256.Sum256(cdjBytes)
        d := sha256.Sum256(append(ad, hcdj[:]...))
        sig, _ := ecdsa.SignASN1(rand.Reader, priv, d[:])
        sig = lowSify(elliptic.P256(), sig)
        m := map[string]any{
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
        b, _ := json.Marshal(m)
        rr := httptest.NewRecorder(); req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
        LoginFinishHandler(cfg, store2, db).ServeHTTP(rr, req)
        return rr.Code
    }(); c != 409 { t.Fatalf("non-increasing signCount got=%d", c) }

    // Malformed DER
    if c := send(func(m map[string]any) {
        resp := m["response"].(map[string]any)
        resp["signature"] = enc.Encode([]byte{0x30, 0x00})
    }); c != 400 { t.Fatalf("malformed DER got=%d", c) }

    // High-S signature
    if c := func() int {
        rpHash := sha256.Sum256([]byte(cfg.RP_ID))
        ad := mkADHdr(rpHash, FlagUV|FlagUP, 12)
        cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
        cdjBytes, _ := json.Marshal(cdj)
        hcdj := sha256.Sum256(cdjBytes)
        d := sha256.Sum256(append(ad, hcdj[:]...))
        sig, _ := ecdsa.SignASN1(rand.Reader, priv, d[:])
        sig = highSify(elliptic.P256(), sig)
        m := map[string]any{
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
        b, _ := json.Marshal(m)
        rr := httptest.NewRecorder(); req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
        LoginFinishHandler(cfg, store2, db).ServeHTTP(rr, req)
        return rr.Code
    }(); c != 401 { t.Fatalf("high-S got=%d", c) }

    // Bad base64 in rawId
    if c := send(func(m map[string]any) {
        m["rawId"] = "!!!!"
    }); c != 400 { t.Fatalf("bad base64 rawId got=%d", c) }
}

func TestLoginFinish_ZeroSignCountAllowed(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, err := storepkg.Open(cfg)
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storepkg.Migrate(db); err != nil { t.Fatalf("db migrate: %v", err) }

    // Generate key and insert account + credential with stored sign_count = 0
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("keygen: %v", err) }
    cose := types.CoseEC2{Kty:2, Alg:-7, Crv:1, X: pad32(priv.X.Bytes()), Y: pad32(priv.Y.Bytes())}
    acctCBOR, _ := enc.EncodeCanonical(cose)
    credID := []byte("cred-zero-1")
    now := time.Now().Unix()
    if _, err := db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, ?)`, acctCBOR, acctThumb(acctCBOR), now); err != nil {
        t.Fatalf("insert acct: %v", err)
    }
    if _, err := db.Exec(`INSERT INTO credentials (credential_id, acct_cbor_fk, sign_count, aaguid, created_at) VALUES (?, ?, ?, ?, ?)`, credID, acctCBOR, int64(0), []byte(nil), now); err != nil {
        t.Fatalf("insert cred: %v", err)
    }

    // Build session/options
    store := NewLoginSessionStore(0)
    opts, err := BuildLoginOptions(cfg, store, time.Now)
    if err != nil { t.Fatalf("build opts: %v", err) }

    // Build CDJ JSON
    cdj := map[string]any{"type": "webauthn.get", "challenge": opts.Challenge, "origin": cfg.Origin}
    cdjBytes, _ := json.Marshal(cdj)

    // AD header with UV and signCount = 0 (counter not supported)
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := mkADHdr(rpHash, FlagUV|FlagUP, 0)

    // Compute digest and sign
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
            "clientDataJSON": enc.Encode(cdjBytes),
            "authenticatorData": enc.Encode(ad),
            "signature": enc.Encode(sigDER),
            "userHandle": "",
        },
    }
    bodyBytes, _ := json.Marshal(body)

    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/login/finish", bytes.NewReader(bodyBytes))
    LoginFinishHandler(cfg, store, db).ServeHTTP(rr, req)
    if rr.Code != 200 {
        t.Fatalf("expected 200 for zero signCount, got %d body=%s", rr.Code, rr.Body.String())
    }
    // Stored sign_count should remain 0
    var sc int64
    if err := db.QueryRow(`SELECT sign_count FROM credentials WHERE credential_id = ?`, credID).Scan(&sc); err != nil || sc != 0 {
        t.Fatalf("stored sign_count expected 0, got=%d err=%v", sc, err)
    }
}
