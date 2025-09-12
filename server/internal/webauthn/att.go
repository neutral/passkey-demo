package webauthn

import (
    "bytes"
    "encoding/binary"
    "errors"
    "log/slog"

    cbor "github.com/fxamacker/cbor/v2"
    enc "github.com/neutral/passkey-demo/internal/encoding"
    types "github.com/neutral/passkey-demo/internal/types"
)

// Attestation parsing errors.
var (
    ErrAttestationCBOR    = errors.New("attestation CBOR decode failed")
    ErrAttestationFormat  = errors.New("unsupported attestation format")
    ErrAttestedDataMissing = errors.New("attested credential data missing (AT flag not set)")
    ErrAuthDataShort      = errors.New("authenticatorData too short")
    ErrCredIDLength       = errors.New("credential ID length invalid")
    ErrCOSEKeyDecode      = errors.New("COSE key decode failed")
)

// cbor modes from encoding package
var (
    _ = enc.EncodeCanonical
    _ = enc.DecodeCanonical
)

// AttestationObject mirrors the WebAuthn attestationObject structure.
type AttestationObject struct {
    Fmt      string                 `cbor:"fmt"`
    AuthData []byte                 `cbor:"authData"`
    AttStmt  map[string]any         `cbor:"attStmt"`
}

// ParseAttestationObject decodes the top-level CBOR attestation object.
func ParseAttestationObject(b []byte) (AttestationObject, error) {
    var ao AttestationObject
    if err := enc.DecodeCanonical(b, &ao); err != nil {
        return AttestationObject{}, ErrAttestationCBOR
    }
    if len(ao.AuthData) < 37 {
        return AttestationObject{}, ErrAuthDataShort
    }
    return ao, nil
}

// ParseAttestedCredentialData parses AAGUID, credentialId, and COSE key from the
// bytes following the 37-byte authenticatorData header.
func ParseAttestedCredentialData(b []byte) (aaguid [16]byte, credID []byte, coseRaw []byte, rest []byte, err error) {
    if len(b) < 18 { // need at least AAGUID(16) + L(2)
        return aaguid, nil, nil, nil, ErrAuthDataShort
    }
    copy(aaguid[:], b[:16])
    credLen := int(binary.BigEndian.Uint16(b[16:18]))
    if len(b) < 18+credLen {
        return aaguid, nil, nil, nil, ErrCredIDLength
    }
    credID = make([]byte, credLen)
    copy(credID, b[18:18+credLen])
    // The COSE_Key CBOR follows immediately. Decode a single CBOR item and return raw bytes and rest.
    rem := b[18+credLen:]
    rdr := bytes.NewReader(rem)
    // Use fxamacker's decMode via a default DecOptions (same as enc.DecodeCanonical under the hood)
    dec, _ := (cbor.DecOptions{}).DecMode()
    d := dec.NewDecoder(rdr)
    var raw cbor.RawMessage
    if err := d.Decode(&raw); err != nil {
        return aaguid, nil, nil, nil, ErrCOSEKeyDecode
    }
    coseRaw = []byte(raw)
    // Remaining bytes after COSE key are extensions (rest)
    rest = rem[len(rem)-rdr.Len():]
    return aaguid, credID, coseRaw, rest, nil
}

