package http

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestBodyLimitMiddleware(t *testing.T) {
    h := BodyLimitMiddleware(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Read the body fully
        buf := new(bytes.Buffer)
        _, _ = buf.ReadFrom(r.Body)
        w.WriteHeader(http.StatusOK)
    }))
    // Within limit
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/", strings.NewReader("12345"))
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK { t.Fatalf("within limit code=%d", rr.Code) }
    // Exceed limit (11 bytes)
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/", strings.NewReader("12345678901"))
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusRequestEntityTooLarge { t.Fatalf("exceed code=%d", rr.Code) }
}

