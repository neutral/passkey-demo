package crypto

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "errors"
    "math/big"

    types "github.com/neutral/passkey-demo/internal/types"
)

// ToECDSA converts a COSE EC2 public key (ES256 / P-256) into a Go ecdsa.PublicKey.
// It validates kty/alg/crv, enforces 32-byte X/Y, and ensures the point lies on P-256.
func ToECDSA(k *types.CoseEC2) (*ecdsa.PublicKey, error) {
    if k == nil {
        return nil, errors.New("nil key")
    }
    // COSE constants for EC2 / ES256 / P-256
    if k.Kty != 2 { // EC2
        return nil, errors.New("unsupported kty: want EC2 (2)")
    }
    if k.Alg != -7 { // ES256
        return nil, errors.New("unsupported alg: want ES256 (-7)")
    }
    if k.Crv != 1 { // P-256
        return nil, errors.New("unsupported crv: want P-256 (1)")
    }
    if len(k.X) != 32 || len(k.Y) != 32 {
        return nil, errors.New("invalid coordinate length: want 32-byte X and Y")
    }
    x := new(big.Int).SetBytes(k.X)
    y := new(big.Int).SetBytes(k.Y)
    curve := elliptic.P256()
    if !curve.IsOnCurve(x, y) {
        return nil, errors.New("point not on P-256 curve")
    }
    return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

