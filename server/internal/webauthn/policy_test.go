package webauthn

import (
    "crypto/sha256"
    "testing"
)

func rpHash(s string) [32]byte { return sha256.Sum256([]byte(s)) }

func TestCheckRpIdHash_HappyAndMismatch(t *testing.T) {
    ad := rpHash("example.com")
    if err := CheckRpIdHash(ad, "example.com"); err != nil {
        t.Fatalf("happy rpId: %v", err)
    }
    if err := CheckRpIdHash(ad, "example.org"); err == nil {
        t.Fatalf("expected mismatch error")
    }
}

func TestCheckRpIdHash_InvalidForms(t *testing.T) {
    ad := rpHash("example.com")
    bads := []string{"https://example.com", "example.com/", "", "ex ample.com"}
    for _, b := range bads {
        if err := CheckRpIdHash(ad, b); err == nil {
            t.Fatalf("expected invalid rpID error for %q", b)
        }
    }
}

func TestCheckRpIdHash_TrailingDotAndAllow(t *testing.T) {
    ad := rpHash("example.com")
    if err := CheckRpIdHash(ad, "example.com."); err != nil {
        t.Fatalf("trailing dot should pass: %v", err)
    }
    ad2 := rpHash("app.example.com")
    if err := CheckRpIdHashAllowed(ad2, "example.com", []string{"app.example.com"}); err != nil {
        t.Fatalf("allowlist rpID should pass: %v", err)
    }
}

func TestCheckOrigin_Basics(t *testing.T) {
    // expected: https://example.com
    if err := CheckOrigin("https://example.com", "https://example.com", nil, false); err != nil {
        t.Fatalf("expected base pass: %v", err)
    }
    // explicit default port
    if err := CheckOrigin("https://example.com:443", "https://example.com", nil, false); err != nil {
        t.Fatalf("expected pass with :443: %v", err)
    }
    // wrong port
    if err := CheckOrigin("https://example.com:444", "https://example.com", nil, false); err == nil {
        t.Fatalf("expected port mismatch error")
    }
    // wrong scheme
    if err := CheckOrigin("http://example.com", "https://example.com", nil, false); err == nil {
        t.Fatalf("expected scheme error without dev localhost gate")
    }
    // wrong host
    if err := CheckOrigin("https://app.example.com", "https://example.com", nil, false); err == nil {
        t.Fatalf("expected host mismatch error")
    }
    // malformed
    if err := CheckOrigin(":", "https://example.com", nil, false); err == nil {
        t.Fatalf("expected malformed error")
    }
    // trailing dot and case
    if err := CheckOrigin("https://EXAMPLE.COM.", "https://example.com", nil, false); err != nil {
        t.Fatalf("expected pass with case+dot: %v", err)
    }
}

func TestCheckOrigin_AllowlistAndDevLocalhost(t *testing.T) {
    allow := []string{"https://app.example.com", "https://example.net:8443"}
    if err := CheckOrigin("https://app.example.com", "https://example.com", allow, false); err != nil {
        t.Fatalf("allow host should pass: %v", err)
    }
    if err := CheckOrigin("https://example.net:8443", "https://example.com", allow, false); err != nil {
        t.Fatalf("allow host:port should pass: %v", err)
    }
    if err := CheckOrigin("https://example.net", "https://example.com", allow, false); err == nil {
        t.Fatalf("expected port mismatch for allowlist origin without port")
    }
    // dev localhost gate
    if err := CheckOrigin("http://localhost:5173", "https://example.com", nil, true); err != nil {
        t.Fatalf("dev localhost should pass: %v", err)
    }
    if err := CheckOrigin("http://127.0.0.1:5173", "https://example.com", nil, true); err == nil {
        t.Fatalf("loopback IP should not be allowed")
    }
}

func TestCheckOrigin_IPv6LiteralAllowedExact(t *testing.T) {
    allow := []string{"https://[2001:db8::1]:8443"}
    if err := CheckOrigin("https://[2001:db8::1]:8443", "https://example.com", allow, false); err != nil {
        t.Fatalf("ipv6 literal exact allow should pass: %v", err)
    }
    if err := CheckOrigin("https://[2001:db8::1]", "https://example.com", allow, false); err == nil {
        t.Fatalf("missing port should fail")
    }
}

func TestIDNA_RpID_Normalization(t *testing.T) {
    // bücher.ch ⇄ xn--bcher-kva.ch
    ascii := "xn--bcher-kva.ch"
    unicode := "bücher.ch"
    ad := rpHash(ascii)
    if err := CheckRpIdHash(ad, unicode); err != nil {
        t.Fatalf("unicode rpID should match ascii hash: %v", err)
    }
    // café.fr ⇄ xn--caf-dma.fr
    ascii2 := "xn--caf-dma.fr"
    unicode2 := "café.fr"
    ad2 := rpHash(ascii2)
    if err := CheckRpIdHash(ad2, unicode2); err != nil {
        t.Fatalf("unicode rpID should match ascii hash: %v", err)
    }
}

func TestIDNA_Origin_Normalization(t *testing.T) {
    // Allowlist unicode, candidate ascii
    allow := []string{"https://bücher.ch"}
    if err := CheckOrigin("https://xn--bcher-kva.ch", "https://example.com", allow, false); err != nil {
        t.Fatalf("ascii candidate should match unicode allow host: %v", err)
    }
    // Allowlist ascii, candidate unicode
    allow2 := []string{"https://xn--caf-dma.fr"}
    if err := CheckOrigin("https://café.fr", "https://example.com", allow2, false); err != nil {
        t.Fatalf("unicode candidate should match ascii allow host: %v", err)
    }
}
