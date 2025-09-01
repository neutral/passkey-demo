package webauthn

import (
    "bytes"
    "crypto/sha256"
    "errors"
    "net"
    "net/url"
    "strings"
)

// Sentinel errors for RP ID and Origin policy checks.
var (
    ErrRpIdInvalid       = errors.New("invalid RP ID")
    ErrRpIdHashMismatch  = errors.New("rpIdHash mismatch")
    ErrOriginMalformed   = errors.New("malformed origin URL")
    ErrOriginScheme      = errors.New("origin scheme not allowed")
    ErrOriginHost        = errors.New("origin host not allowed")
    ErrOriginPort        = errors.New("origin port not allowed")
    ErrOriginNotAllowed  = errors.New("origin not allowed")
)

func normalizeHost(h string) string {
    h = strings.ToLower(strings.TrimSpace(h))
    // Accept a single trailing dot
    if strings.HasSuffix(h, ".") {
        h = strings.TrimSuffix(h, ".")
    }
    return h
}

func validateAndNormalizeRpID(rpID string) (string, error) {
    rp := strings.TrimSpace(strings.ToLower(rpID))
    if rp == "" {
        return "", ErrRpIdInvalid
    }
    if strings.Contains(rp, "://") || strings.Contains(rp, "/") || strings.ContainsAny(rp, " \t\n") {
        return "", ErrRpIdInvalid
    }
    // Accept a single trailing dot
    if strings.HasSuffix(rp, ".") {
        rp = strings.TrimSuffix(rp, ".")
    }
    // Disallow IP addresses as RP ID (per spec expectations); allow "localhost" specially
    if rp != "localhost" && net.ParseIP(rp) != nil {
        return "", ErrRpIdInvalid
    }
    return rp, nil
}

// CheckRpIdHash validates the rpID string and compares SHA-256(rpID) to the authenticator's rpIdHash.
func CheckRpIdHash(adRp [32]byte, rpID string) error {
    rp, err := validateAndNormalizeRpID(rpID)
    if err != nil {
        return err
    }
    sum := sha256.Sum256([]byte(rp))
    if !bytes.Equal(sum[:], adRp[:]) {
        return ErrRpIdHashMismatch
    }
    return nil
}

// CheckRpIdHashAllowed passes if any of the provided RP IDs matches the rpIdHash.
func CheckRpIdHashAllowed(adRp [32]byte, primary string, allow []string) error {
    // Try primary first
    if err := CheckRpIdHash(adRp, primary); err == nil {
        return nil
    }
    // Try allowlist
    for _, rp := range allow {
        if err := CheckRpIdHash(adRp, rp); err == nil {
            return nil
        }
    }
    return ErrRpIdHashMismatch
}

func defaultPortForScheme(scheme string) string {
    switch scheme {
    case "http":
        return "80"
    case "https":
        return "443"
    default:
        return ""
    }
}

type parsedOrigin struct {
    scheme string
    host   string
    port   string
}

func parseAndNormalizeOrigin(s string) (parsedOrigin, error) {
    u, err := url.Parse(s)
    if err != nil {
        return parsedOrigin{}, ErrOriginMalformed
    }
    if u.Scheme == "" || u.Host == "" {
        return parsedOrigin{}, ErrOriginMalformed
    }
    host := normalizeHost(u.Hostname())
    port := u.Port()
    if port == "" {
        port = defaultPortForScheme(strings.ToLower(u.Scheme))
    }
    return parsedOrigin{scheme: strings.ToLower(u.Scheme), host: host, port: port}, nil
}

// CheckOrigin validates that the provided origin is allowed given an expected origin and optional allowlist.
// devLocalhostOK permits http://localhost:<port> when true.
func CheckOrigin(origin string, expected string, allow []string, devLocalhostOK bool) error {
    cand, err := parseAndNormalizeOrigin(strings.TrimSpace(origin))
    if err != nil {
        return err
    }
    // Scheme policy: require https; allow http only for localhost when devLocalhostOK.
    if cand.scheme != "https" {
        if !(devLocalhostOK && cand.scheme == "http" && cand.host == "localhost") {
            return ErrOriginScheme
        }
    }
    // Dev localhost shortcut: allow http://localhost:<port> when enabled.
    if devLocalhostOK && cand.scheme == "http" && cand.host == "localhost" {
        return nil
    }

    // Build allowed set
    allowed := append([]string{expected}, allow...)
    var hostMatch, schemeMatch bool
    for _, a := range allowed {
        ao, err := parseAndNormalizeOrigin(a)
        if err != nil {
            // skip malformed allow entries
            continue
        }
        if ao.host == cand.host {
            hostMatch = true
        }
        if ao.scheme == cand.scheme {
            schemeMatch = true
        }
        if ao.scheme == cand.scheme && ao.host == cand.host && ao.port == cand.port {
            return nil
        }
    }
    // Granular error classification
    if !hostMatch {
        return ErrOriginHost
    }
    if !schemeMatch {
        return ErrOriginScheme
    }
    return ErrOriginPort
}
