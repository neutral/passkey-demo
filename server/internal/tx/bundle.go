package tx

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
)

// AnchoredBundle holds the parsed logical bundle, its canonical CBOR bytes B,
// and the derived anchors used for WebAuthn challenges and transaction identity.
//
// challenge = SHA256("CHALv1" || B)
// tx_id     = SHA256("TXIDv1" || B)
type AnchoredBundle struct {
    Bundle    types.Bundle
    B         []byte
    Challenge [32]byte
    TxID      [32]byte
}

// Sentinel errors for validation failures.
var (
    ErrBundleBase64      = errors.New("invalid base64url bundle")
    ErrBundleCBOR        = errors.New("invalid bundle CBOR")
    ErrSenderKeyMismatch = errors.New("sender_key does not match account key")
    ErrNonceNotMonotonic = errors.New("nonce must be strictly increasing")
)

const (
    anchorChallengePrefix = "CHALv1"
    anchorTxIDPrefix      = "TXIDv1"
)

// ValidateAndAnchorBundle decodes a client-supplied base64url CBOR bundle, re-encodes it
// canonically to form B, enforces account binding and nonce monotonicity, and computes
// the signing anchors (challenge, tx_id). It performs a read-only DB query to obtain the
// last seen nonce for the account.
func ValidateAndAnchorBundle(ctx context.Context, db *sql.DB, acctCBOR []byte, bundleCBORBase64 string) (*AnchoredBundle, error) {
    // Decode base64url → raw CBOR
    raw, err := enc.Decode(bundleCBORBase64)
    if err != nil {
        return nil, ErrBundleBase64
    }
    // Decode CBOR → logical Bundle
    var bun types.Bundle
    if err := enc.DecodeCanonical(raw, &bun); err != nil {
        return nil, ErrBundleCBOR
    }
    // Canonical re-encode → B
    B, err := enc.EncodeCanonical(bun)
    if err != nil {
        return nil, ErrBundleCBOR
    }
    // Account binding: canonical CBOR of SenderKey must equal acctCBOR
    senderCBOR, err := enc.EncodeCanonical(bun.SenderKey)
    if err != nil {
        return nil, ErrBundleCBOR
    }
    if !bytes.Equal(senderCBOR, acctCBOR) {
        return nil, ErrSenderKeyMismatch
    }
    // Nonce policy: strictly increasing per account
    var last sql.NullInt64
    if err := db.QueryRowContext(ctx, `SELECT MAX(nonce) FROM transactions WHERE acct_cbor = ?`, acctCBOR).Scan(&last); err != nil {
        return nil, err
    }
    if last.Valid {
        if bun.Nonce <= uint64(last.Int64) {
            return nil, ErrNonceNotMonotonic
        }
    }
    // Anchors
    var out AnchoredBundle
    out.Bundle = bun
    out.B = B
    out.Challenge = sha256.Sum256(append([]byte(anchorChallengePrefix), B...))
    out.TxID = sha256.Sum256(append([]byte(anchorTxIDPrefix), B...))
    return &out, nil
}

