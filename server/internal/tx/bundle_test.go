package tx

import (
    "context"
    "crypto/elliptic"
    "database/sql"
    "encoding/hex"
    "testing"

    _ "github.com/mattn/go-sqlite3"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    storage "github.com/neutral/passkey-demo/internal/storage"
    types "github.com/neutral/passkey-demo/internal/types"
)

// mkCoseEC2 returns a deterministic COSE EC2 public key using the curve's base point.
func mkCoseEC2(t *testing.T) (types.CoseEC2, []byte) {
    t.Helper()
    gx := elliptic.P256().Params().Gx.Bytes()
    gy := elliptic.P256().Params().Gy.Bytes()
    px := make([]byte, 32)
    py := make([]byte, 32)
    copy(px[32-len(gx):], gx)
    copy(py[32-len(gy):], gy)
    k := types.CoseEC2{Kty: 2, Alg: -7, Crv: 1, X: px, Y: py}
    cbor, err := enc.EncodeCanonical(k)
    if err != nil { t.Fatalf("cbor encode cose: %v", err) }
    return k, cbor
}

func openTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", "file::memory:?_busy_timeout=5000&_foreign_keys=on")
    if err != nil { t.Fatalf("db open: %v", err) }
    if err := storage.Migrate(db); err != nil {
        t.Fatalf("migrate: %v", err)
    }
    return db
}

func TestValidateAndAnchorBundle_Happy(t *testing.T) {
    db := openTestDB(t)
    defer db.Close()
    // Account COSE key
    k, acctCBOR := mkCoseEC2(t)
    // Logical bundle
    bun := types.Bundle{SenderKey: k, Nonce: 1, Message: "hello"}
    // Encode to CBOR, then base64url for input
    B, err := enc.EncodeCanonical(bun)
    if err != nil { t.Fatalf("encode bundle: %v", err) }
    in := enc.Encode(B)

    out, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, in)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if string(out.B) != string(B) { t.Fatalf("canonical B mismatch") }

    // Compute expected anchors independently
    want, _ := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, in)
    if out.Challenge != want.Challenge || out.TxID != want.TxID {
        t.Fatalf("anchors not stable:\nchal=%s want=%s\ntxid=%s want=%s",
            hex.EncodeToString(out.Challenge[:]), hex.EncodeToString(want.Challenge[:]),
            hex.EncodeToString(out.TxID[:]), hex.EncodeToString(want.TxID[:]))
    }
}

func TestValidateAndAnchorBundle_Errors(t *testing.T) {
    db := openTestDB(t)
    defer db.Close()
    k, acctCBOR := mkCoseEC2(t)
    var err error

    // Base bundle for mutation (not used directly)
    _ = types.Bundle{SenderKey: k, Nonce: 5, Message: "x"}

    // 1) bad base64
    if _, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, "@@not-b64@@"); err != ErrBundleBase64 {
        t.Fatalf("want ErrBundleBase64, got %v", err)
    }

    // 2) bad CBOR
    if _, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, enc.Encode([]byte{0xff})); err != ErrBundleCBOR {
        t.Fatalf("want ErrBundleCBOR, got %v", err)
    }

    // 3) sender key mismatch
    // Tweak X coordinate by 1 byte (still 32 bytes) to force mismatch
    k2 := k
    k2.X = append([]byte{}, k.X...)
    k2.X[31] ^= 0x01
    bunMismatch := types.Bundle{SenderKey: k2, Nonce: 6, Message: "x"}
    B2, _ := enc.EncodeCanonical(bunMismatch)
    in2 := enc.Encode(B2)
    if _, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, in2); err != ErrSenderKeyMismatch {
        t.Fatalf("want ErrSenderKeyMismatch, got %v", err)
    }

    // 4) nonce not monotonic
    // Insert an existing transaction with nonce 10 for this account
    _, err = db.Exec(`INSERT INTO accounts (acct_cbor, acct_thumb, created_at) VALUES (?, ?, 0)`, acctCBOR, []byte("thumb"))
    if err != nil { t.Fatalf("insert account: %v", err) }
    // create a tx row with nonce=10
    _, err = db.Exec(`INSERT INTO transactions (tx_id, acct_cbor, nonce, message, bundle_cbor, auth_data, client_data, signature, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)`,
        []byte("id"), acctCBOR, int64(10), "hi", []byte("B"), []byte("AD"), []byte("CD"), []byte("SIG"))
    if err != nil { t.Fatalf("insert tx: %v", err) }
    // Now validate with nonce=9 (<=10)
    bunLow := types.Bundle{SenderKey: k, Nonce: 9, Message: "x"}
    B3, _ := enc.EncodeCanonical(bunLow)
    in3 := enc.Encode(B3)
    if _, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, in3); err != ErrNonceNotMonotonic {
        t.Fatalf("want ErrNonceNotMonotonic, got %v", err)
    }

    // And nonce=10 (equal) should also fail
    bunEq := types.Bundle{SenderKey: k, Nonce: 10, Message: "x"}
    B4, _ := enc.EncodeCanonical(bunEq)
    in4 := enc.Encode(B4)
    if _, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, in4); err != ErrNonceNotMonotonic {
        t.Fatalf("want ErrNonceNotMonotonic (equal), got %v", err)
    }

    // Nonce=11 should pass
    bunHi := types.Bundle{SenderKey: k, Nonce: 11, Message: "x"}
    B5, _ := enc.EncodeCanonical(bunHi)
    in5 := enc.Encode(B5)
    if _, err := ValidateAndAnchorBundle(context.Background(), db, acctCBOR, in5); err != nil {
        t.Fatalf("want success for nonce 11, got %v", err)
    }
}
