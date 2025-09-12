package tx

import (
    "bytes"
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"

    enc "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
    cbor "github.com/fxamacker/cbor/v2"
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
    // Fallback: some encoders may use tags/extensions that trip struct decode for nested COSE.
    // If SenderKey appears empty, try a generic decode to extract COSE fields.
    if bun.SenderKey.Kty == 0 && bun.SenderKey.Alg == 0 && bun.SenderKey.Crv == 0 &&
        len(bun.SenderKey.X) == 0 && len(bun.SenderKey.Y) == 0 {
        // Robust fallback using fxamacker RawMessage to extract nested COSE map (key 0)
        var rm cbor.RawMessage
        // Try map[int64]
        if rm == nil {
            var top map[int64]cbor.RawMessage
            if err := cbor.Unmarshal(raw, &top); err == nil {
                if v, ok := top[0]; ok && len(v) > 0 { rm = v }
            }
        }
        // Try map[uint64]
        if rm == nil {
            var top map[uint64]cbor.RawMessage
            if err := cbor.Unmarshal(raw, &top); err == nil {
                if v, ok := top[0]; ok && len(v) > 0 { rm = v }
            }
        }
        // Try map[any]
        if rm == nil {
            var top map[any]cbor.RawMessage
            if err := cbor.Unmarshal(raw, &top); err == nil {
                for k, v := range top {
                    switch kv := k.(type) {
                    case int:
                        if kv == 0 && len(v) > 0 { rm = v }
                    case int64:
                        if kv == 0 && len(v) > 0 { rm = v }
                    case uint64:
                        if kv == 0 && len(v) > 0 { rm = v }
                    }
                }
            }
        }
        if len(rm) > 0 {
            // First try typed struct
            var cose struct {
                Kty int    `cbor:"1"`
                Alg int    `cbor:"3"`
                Crv int    `cbor:"-1"`
                X   []byte `cbor:"-2"`
                Y   []byte `cbor:"-3"`
            }
            if err := cbor.Unmarshal(rm, &cose); err == nil {
                bun.SenderKey.Kty = cose.Kty
                bun.SenderKey.Alg = cose.Alg
                bun.SenderKey.Crv = cose.Crv
                bun.SenderKey.X = append([]byte(nil), cose.X...)
                bun.SenderKey.Y = append([]byte(nil), cose.Y...)
            }
            // If still empty, try generic map decode and coerce
            if bun.SenderKey.Kty == 0 && bun.SenderKey.Alg == 0 && bun.SenderKey.Crv == 0 &&
                len(bun.SenderKey.X) == 0 && len(bun.SenderKey.Y) == 0 {
                var gm map[any]any
                if err := cbor.Unmarshal(rm, &gm); err == nil {
                    getInt := func(v any) (int, bool) {
                        switch t := v.(type) {
                        case int:
                            return t, true
                        case int64:
                            return int(t), true
                        case uint64:
                            return int(t), true
                        case float64:
                            return int(t), true
                        default:
                            return 0, false
                        }
                    }
                    getBytes := func(v any) ([]byte, bool) {
                        switch t := v.(type) {
                        case []byte:
                            return append([]byte(nil), t...), true
                        case cbor.Tag:
                            // unwrap tagged content if it's byte string
                            if b, ok := t.Content.([]byte); ok {
                                return append([]byte(nil), b...), true
                            }
                            return nil, false
                        default:
                            return nil, false
                        }
                    }
                    for k, v := range gm {
                        var keyInt int
                        ok := false
                        switch kk := k.(type) {
                        case int:
                            keyInt, ok = kk, true
                        case int64:
                            keyInt, ok = int(kk), true
                        case uint64:
                            keyInt, ok = int(kk), true
                        case string:
                            // attempt to parse string form of numeric key
                            // ignore errors
                        }
                        if !ok {
                            continue
                        }
                        switch keyInt {
                        case 1:
                            if i, ok := getInt(v); ok { bun.SenderKey.Kty = i }
                        case 3:
                            if i, ok := getInt(v); ok { bun.SenderKey.Alg = i }
                        case -1:
                            if i, ok := getInt(v); ok { bun.SenderKey.Crv = i }
                        case -2:
                            if b, ok := getBytes(v); ok { bun.SenderKey.X = b }
                        case -3:
                            if b, ok := getBytes(v); ok { bun.SenderKey.Y = b }
                        }
                    }
                }
            }
        }
    }
    // If nonce wasn't populated due to encoder variance, attempt a generic decode
    // to extract `nonce` (key 1) and `message` (key 2). This guards against cases
    // where CBOR libraries encode small integers as floats or the typed struct
    // mapping is bypassed.
    if bun.Nonce == 0 || bun.Message == "" {
        var gm map[any]any
        if err := enc.DecodeCanonical(raw, &gm); err == nil {
            // helper to coerce numeric to uint64
            getU64 := func(v any) (uint64, bool) {
                switch t := v.(type) {
                case uint64:
                    return t, true
                case int64:
                    if t >= 0 { return uint64(t), true }
                case int:
                    if t >= 0 { return uint64(t), true }
                case float64:
                    if t >= 0 && t == float64(uint64(t)) { return uint64(t), true }
                }
                return 0, false
            }
            for k, v := range gm {
                var keyInt int
                switch kk := k.(type) {
                case int:
                    keyInt = kk
                case int64:
                    keyInt = int(kk)
                case uint64:
                    keyInt = int(kk)
                default:
                    continue
                }
                switch keyInt {
                case 1:
                    if bun.Nonce == 0 { if u, ok := getU64(v); ok { bun.Nonce = u } }
                case 2:
                    if bun.Message == "" { if s, ok := v.(string); ok { bun.Message = s } }
                }
            }
        }
    }
    // Reject zero nonce (must be positive integer)
    if bun.Nonce == 0 {
        return nil, ErrBundleCBOR
    }
    // Canonical re-encode → B
    B, err := enc.EncodeCanonical(bun)
    if err != nil {
        return nil, ErrBundleCBOR
    }
    // Account binding: compare logical COSE fields to tolerate encoder differences
    var acct types.CoseEC2
    if err := enc.DecodeCanonical(acctCBOR, &acct); err != nil {
        return nil, ErrBundleCBOR
    }
    if bun.SenderKey.Kty != acct.Kty || bun.SenderKey.Alg != acct.Alg || bun.SenderKey.Crv != acct.Crv ||
        !bytes.Equal(bun.SenderKey.X, acct.X) || !bytes.Equal(bun.SenderKey.Y, acct.Y) {
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
