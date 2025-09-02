package webauthn

import (
    "bytes"
    "crypto/elliptic"
    "crypto/sha256"
    "encoding/binary"
    "encoding/json"
    "net/http/httptest"
    "path/filepath"
    "testing"
    "time"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    storepkg "github.com/neutral/passkey-demo/internal/storage"
)

func mkADWithParams(rpHash [32]byte, flags byte, signCount uint32, aaguid [16]byte, credID []byte, cose []byte) []byte {
    b := make([]byte, 37)
    copy(b[:32], rpHash[:])
    b[32] = flags
    binary.BigEndian.PutUint32(b[33:37], signCount)
    out := append([]byte{}, b...)
    out = append(out, aaguid[:]...)
    l := make([]byte, 2)
    binary.BigEndian.PutUint16(l, uint16(len(credID)))
    out = append(out, l...)
    out = append(out, credID...)
    out = append(out, cose...)
    return out
}

func mkRegFinishPayload(t *testing.T, cfg *cfgpkg.Config, regSessionID string, challengeB64 string) []byte {
    t.Helper()
    // Build CDJ JSON string matching challenge and origin
    cdj := map[string]any{
        "type": "webauthn.create",
        "challenge": challengeB64,
        "origin": cfg.Origin,
    }
    cdjBytes, _ := json.Marshal(cdj)
    cdjB64 := enc.Encode(cdjBytes)
    // Build minimal struct (id/rawId = credID placeholder; set later by caller)
    body := map[string]any{
        "reg_session_id": regSessionID,
        "id": "",
        "rawId": "",
        "type": "public-key",
        "response": map[string]any{
            "clientDataJSON": cdjB64,
            "attestationObject": "",
        },
    }
    b, _ := json.Marshal(body)
    return b
}

func TestRegistrationFinish_Happy(t *testing.T) {
    // Config for localhost dev
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    // DB
    db, err := storepkg.Open(cfg)
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storepkg.Migrate(db); err != nil { t.Fatalf("db migrate: %v", err) }
    // Store + options to create a valid session
    store := NewRegSessionStore(0)
    opts, err := BuildRegistrationOptions(cfg, store, time.Now)
    if err != nil { t.Fatalf("build opts: %v", err) }

    // Prepare attestationObject
    // COSE EC2 using curve base point
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32); copy(px[32-len(gx):], gx)
    py := make([]byte, 32); copy(py[32-len(gy):], gy)
    cose := struct{
        Kty int    `cbor:"1"`
        Alg int    `cbor:"3"`
        Crv int    `cbor:"-1"`
        X   []byte `cbor:"-2"`
        Y   []byte `cbor:"-3"`
    }{Kty:2, Alg:-7, Crv:1, X:px, Y:py}
    coseCBOR, _ := enc.EncodeCanonical(cose)
    var aaguid [16]byte
    copy(aaguid[:], []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16})
    credID := []byte("cred-123456")
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := mkADWithParams(rpHash, FlagAT|FlagUV|FlagUP, 1, aaguid, credID, coseCBOR)
    attObj := map[string]any{"fmt": "none", "authData": ad, "attStmt": map[string]any{}}
    attCBOR, _ := enc.EncodeCanonical(attObj)

    // Build payload JSON with placeholders
    body := mkRegFinishPayload(t, cfg, opts.RegSessionID, opts.Challenge)
    // Patch in id/rawId and attestationObject
    var m map[string]any
    _ = json.Unmarshal(body, &m)
    m["id"] = enc.Encode(credID)
    m["rawId"] = enc.Encode(credID)
    resp := m["response"].(map[string]any)
    resp["attestationObject"] = enc.Encode(attCBOR)
    body, _ = json.Marshal(m)

    h := RegistrationFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/registration/finish", bytes.NewReader(body))
    h.ServeHTTP(rr, req)
    if rr.Code != 201 { t.Fatalf("status: %d, body=%s", rr.Code, rr.Body.String()) }

    // Verify DB rows exist
    var count int
    if err := db.QueryRow("SELECT COUNT(1) FROM accounts").Scan(&count); err != nil || count != 1 {
        t.Fatalf("accounts count err=%v count=%d", err, count)
    }
    if err := db.QueryRow("SELECT COUNT(1) FROM credentials").Scan(&count); err != nil || count != 1 {
        t.Fatalf("credentials count err=%v count=%d", err, count)
    }
    // Ensure session deleted (single-use)
    if _, ok := store.Get(opts.RegSessionID); ok {
        t.Fatalf("session should be deleted")
    }
}

