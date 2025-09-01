package webauthn

import (
    "bytes"
    "encoding/binary"
    "errors"

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
    if err := enc.DecodeCanonical(b, &k); err != nil {
        return types.CoseEC2{}, ErrCOSEKeyDecode
    }
    return k, nil
}

// ExtractRegistrationData parses the attestation object, validates fmt, and returns
// the parsed authenticatorData header and attested credential data (AAGUID, credID, COSE EC2 key).
func ExtractRegistrationData(attObjB []byte) (ad AuthData, aaguid [16]byte, credID []byte, cose types.CoseEC2, err error) {
    ao, err := ParseAttestationObject(attObjB)
    if err != nil {
        return ad, aaguid, nil, cose, err
    }
    if ao.Fmt != "none" {
        return ad, aaguid, nil, cose, ErrAttestationFormat
    }
    // Parse AD header
    ad, remainder, err := ParseAuthData(ao.AuthData)
    if err != nil {
        return ad, aaguid, nil, cose, err
    }
    if (ad.Flags & FlagAT) == 0 {
        return ad, aaguid, nil, cose, ErrAttestedDataMissing
    }
    // Parse attested cred data
    aaguid, credID, coseRaw, _, err := ParseAttestedCredentialData(remainder)
    if err != nil {
        return ad, aaguid, nil, cose, err
    }
    // Decode COSE EC2
    cose, err = ParseCOSEKeyEC2(coseRaw)
    if err != nil {
        return ad, aaguid, nil, cose, err
    }
    return ad, aaguid, credID, cose, nil
}

