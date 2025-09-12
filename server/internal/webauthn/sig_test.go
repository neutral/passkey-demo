package webauthn

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "encoding/asn1"
    "errors"
    "math/big"
    "testing"
)

// digest mirrors the VerifyAssertion message digest construction.
func digest(ad, cdj []byte) [32]byte {
    hcdj := sha256.Sum256(cdj)
    msg := append([]byte{}, ad...)
    msg = append(msg, hcdj[:]...)
    return sha256.Sum256(msg)
}

func lowSify(curve elliptic.Curve, der []byte) []byte {
    var s ecdsaSig
    _, _ = asn1.Unmarshal(der, &s)
    if s.R == nil || s.S == nil {
        return der
    }
    if !isLowS(curve, s.S) {
        // convert to low-S: s = N - s
        s.S = new(big.Int).Sub(curve.Params().N, s.S)
    }
    out, _ := asn1.Marshal(s)
    return out
}

func highSify(curve elliptic.Curve, der []byte) []byte {
    var s ecdsaSig
    _, _ = asn1.Unmarshal(der, &s)
    if s.R == nil || s.S == nil {
        return der
    }
    if isLowS(curve, s.S) {
        // convert to high-S: s = N - s
        s.S = new(big.Int).Sub(curve.Params().N, s.S)
    }
    out, _ := asn1.Marshal(s)
    return out
}

func TestVerifyAssertion_HappyPath(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping crypto-heavy verify tests in -short mode")
    }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("authData-bytes-123")
    cdj := []byte("clientDataJSON-bytes-456")
    d := digest(ad, cdj)
    sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sig = lowSify(elliptic.P256(), sig)
    if err := VerifyAssertion(&priv.PublicKey, ad, cdj, sig); err != nil {
        t.Fatalf("verify: %v", err)
    }
}

func TestVerifyAssertion_HighSRejected(t *testing.T) {
    if testing.Short() { t.Skip("skipping crypto-heavy verify tests in -short mode") }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("A")
    cdj := []byte("B")
    d := digest(ad, cdj)
    sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sigHigh := highSify(elliptic.P256(), sig)
    if err := VerifyAssertion(&priv.PublicKey, ad, cdj, sigHigh); err == nil || !errors.Is(err, ErrHighS) {
        t.Fatalf("expected ErrHighS, got: %v", err)
    }
}

func TestVerifyAssertion_TamperAD(t *testing.T) {
    if testing.Short() { t.Skip("skipping crypto-heavy verify tests in -short mode") }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("authData-bytes-123")
    cdj := []byte("clientDataJSON-bytes-456")
    d := digest(ad, cdj)
    sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sig = lowSify(elliptic.P256(), sig)
    ad[0] ^= 0x01
    if err := VerifyAssertion(&priv.PublicKey, ad, cdj, sig); err == nil || !errors.Is(err, ErrBadSignature) {
        t.Fatalf("expected ErrBadSignature after AD tamper, got: %v", err)
    }
}

func TestVerifyAssertion_TamperCDJ(t *testing.T) {
    if testing.Short() { t.Skip("skipping crypto-heavy verify tests in -short mode") }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("authData-bytes-123")
    cdj := []byte("clientDataJSON-bytes-456")
    d := digest(ad, cdj)
    sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sig = lowSify(elliptic.P256(), sig)
    cdj[0] ^= 0x01
    if err := VerifyAssertion(&priv.PublicKey, ad, cdj, sig); err == nil || !errors.Is(err, ErrBadSignature) {
        t.Fatalf("expected ErrBadSignature after CDJ tamper, got: %v", err)
    }
}

func TestVerifyAssertion_MalformedDER(t *testing.T) {
    if testing.Short() { t.Skip("skipping crypto-heavy verify tests in -short mode") }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("authData")
    cdj := []byte("clientData")
    // Clearly invalid/truncated DER
    bad := []byte{0x30, 0x00}
    if err := VerifyAssertion(&priv.PublicKey, ad, cdj, bad); err == nil || !errors.Is(err, ErrMalformedDER) {
        t.Fatalf("expected ErrMalformedDER, got: %v", err)
    }
}

func TestVerifyAssertion_TrailingBytesDER(t *testing.T) {
    if testing.Short() { t.Skip("skipping crypto-heavy verify tests in -short mode") }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("authData")
    cdj := []byte("clientData")
    d := digest(ad, cdj)
    sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sig = lowSify(elliptic.P256(), sig)
    sig = append(sig, 0x00) // add trailing byte
    if err := VerifyAssertion(&priv.PublicKey, ad, cdj, sig); err == nil || !errors.Is(err, ErrMalformedDER) {
        t.Fatalf("expected ErrMalformedDER for trailing bytes, got: %v", err)
    }
}

func TestVerifyAssertion_WrongCurveAndNil(t *testing.T) {
    if testing.Short() { t.Skip("skipping crypto-heavy verify tests in -short mode") }
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil { t.Fatalf("gen key: %v", err) }
    ad := []byte("authData")
    cdj := []byte("clientData")
    d := digest(ad, cdj)
    sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
    if err != nil { t.Fatalf("sign: %v", err) }
    sig = lowSify(elliptic.P256(), sig)

    if err := VerifyAssertion(nil, ad, cdj, sig); err == nil || !errors.Is(err, ErrUnsupportedCurve) {
        t.Fatalf("expected ErrUnsupportedCurve for nil pub, got: %v", err)
    }
    wrong := &ecdsa.PublicKey{Curve: elliptic.P384(), X: priv.X, Y: priv.Y}
    if err := VerifyAssertion(wrong, ad, cdj, sig); err == nil || !errors.Is(err, ErrUnsupportedCurve) {
        t.Fatalf("expected ErrUnsupportedCurve for wrong curve, got: %v", err)
    }
}