func mkFinishBody(t *testing.T, cfg *cfgpkg.Config, regSessionID string, challengeB64 string, origin string, credID []byte, attCBOR []byte) []byte {
    t.Helper()
    cdj := map[string]any{"type": "webauthn.create", "challenge": challengeB64, "origin": origin}
    cdjBytes, _ := json.Marshal(cdj)
    cdjB64 := enc.Encode(cdjBytes)
    body := map[string]any{
        "reg_session_id": regSessionID,
        "id": enc.Encode(credID),
        "rawId": enc.Encode(credID),
        "type": "public-key",
        "response": map[string]any{
            "clientDataJSON": cdjB64,
            "attestationObject": enc.Encode(attCBOR),
        },
    }
    b, _ := json.Marshal(body)
    return b
}

func TestRegistrationFinish_ExpiredSession(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, _ := storepkg.Open(cfg); _ = storepkg.Migrate(db)
    store := NewRegSessionStore(0)
    // Insert expired session
    _ = store.Put("expired", RegSession{Challenge: []byte{1,2,3}, RP_ID: cfg.RP_ID, Origin: cfg.Origin, ExpiresAt: time.Now().Add(-time.Minute)})
    h := RegistrationFinishHandler(cfg, store, db)
    // Minimal body
    body := []byte(`{"reg_session_id":"expired","response":{"clientDataJSON":"","attestationObject":""}}`)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/registration/finish", bytes.NewReader(body))
    h.ServeHTTP(rr, req)
    if rr.Code != 401 {
        t.Fatalf("want 401, got %d", rr.Code)
    }
}

func TestRegistrationFinish_WrongChallenge(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, _ := storepkg.Open(cfg); _ = storepkg.Migrate(db)
    store := NewRegSessionStore(0)
    opts, _ := BuildRegistrationOptions(cfg, store, time.Now)
    // Valid attestation object (fmt none) with correct rpIdHash and UV
    gx := elliptic.P256().Params().Gx.Bytes(); gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32); copy(px[32-len(gx):], gx)
    py := make([]byte, 32); copy(py[32-len(gy):], gy)
    cose := struct{
        Kty int    `cbor:"1"`
        Alg int    `cbor:"3"`
        Crv int    `cbor:"-1"`
        X   []byte `cbor:"-2"`
        Y   []byte `cbor:"-3"`
    }{Kty:2,Alg:-7,Crv:1,X:px,Y:py}
    coseCBOR, _ := enc.EncodeCanonical(cose)
    var aaguid [16]byte
    credID := []byte("credX")
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := mkADWithParams(rpHash, FlagAT|FlagUV, 1, aaguid, credID, coseCBOR)
    attObj := map[string]any{"fmt":"none","authData": ad, "attStmt": map[string]any{}}
    attCBOR, _ := enc.EncodeCanonical(attObj)
    // Use wrong challenge
    wrongChB64 := enc.Encode([]byte("not-the-challenge"))
    body := mkFinishBody(t, cfg, opts.RegSessionID, wrongChB64, cfg.Origin, credID, attCBOR)
    h := RegistrationFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/registration/finish", bytes.NewReader(body))
    h.ServeHTTP(rr, req)
    if rr.Code != 401 { t.Fatalf("want 401, got %d", rr.Code) }
}

