package vectors

import (
    "crypto/elliptic"
    "crypto/sha256"
    "encoding/hex"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
)

// Vectors is the JSON-serializable shape for the golden reference.
type Vectors struct {
    Version string `json:"version"`
    Inputs  struct {
        Message           string `json:"message"`
        Nonce             uint64 `json:"nonce"`
        SenderKeyCBORB64  string `json:"sender_key_cbor_b64"`
        SenderKeyCBORHex  string `json:"sender_key_cbor_hex"`
    } `json:"inputs"`
    Bundle struct {
        BundleCBORB64 string `json:"bundle_cbor_b64"`
        BundleCBORHex string `json:"bundle_cbor_hex"`
    } `json:"bundle"`
    Anchors struct {
        ChallengeB64 string `json:"challenge_b64"`
        ChallengeHex string `json:"challenge_hex"`
        TxIDHex      string `json:"tx_id_hex"`
    } `json:"anchors"`
}

// mkDeterministicCose returns a COSE EC2 key at the P-256 base point,
// with X and Y left-padded to 32 bytes.
func mkDeterministicCose() types.CoseEC2 {
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32)
    py := make([]byte, 32)
    copy(px[32-len(gx):], gx)
    copy(py[32-len(gy):], gy)
    return types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: px, Y: py}
}

// Generate produces the golden vectors using deterministic inputs.
//
// - COSE EC2 key derived from P-256 base point (Gx,Gy)
// - Bundle: { sender_key=k, nonce=1, message="hello" }
// - Encodings: canonical CBOR; base64url unpadded strings
// - Anchors:
//     challenge = SHA-256("CHALv1" || B)
//     tx_id     = SHA-256("TXIDv1" || B)
func Generate() (Vectors, error) {
    var out Vectors
    out.Version = "tx-bundle-v1"

    // Deterministic inputs
    k := mkDeterministicCose()
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "hello"}

    // Canonical encodings
    acctCBOR, err := enc.EncodeCanonical(k)
    if err != nil { return out, err }
    B, err := enc.EncodeCanonical(bun)
    if err != nil { return out, err }

    // Anchors
    chal := sha256.Sum256(append([]byte("CHALv1"), B...))
    txid := sha256.Sum256(append([]byte("TXIDv1"), B...))

    // Fill outputs
    out.Inputs.Message = bun.Message
    out.Inputs.Nonce = bun.Nonce
    out.Inputs.SenderKeyCBORB64 = enc.Encode(acctCBOR)
    out.Inputs.SenderKeyCBORHex = hex.EncodeToString(acctCBOR)

    out.Bundle.BundleCBORB64 = enc.Encode(B)
    out.Bundle.BundleCBORHex = hex.EncodeToString(B)

    out.Anchors.ChallengeB64 = enc.Encode(chal[:])
    out.Anchors.ChallengeHex = hex.EncodeToString(chal[:])
    out.Anchors.TxIDHex = hex.EncodeToString(txid[:])

    return out, nil
}

