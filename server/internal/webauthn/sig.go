package webauthn

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/sha256"
    "encoding/asn1"
    "errors"
    "math/big"
)

type ecdsaSig struct {
    R, S *big.Int
}

// Exported sentinel errors for granular handling/telemetry.
var (
    ErrUnsupportedCurve = errors.New("unsupported public key/curve: require P-256")
    ErrMalformedDER     = errors.New("malformed ECDSA DER signature")
    ErrHighS            = errors.New("non-low-S signature rejected")
    ErrBadSignature     = errors.New("signature verification failed")
)

func isLowS(curve elliptic.Curve, s *big.Int) bool {
    // low-S means s <= N/2 where N is curve order
    halfN := new(big.Int).Rsh(curve.Params().N, 1)
    return s.Cmp(halfN) <= 0
}

// VerifyAssertion verifies an ES256 signature over SHA256(ad || SHA256(cdj)).
// It enforces P-256 curve, strict DER, and low-S signatures.
func VerifyAssertion(pub *ecdsa.PublicKey, ad, cdj, sigDER []byte) error {
    if pub == nil || pub.Curve != elliptic.P256() {
        return ErrUnsupportedCurve
    }
    hcdj := sha256.Sum256(cdj)
    msg := append([]byte{}, ad...)
    msg = append(msg, hcdj[:]...)
    digest := sha256.Sum256(msg)

    var sig ecdsaSig
    if rest, err := asn1.Unmarshal(sigDER, &sig); err != nil || sig.R == nil || sig.S == nil || len(rest) != 0 {
        return ErrMalformedDER
    }
    if !isLowS(pub.Curve, sig.S) {
        return ErrHighS
    }
    if !ecdsa.Verify(pub, digest[:], sig.R, sig.S) {
        return ErrBadSignature
    }
    return nil
}