func TestRegistrationFinish_BadOrigin(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, _ := storepkg.Open(cfg); _ = storepkg.Migrate(db)
    store := NewRegSessionStore(0)
    opts, _ := BuildRegistrationOptions(cfg, store, time.Now)
    // Valid attestation
    gx := elliptic.P256().Params().Gx.Bytes(); gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32); copy(px[32-len(gx):], gx)
    py := make([]byte, 32); copy(py[32-len(gy):], gy)
    cose := struct{
        Kty int    `cbor:"1"`
        Alg int    `cbor:"3"`
        Crv int    `cbor:"-1"`
        X   []byte `cbor:"-2"`
        Y   []byte `cbor:"-3"`
    }{Kty:2,Alg:-7,Crv:1,X:px,Y:py}
    coseCBOR, _ := enc.EncodeCanonical(cose)
    var aaguid [16]byte
    credID := []byte("credY")
    rpHash := sha256.Sum256([]byte(cfg.RP_ID))
    ad := mkADWithParams(rpHash, FlagAT|FlagUV, 1, aaguid, credID, coseCBOR)
    attObj := map[string]any{"fmt":"none","authData": ad, "attStmt": map[string]any{}}
    attCBOR, _ := enc.EncodeCanonical(attObj)
    // Wrong origin (https://example.com) should be 400 due to scheme policy vs dev localhost
    body := mkFinishBody(t, cfg, opts.RegSessionID, opts.Challenge, "https://example.com", credID, attCBOR)
    h := RegistrationFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/registration/finish", bytes.NewReader(body))
    h.ServeHTTP(rr, req)
    if rr.Code != 403 { t.Fatalf("want 403, got %d", rr.Code) }
}

func TestRegistrationFinish_RpIdMismatch_NoUV_Duplicate(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, _ := storepkg.Open(cfg); _ = storepkg.Migrate(db)
    store := NewRegSessionStore(0)
    // Build attestation with mismatched rpIdHash
    gx := elliptic.P256().Params().Gx.Bytes(); gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32); copy(px[32-len(gx):], gx)
    py := make([]byte, 32); copy(py[32-len(gy):], gy)
    cose := struct{
        Kty int    `cbor:"1"`
        Alg int    `cbor:"3"`
        Crv int    `cbor:"-1"`
        X   []byte `cbor:"-2"`
        Y   []byte `cbor:"-3"`
    }{Kty:2,Alg:-7,Crv:1,X:px,Y:py}
    coseCBOR, _ := enc.EncodeCanonical(cose)
    var aaguid [16]byte
    credID := []byte("credZ")
    wrongRpHash := sha256.Sum256([]byte("other.com"))
    adWrong := mkADWithParams(wrongRpHash, FlagAT|FlagUV, 1, aaguid, credID, coseCBOR)
    attWrong := map[string]any{"fmt":"none","authData": adWrong, "attStmt": map[string]any{}}
    attWrongCBOR, _ := enc.EncodeCanonical(attWrong)
    // Session
    opts, _ := BuildRegistrationOptions(cfg, store, time.Now)
    bodyWrong := mkFinishBody(t, cfg, opts.RegSessionID, opts.Challenge, cfg.Origin, credID, attWrongCBOR)
    h := RegistrationFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder(); req := httptest.NewRequest("POST", "/", bytes.NewReader(bodyWrong))
    h.ServeHTTP(rr, req)
    if rr.Code != 403 { t.Fatalf("want 403 for rpId mismatch, got %d", rr.Code) }

    // No UV
    store2 := NewRegSessionStore(0)
    opts2, _ := BuildRegistrationOptions(cfg, store2, time.Now)
    adNoUV := mkADWithParams(sha256.Sum256([]byte(cfg.RP_ID)), FlagAT|FlagUP /* no UV */, 1, aaguid, credID, coseCBOR)
    attNoUV := map[string]any{"fmt":"none","authData": adNoUV, "attStmt": map[string]any{}}
    attNoUVCBOR, _ := enc.EncodeCanonical(attNoUV)
    bodyNoUV := mkFinishBody(t, cfg, opts2.RegSessionID, opts2.Challenge, cfg.Origin, credID, attNoUVCBOR)
    h2 := RegistrationFinishHandler(cfg, store2, db)
    rr2 := httptest.NewRecorder(); req2 := httptest.NewRequest("POST", "/", bytes.NewReader(bodyNoUV))
    h2.ServeHTTP(rr2, req2)
    if rr2.Code != 403 { t.Fatalf("want 403 for no UV, got %d", rr2.Code) }

    // Duplicate credential id: perform happy path twice with same credID using new session
    store3 := NewRegSessionStore(0)
    opts3, _ := BuildRegistrationOptions(cfg, store3, time.Now)
    adGood := mkADWithParams(sha256.Sum256([]byte(cfg.RP_ID)), FlagAT|FlagUV, 1, aaguid, credID, coseCBOR)
    attGood := map[string]any{"fmt":"none","authData": adGood, "attStmt": map[string]any{}}
    attGoodCBOR, _ := enc.EncodeCanonical(attGood)
    bodyGood := mkFinishBody(t, cfg, opts3.RegSessionID, opts3.Challenge, cfg.Origin, credID, attGoodCBOR)
    h3 := RegistrationFinishHandler(cfg, store3, db)
    rr3 := httptest.NewRecorder(); req3 := httptest.NewRequest("POST", "/", bytes.NewReader(bodyGood))
    h3.ServeHTTP(rr3, req3)
    if rr3.Code != 201 { t.Fatalf("first insert status: %d", rr3.Code) }
    // Second try with new session
    store4 := NewRegSessionStore(0)
    opts4, _ := BuildRegistrationOptions(cfg, store4, time.Now)
    bodyDup := mkFinishBody(t, cfg, opts4.RegSessionID, opts4.Challenge, cfg.Origin, credID, attGoodCBOR)
    h4 := RegistrationFinishHandler(cfg, store4, db)
    rr4 := httptest.NewRecorder(); req4 := httptest.NewRequest("POST", "/", bytes.NewReader(bodyDup))
    h4.ServeHTTP(rr4, req4)
    if rr4.Code != 409 { t.Fatalf("want 409 duplicate, got %d", rr4.Code) }
}

