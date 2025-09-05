package http

import (
    "net/http"
    "net/http/httptest"
    "testing"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
)

func nextOK() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
}

func TestCORS_Preflight_Allowed(t *testing.T) {
    cfg := &cfgpkg.Config{Origin: "http://localhost:5173", OriginAllowlist: []string{}}
    h := CORSMiddleware(cfg)(nextOK())

    req := httptest.NewRequest(http.MethodOptions, "/authn/passkey/login/finish", nil)
    req.Header.Set("Origin", "http://localhost:5173")
    req.Header.Set("Access-Control-Request-Method", http.MethodPost)
    req.Header.Set("Access-Control-Request-Headers", "content-type")
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusNoContent {
        t.Fatalf("status=%d", rr.Code)
    }
    if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
        t.Fatalf("ACAO=%q", got)
    }
    if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" || got == "*" {
        t.Fatalf("ACAM invalid=%q", got)
    }
    if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
        t.Fatalf("ACAH missing")
    }
    if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
        t.Fatalf("ACAC=%q", got)
    }
    if got := rr.Header().Get("Vary"); got == "" {
        t.Fatalf("Vary missing")
    }
}

func TestCORS_Simple_Allowed(t *testing.T) {
    cfg := &cfgpkg.Config{Origin: "http://localhost:5173"}
    h := CORSMiddleware(cfg)(nextOK())

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    req.Header.Set("Origin", "http://localhost:5173")
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("status=%d", rr.Code)
    }
    if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
        t.Fatalf("ACAO=%q", got)
    }
    if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
        t.Fatalf("ACAC=%q", got)
    }
}

func TestCORS_Preflight_Disallowed(t *testing.T) {
    cfg := &cfgpkg.Config{Origin: "http://localhost:5173"}
    h := CORSMiddleware(cfg)(nextOK())

    req := httptest.NewRequest(http.MethodOptions, "/authn/passkey/login/finish", nil)
    req.Header.Set("Origin", "http://evil.local")
    req.Header.Set("Access-Control-Request-Method", http.MethodPost)
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusForbidden {
        t.Fatalf("status=%d", rr.Code)
    }
    if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
        t.Fatalf("ACAO present: %q", got)
    }
}

func TestCORS_Simple_Disallowed_NoHeaders(t *testing.T) {
    cfg := &cfgpkg.Config{Origin: "http://localhost:5173"}
    h := CORSMiddleware(cfg)(nextOK())

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    req.Header.Set("Origin", "http://evil.local")
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("status=%d", rr.Code)
    }
    if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
        t.Fatalf("ACAO present: %q", got)
    }
}

