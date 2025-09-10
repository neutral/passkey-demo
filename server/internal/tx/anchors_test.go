package tx

import (
    "crypto/elliptic"
    "crypto/sha256"
    "encoding/hex"
    "testing"

    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    storage "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
)

// mkDeterministicCose returns a COSE EC2 key at the P-256 base point.
func mkDeterministicCose(t *testing.T) types.CoseEC2 {
    t.Helper()
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32)
    py := make([]byte, 32)
    copy(px[32-len(gx):], gx)
    copy(py[32-len(gy):], gy)
    return types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: px, Y: py}
}

func TestAnchors_ComputedMatchFormula(t *testing.T) {
    // Open ephemeral DB for nonce check inside ValidateAndAnchorBundle
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil { t.Fatalf("migrate: %v", err) }
    defer db.Close()
    k := mkDeterministicCose(t)
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "hello"}
    B, err := enc.EncodeCanonical(bun)
    if err != nil {
        t.Fatalf("encode canonical: %v", err)
    }
    // Compute anchors as per spec
    ch := sha256.Sum256(append([]byte(anchorChallengePrefix), B...))
    tx := sha256.Sum256(append([]byte(anchorTxIDPrefix), B...))
    if hex.EncodeToString(ch[:]) == hex.EncodeToString(tx[:]) {
        t.Fatalf("challenge and tx_id must differ for same B")
    }
    // ValidateAndAnchorBundle must return same anchors
    // Use a dummy DB-less path by encoding acct_cbor to match sender_key and calling encoder/decoder
    acctCBOR, _ := enc.EncodeCanonical(k)
    out, err := ValidateAndAnchorBundle(nil, db, acctCBOR, enc.Encode(B))
    if err != nil {
        t.Fatalf("validate bundle: %v", err)
    }
    if hex.EncodeToString(out.Challenge[:]) != hex.EncodeToString(ch[:]) ||
        hex.EncodeToString(out.TxID[:]) != hex.EncodeToString(tx[:]) {
        t.Fatalf("anchor mismatch: got chal=%x tx=%x", out.Challenge, out.TxID)
    }
}

func TestAnchors_BChangeFlipsAnchors(t *testing.T) {
    k := mkDeterministicCose(t)
    bun1 := types.Bundle{SenderKey: k, Nonce: 1, Message: "hello"}
    bun2 := types.Bundle{SenderKey: k, Nonce: 2, Message: "hello"}
    B1, _ := enc.EncodeCanonical(bun1)
    B2, _ := enc.EncodeCanonical(bun2)
    ch1 := sha256.Sum256(append([]byte(anchorChallengePrefix), B1...))
    ch2 := sha256.Sum256(append([]byte(anchorChallengePrefix), B2...))
    if hex.EncodeToString(ch1[:]) == hex.EncodeToString(ch2[:]) {
        t.Fatalf("changing nonce should change challenge")
    }
}