func TestRegistrationFinish_UnsupportedFmt_BadAttestation(t *testing.T) {
    cfg := &cfgpkg.Config{RP_ID: "example.com", Origin: "http://localhost:5173", DBPath: filepath.Join(t.TempDir(), "test.db")}
    db, _ := storepkg.Open(cfg); _ = storepkg.Migrate(db)
    store := NewRegSessionStore(0)
    opts, _ := BuildRegistrationOptions(cfg, store, time.Now)
    // Unsupported fmt
    ad := mkADWithParams(sha256.Sum256([]byte(cfg.RP_ID)), FlagAT|FlagUV, 1, [16]byte{}, []byte("cid"), []byte{0xa0})
    attFmt := map[string]any{"fmt":"packed","authData": ad, "attStmt": map[string]any{}}
    attFmtCBOR, _ := enc.EncodeCanonical(attFmt)
    bodyFmt := mkFinishBody(t, cfg, opts.RegSessionID, opts.Challenge, cfg.Origin, []byte("cid"), attFmtCBOR)
    h := RegistrationFinishHandler(cfg, store, db)
    rr := httptest.NewRecorder(); req := httptest.NewRequest("POST", "/", bytes.NewReader(bodyFmt))
    h.ServeHTTP(rr, req)
    if rr.Code != 400 { t.Fatalf("want 400 for unsupported fmt, got %d", rr.Code) }

    // Bad CBOR (attestationObject not decodable)
    store2 := NewRegSessionStore(0)
    opts2, _ := BuildRegistrationOptions(cfg, store2, time.Now)
    bodyBad := mkFinishBody(t, cfg, opts2.RegSessionID, opts2.Challenge, cfg.Origin, []byte("cid"), []byte{0xff})
    h2 := RegistrationFinishHandler(cfg, store2, db)
    rr2 := httptest.NewRecorder(); req2 := httptest.NewRequest("POST", "/", bytes.NewReader(bodyBad))
    h2.ServeHTTP(rr2, req2)
    if rr2.Code != 400 { t.Fatalf("want 400 for bad attestation, got %d", rr2.Code) }
}
