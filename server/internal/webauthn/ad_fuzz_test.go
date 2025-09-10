package webauthn

import "testing"

func FuzzParseAuthData(f *testing.F) {
    // Seed with minimal valid length (37 bytes) and one short case
    f.Add(make([]byte, 37))
    f.Add([]byte{1, 2, 3})
    f.Fuzz(func(t *testing.T, data []byte) {
        _, _, _ = ParseAuthData(data)
    })
}

