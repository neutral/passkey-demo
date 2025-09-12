package crypto

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "math/big"
    "testing"

    types "github.com/neutral/passkey-demo/internal/types"
)

func pad32(b []byte) []byte {
    if len(b) >= 32 {
        return b
    }
    out := make([]byte, 32)
    copy(out[32-len(b):], b)
    return out
}

func TestToECDSA_ValidP256(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping crypto-heavy ToECDSA test in -short mode")
    }
    // Generate a valid P-256 keypair and convert the public part to COSE EC2
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        t.Fatalf("generate key: %v", err)
    }
    k := &types.CoseEC2{
        Kty: 2,
        Alg: -7,
        Crv: 1,
        X:   pad32(priv.PublicKey.X.Bytes()),
        Y:   pad32(priv.PublicKey.Y.Bytes()),
    }
    pub, err := ToECDSA(k)
    if err != nil {
        t.Fatalf("ToECDSA valid: %v", err)
    }
    if pub.Curve != elliptic.P256() {
        t.Fatalf("curve mismatch")
    }
    if pub.X.Cmp(priv.PublicKey.X) != 0 || pub.Y.Cmp(priv.PublicKey.Y) != 0 {
        t.Fatalf("coordinates mismatch")
    }
    if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
        t.Fatalf("not on curve")
    }
}

func TestToECDSA_InvalidParams(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping crypto-heavy ToECDSA invalid params in -short mode")
    }
    // Start from a valid point
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        t.Fatalf("generate key: %v", err)
    }
    base := &types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: pad32(priv.PublicKey.X.Bytes()), Y: pad32(priv.PublicKey.Y.Bytes())}

    cases := []struct{
        name string
        mutate func(*types.CoseEC2)
    }{
        {"nil", func(k *types.CoseEC2) { /* handled separately */ }},
        {"bad-kty", func(k *types.CoseEC2) { k.Kty = 1 }},
        {"bad-alg", func(k *types.CoseEC2) { k.Alg = -8 }},
        {"bad-crv", func(k *types.CoseEC2) { k.Crv = 2 }},
        {"short-x", func(k *types.CoseEC2) { k.X = []byte{1} }},
        {"short-y", func(k *types.CoseEC2) { k.Y = []byte{1} }},
        {"zero-point", func(k *types.CoseEC2) { k.X = make([]byte, 32); k.Y = make([]byte, 32) }},
        {"off-curve", func(k *types.CoseEC2) {
            yy := new(big.Int).Add(new(big.Int).SetBytes(k.Y), big.NewInt(1))
            k.Y = pad32(yy.Bytes())
        }},
    }

    // nil case
    if _, err := ToECDSA(nil); err == nil {
        t.Fatalf("expected error for nil key")
    }

    for _, tc := range cases[1:] {
        t.Run(tc.name, func(t *testing.T) {
            // copy base
            k := *base
            tc.mutate(&k)
            if _, err := ToECDSA(&k); err == nil {
                t.Fatalf("expected error for %s", tc.name)
            }
        })
    }
}

// no extra helpers
