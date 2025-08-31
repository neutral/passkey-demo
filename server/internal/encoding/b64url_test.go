package encoding

import (
    crand "crypto/rand"
    "encoding/base64"
    "io"
    "testing"
)

func TestEncodeDecode_RoundtripVariousSizes(t *testing.T) {
    sizes := []int{0, 1, 2, 3, 4, 31, 32, 33, 1024, 65536}
    for _, n := range sizes {
        buf := make([]byte, n)
        if _, err := io.ReadFull(crand.Reader, buf); err != nil {
            t.Fatalf("rand: %v", err)
        }
        enc := Encode(buf)
        out, err := Decode(enc)
        if err != nil {
            t.Fatalf("decode %d: %v", n, err)
        }
        if len(out) != len(buf) {
            t.Fatalf("len mismatch %d", n)
        }
        for i := range out {
            if out[i] != buf[i] {
                t.Fatalf("byte mismatch at %d size %d", i, n)
            }
        }
    }
}

func TestDecode_TolerantPaddedAndUnpadded(t *testing.T) {
    data := []byte("hello world")
    padded := base64.URLEncoding.EncodeToString(data)
    unpadded := base64.RawURLEncoding.EncodeToString(data)

    a, err := Decode(padded)
    if err != nil { t.Fatalf("padded: %v", err) }
    b, err := Decode(unpadded)
    if err != nil { t.Fatalf("unpadded: %v", err) }
    if string(a) != string(data) || string(b) != string(data) {
        t.Fatalf("decoded mismatch")
    }
}

func TestDecode_InvalidInputs(t *testing.T) {
    bad := []string{"*", "====", "ab==?", "A-_/=="}
    for _, s := range bad {
        if _, err := Decode(s); err == nil {
            t.Fatalf("expected error for %q", s)
        }
    }
}

