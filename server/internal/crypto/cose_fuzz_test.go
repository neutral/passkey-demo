package crypto

import (
    "testing"
    types "github.com/neutral/passkey-demo/internal/types"
)

func FuzzToECDSA(f *testing.F) {
    f.Add(make([]byte, 32), make([]byte, 32))
    f.Fuzz(func(t *testing.T, x []byte, y []byte) {
        k := &types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: pad32(x), Y: pad32(y)}
        _, _ = ToECDSA(k)
    })
}
