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
