package types

import (
    "testing"
    enc "github.com/neutral/passkey-demo/internal/encoding"
)

func TestBundle_Roundtrip_WithAndWithoutValidUntil(t *testing.T) {
    k := CoseEC2{Kty:2, Alg:-7, Crv:1, X: make([]byte,32), Y: make([]byte,32)}
    b1 := Bundle{SenderKey: k, Nonce: 42, Message: "hello", ValidUntil: nil}
    c1, err := enc.EncodeCanonical(b1)
    if err != nil { t.Fatalf("enc1: %v", err) }
    var out1 Bundle
    if err := enc.DecodeCanonical(c1, &out1); err != nil { t.Fatalf("dec1: %v", err) }
    if out1.ValidUntil != nil { t.Fatalf("valid_until should be omitted when nil") }

    vu := uint64(1234567890)
    b2 := Bundle{SenderKey: k, Nonce: 42, Message: "hello", ValidUntil: &vu}
    c2, err := enc.EncodeCanonical(b2)
    if err != nil { t.Fatalf("enc2: %v", err) }
    var out2 Bundle
    if err := enc.DecodeCanonical(c2, &out2); err != nil { t.Fatalf("dec2: %v", err) }
    if out2.ValidUntil == nil || *out2.ValidUntil != vu { t.Fatalf("valid_until lost in roundtrip") }

    // Re-encode and compare bytes to ensure stability of roundtrip
    c1b, _ := enc.EncodeCanonical(out1)
    c2b, _ := enc.EncodeCanonical(out2)
    if string(c1b) != string(c1) || string(c2b) != string(c2) {
        t.Fatalf("canonical re-encode mismatch") }
}