// ParseCOSEKeyEC2 decodes a COSE EC2 key into our struct.
func ParseCOSEKeyEC2(b []byte) (types.CoseEC2, error) {
    var k types.CoseEC2
    // First try decoding as a COSE_Key map directly.
    if err := enc.DecodeCanonical(b, &k); err == nil {
        if k.Kty != 0 && len(k.X) > 0 && len(k.Y) > 0 {
            return k, nil
        }
    }
    // Some authenticators wrap the COSE_Key inside a CBOR byte string (bstr) or tag 24 "encoded CBOR".
    // Try to decode a nested byte string and then decode that as a COSE_Key.
    var inner []byte
    if err := enc.DecodeCanonical(b, &inner); err == nil && len(inner) > 0 {
        var k2 types.CoseEC2
        if err2 := enc.DecodeCanonical(inner, &k2); err2 == nil {
            if k2.Kty != 0 && len(k2.X) > 0 && len(k2.Y) > 0 {
                return k2, nil
            }
        }
    }
    // Try to decode as a CBOR tag (e.g., tag 24) that wraps a byte string with COSE_Key CBOR inside.
    var tag cbor.Tag
    if err := enc.DecodeCanonical(b, &tag); err == nil {
        if tag.Number == 24 {
            if payload, ok := tag.Content.([]byte); ok && len(payload) > 0 {
                var k3 types.CoseEC2
                if err3 := enc.DecodeCanonical(payload, &k3); err3 == nil {
                    if k3.Kty != 0 && len(k3.X) > 0 && len(k3.Y) > 0 {
                        return k3, nil
                    }
                }
            }
        }
    }
    // Last resort: decode into a generic map and extract fields 1,3,-1,-2,-3.
    var generic map[any]any
    if err := enc.DecodeCanonical(b, &generic); err == nil && len(generic) > 0 {
        // Helper to pull an int or uint into Go int
        getInt := func(key any) (int, bool) {
            switch t := key.(type) {
            case int:
                return t, true
            case int64:
                return int(t), true
            case uint64:
                return int(t), true
            default:
                return 0, false
            }
        }
        // Re-map by integer keys if possible
        ints := make(map[int]any)
        for k0, v := range generic {
            if ik, ok := getInt(k0); ok {
                ints[ik] = v
            }
        }
        var out types.CoseEC2
        if v, ok := ints[1]; ok {
            if i, ok2 := getInt(v); ok2 {
                out.Kty = i
            }
        }
        if v, ok := ints[3]; ok {
            if i, ok2 := getInt(v); ok2 {
                out.Alg = i
            }
        }
        if v, ok := ints[-1]; ok {
            if i, ok2 := getInt(v); ok2 {
                out.Crv = i
            }
        }
        if v, ok := ints[-2]; ok {
            if bs, ok2 := v.([]byte); ok2 {
                out.X = bs
            }
        }
        if v, ok := ints[-3]; ok {
            if bs, ok2 := v.([]byte); ok2 {
                out.Y = bs
            }
        }
        if out.Kty != 0 && len(out.X) > 0 && len(out.Y) > 0 {
            return out, nil
        }
    }
    return types.CoseEC2{}, ErrCOSEKeyDecode
}

// ExtractRegistrationData parses the attestation object, validates fmt, and returns
// the parsed authenticatorData header and attested credential data (AAGUID, credID, COSE EC2 key).
func ExtractRegistrationData(attObjB []byte) (ad AuthData, aaguid [16]byte, credID []byte, cose types.CoseEC2, err error) {
    ao, err := ParseAttestationObject(attObjB)
    if err != nil {
        slog.Info("reg_finish_debug", slog.String("detail", "attestation_cbor_decode_failed"))
        return ad, aaguid, nil, cose, err
    }
    // Demo-grade: accept attestation fmt "none" and "packed" without trust evaluation.
    // We only rely on the attested credential data (AAGUID, credential ID, COSE key).
    if ao.Fmt != "none" && ao.Fmt != "packed" {
        slog.Info("reg_finish_debug", slog.String("detail", "unsupported_attestation_fmt"), slog.String("fmt", ao.Fmt))
        return ad, aaguid, nil, cose, ErrAttestationFormat
    }
    // Parse AD header
    ad, remainder, err := ParseAuthData(ao.AuthData)
    if err != nil {
        slog.Info("reg_finish_debug", slog.String("detail", "parse_authdata_failed"), slog.Int("authdata_len", len(ao.AuthData)))
        return ad, aaguid, nil, cose, err
    }
    if (ad.Flags & FlagAT) == 0 {
        slog.Info("reg_finish_debug", slog.String("detail", "at_flag_not_set"), slog.Int("flags", int(ad.Flags)))
        return ad, aaguid, nil, cose, ErrAttestedDataMissing
    }
    // Parse attested cred data
    aaguid, credID, coseRaw, _, err := ParseAttestedCredentialData(remainder)
    if err != nil {
        slog.Info("reg_finish_debug", slog.String("detail", "parse_attested_data_failed"), slog.Int("rem_len", len(remainder)))
        return ad, aaguid, nil, cose, err
    }
    // Decode COSE EC2
    cose, err = ParseCOSEKeyEC2(coseRaw)
    if err != nil {
        slog.Info("reg_finish_debug", slog.String("detail", "parse_cose_key_failed"), slog.Int("cose_len", len(coseRaw)))
        return ad, aaguid, nil, cose, err
    }
    return ad, aaguid, credID, cose, nil
}
