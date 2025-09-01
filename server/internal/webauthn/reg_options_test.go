package webauthn

import (
    "encoding/json"
    "net/http/httptest"
    "testing"
    "time"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
)

func testCfg() *cfgpkg.Config {
    return &cfgpkg.Config{RP_ID: "example.com", Origin: "https://example.com"}
}

func TestBuildRegistrationOptions_Happy(t *testing.T) {
    store := NewRegSessionStore(0)
    now := func() time.Time { return time.Unix(1_700_000_000, 0) }
    resp, err := BuildRegistrationOptions(testCfg(), store, now)
    if err != nil { t.Fatalf("build: %v", err) }
    // Challenge length
    ch, err := b64.Decode(resp.Challenge)
    if err != nil || len(ch) != 32 { t.Fatalf("challenge decode/len: %v %d", err, len(ch)) }
    // Session id entropy
    sid, err := b64.Decode(resp.RegSessionID)
    if err != nil || len(sid) < 16 { t.Fatalf("sid decode/len: %v %d", err, len(sid)) }
    // Options fields
    if resp.Options.RP_ID != "example.com" || resp.Options.Origin != "https://example.com" || !resp.Options.UVRequired || resp.Options.Attestation != "none" {
        t.Fatalf("options mismatch: %+v", resp.Options)
    }
    // Expires at ≈ now + 5m
    want := now().Add(5 * time.Minute).Unix()
    if resp.ExpiresAt != want {
        t.Fatalf("expires_at mismatch: got %d want %d", resp.ExpiresAt, want)
    }
    // Store entry exists and matches
    sess, ok := store.Get(resp.RegSessionID)
    if !ok { t.Fatalf("session not stored") }
    if string(sess.Challenge) != string(ch) || sess.RP_ID != "example.com" || sess.Origin != "https://example.com" {
        t.Fatalf("session mismatch: %+v", sess)
    }
}

func TestRegistrationOptionsHandler_JSON(t *testing.T) {
    store := NewRegSessionStore(0)
    h := RegistrationOptionsHandler(testCfg(), store)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/authn/passkey/registration/options", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("status: %d", rr.Code) }
    var out RegistrationOptionsResponse
    if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
        t.Fatalf("json: %v", err)
    }
    if out.RegSessionID == "" || out.Challenge == "" || out.Options.RP_ID == "" || out.Options.Origin == "" || out.ExpiresAt == 0 {
        t.Fatalf("missing fields in response: %+v", out)
    }
    // Verify store has entry
    if _, ok := store.Get(out.RegSessionID); !ok {
        t.Fatalf("store entry missing for %s", out.RegSessionID)
    }
}

