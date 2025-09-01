package webauthn

import (
    "bytes"
    "crypto/elliptic"
    "encoding/binary"
    "testing"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
)

func mkADWithAtt(flags byte, signCount uint32, aaguid [16]byte, credID []byte, cose []byte) []byte {
    // Base AD: rpIdHash (32 x 0xaa), flags, signCount
    b := make([]byte, 37)
    for i := 0; i < 32; i++ { b[i] = 0xaa }
    b[32] = flags
    binary.BigEndian.PutUint32(b[33:37], signCount)
    // Attested credential data
    out := append([]byte{}, b...)
    out = append(out, aaguid[:]...)
    l := make([]byte, 2)
    binary.BigEndian.PutUint16(l, uint16(len(credID)))
    out = append(out, l...)
    out = append(out, credID...)
    out = append(out, cose...)
    return out
}

func mkCoseEC2(t *testing.T) ([]byte, types.CoseEC2) {
    t.Helper()
    // Pick a valid point from P-256 base point multiples (use Gx,Gy)
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    // Pad to 32
    px := make([]byte, 32); copy(px[32-len(gx):], gx)
    py := make([]byte, 32); copy(py[32-len(gy):], gy)
    k := types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: px, Y: py}
    b, err := enc.EncodeCanonical(k)
    if err != nil { t.Fatalf("cbor: %v", err) }
    return b, k
}

func mkAttObj(t *testing.T, ad []byte, fmtStr string) []byte {
    t.Helper()
    m := map[string]any{
        "fmt": fmtStr,
        "authData": ad,
        "attStmt": map[string]any{},
    }
    b, err := enc.EncodeCanonical(m)
    if err != nil { t.Fatalf("attobj cbor: %v", err) }
    return b
}

func TestExtractRegistrationData_Happy(t *testing.T) {
    coseB, _ := mkCoseEC2(t)
    var aaguid [16]byte
    copy(aaguid[:], []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16})
    credID := []byte("credential-id-1234567890")
    ad := mkADWithAtt(FlagAT|FlagUP|FlagUV, 7, aaguid, credID, coseB)
    attObj := mkAttObj(t, ad, "none")

    adOut, aag, cid, cose, err := ExtractRegistrationData(attObj)
    if err != nil { t.Fatalf("extract: %v", err) }
    if adOut.SignCount != 7 || adOut.Flags&(FlagAT) == 0 { t.Fatalf("ad header mismatch") }
    if aag != aaguid { t.Fatalf("aaguid mismatch") }
    if !bytes.Equal(cid, credID) { t.Fatalf("credID mismatch") }
    if cose.Kty != 2 || cose.Alg != -7 || cose.Crv != 1 || len(cose.X) != 32 || len(cose.Y) != 32 { t.Fatalf("cose mismatch: %+v", cose) }
}

func TestExtractRegistrationData_Errors(t *testing.T) {
    coseB, _ := mkCoseEC2(t)
    var aaguid [16]byte
    credID := []byte("cred")
    // Missing AT flag
    adNoAT := mkADWithAtt(0, 0, aaguid, credID, coseB)
    att1 := mkAttObj(t, adNoAT, "none")
    if _, _, _, _, err := ExtractRegistrationData(att1); err == nil {
        t.Fatalf("expected missing AT error")
    }
    // Unsupported fmt
    ad := mkADWithAtt(FlagAT, 0, aaguid, credID, coseB)
    att2 := mkAttObj(t, ad, "packed")
    if _, _, _, _, err := ExtractRegistrationData(att2); err == nil {
        t.Fatalf("expected fmt error")
    }
    // Truncated AD (short)
    bad := ad[:20]
    att3 := mkAttObj(t, bad, "none")
    if _, _, _, _, err := ExtractRegistrationData(att3); err == nil {
        t.Fatalf("expected auth data short error")
    }
    // Bad COSE
    badCbor := []byte{0xff} // invalid CBOR
    adBadCose := mkADWithAtt(FlagAT, 0, aaguid, credID, badCbor)
    att4 := mkAttObj(t, adBadCose, "none")
    if _, _, _, _, err := ExtractRegistrationData(att4); err == nil {
        t.Fatalf("expected cose decode error")
    }
}
