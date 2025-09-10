package webauthn

import "testing"

func FuzzParseClientDataJSON(f *testing.F) {
    f.Add([]byte(`{"type":"webauthn.get","challenge":"AA","origin":"http://localhost:5173"}`))
    f.Fuzz(func(t *testing.T, data []byte) {
        _, _ = ParseClientDataJSON(data)
    })
}

