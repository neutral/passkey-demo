package webauthn

import (
    "encoding/base64"
    "encoding/json"
    "testing"

    b64 "github.com/neutral/passkey-demo/internal/encoding"
)

func cdjJSON(t *testing.T, ch string, typ string, origin string) []byte {
    t.Helper()
    m := map[string]any{
        "type":      typ,
        "challenge": ch,
        "origin":    origin,
    }
    b, _ := json.Marshal(m)
    return b
}

func TestParseClientDataJSON_HappyGet_Unpadded(t *testing.T) {
    raw := []byte{1, 2, 3, 4, 5}
    ch := b64.Encode(raw) // unpadded
    in := cdjJSON(t, ch, "webauthn.get", "http://localhost:5173")
    cd, err := ParseClientDataJSON(in)
    if err != nil { t.Fatalf("parse: %v", err) }
    if !IsGet(cd) || IsCreate(cd) { t.Fatalf("type helpers wrong") }
    if cd.Origin != "http://localhost:5173" { t.Fatalf("origin mismatch") }
    if string(cd.Challenge) != string(raw) { t.Fatalf("challenge mismatch") }
}

func TestParseClientDataJSON_HappyCreate_Padded(t *testing.T) {
    raw := []byte{9, 8, 7}
    ch := base64.URLEncoding.EncodeToString(raw) // padded
    in := cdjJSON(t, ch, "webauthn.create", "http://localhost:5173")
    cd, err := ParseClientDataJSON(in)
    if err != nil { t.Fatalf("parse: %v", err) }
    if !IsCreate(cd) || IsGet(cd) { t.Fatalf("type helpers wrong") }
    if string(cd.Challenge) != string(raw) { t.Fatalf("challenge mismatch") }
}

func TestParseClientDataJSON_Invalids(t *testing.T) {
    // malformed JSON
    if _, err := ParseClientDataJSON([]byte("{")); err == nil {
        t.Fatalf("expected error for bad json")
    }
    // missing fields
    if _, err := ParseClientDataJSON(cdjJSON(t, "", "webauthn.get", "origin")); err == nil {
        t.Fatalf("expected error for missing challenge")
    }
    if _, err := ParseClientDataJSON(cdjJSON(t, "abc", "", "origin")); err == nil {
        t.Fatalf("expected error for missing type")
    }
    if _, err := ParseClientDataJSON(cdjJSON(t, "abc", "webauthn.get", "")); err == nil {
        t.Fatalf("expected error for missing origin")
    }
    // unknown type
    if _, err := ParseClientDataJSON(cdjJSON(t, "abc", "other", "http://localhost")); err == nil {
        t.Fatalf("expected error for unknown type")
    }
    // bad base64
    if _, err := ParseClientDataJSON(cdjJSON(t, "*", "webauthn.get", "http://localhost")); err == nil {
        t.Fatalf("expected error for bad base64 challenge")
    }
}

