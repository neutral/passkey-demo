package webauthn

import (
    "encoding/binary"
    "errors"
)

// AuthData holds parsed fields from WebAuthn authenticatorData needed for assertions.
type AuthData struct {
    RpIDHash  [32]byte
    Flags     byte
    SignCount uint32
}

// Authenticator Data flag bits (from WebAuthn spec)
const (
    FlagUP byte = 1 << 0 // User Present
    FlagUV byte = 1 << 2 // User Verified
    FlagAT byte = 1 << 6 // Attested credential data included
    FlagED byte = 1 << 7 // Extension data included
)

// HasUP returns true if the User Present bit is set.
func HasUP(flags byte) bool { return flags&FlagUP != 0 }

// HasUV returns true if the User Verified bit is set.
func HasUV(flags byte) bool { return flags&FlagUV != 0 }

// ParseAuthData parses the fixed header (37 bytes) of authenticatorData and returns
// the parsed fields and the remaining bytes (which may include attested credential
// data and/or extensions for registration flows).
func ParseAuthData(b []byte) (AuthData, []byte, error) {
    var ad AuthData
    if len(b) < 37 {
        return ad, nil, errors.New("authenticatorData too short")
    }
    copy(ad.RpIDHash[:], b[:32])
    ad.Flags = b[32]
    ad.SignCount = binary.BigEndian.Uint32(b[33:37])
    return ad, b[37:], nil
}

