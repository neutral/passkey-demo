package encoding

import (
    "encoding/base64"
    "fmt"
    "strings"
)

// Encode returns a base64url string without padding for the given bytes.
func Encode(b []byte) string {
    if len(b) == 0 {
        return ""
    }
    return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a base64url string into bytes.
// It is tolerant of presence/absence of '=' padding.
func Decode(s string) ([]byte, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return []byte{}, nil
    }
    // Try unpadded URL-safe first
    if dst, err := base64.RawURLEncoding.DecodeString(s); err == nil {
        return dst, nil
    }
    // Try padded URL-safe
    if dst, err := base64.URLEncoding.DecodeString(s); err == nil {
        return dst, nil
    }
    // If padding seems off, normalize to multiple of 4 and retry
    if rem := len(s) % 4; rem != 0 {
        s2 := s + strings.Repeat("=", 4-rem)
        if dst, err := base64.URLEncoding.DecodeString(s2); err == nil {
            return dst, nil
        }
    }
    return nil, fmt.Errorf("invalid base64url input")
}

