package http

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestRateLimiter_BasicAndExceed(t *testing.T) {
    rl := NewRateLimiter(2, 0) // capacity 2, no refill
    // allow two
    if !rl.Allow("1.2.3.4", time.Now()) { t.Fatal("first should allow") }
    if !rl.Allow("1.2.3.4", time.Now()) { t.Fatal("second should allow") }
    if rl.Allow("1.2.3.4", time.Now()) { t.Fatal("third should block") }
}

func TestRateLimitMiddleware_PerIPIsolation(t *testing.T) {
    rl := NewRateLimiter(1, 0) // allow 1 each
    h := RateLimitMiddleware(rl, func(r *http.Request) string { return r.RemoteAddr })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    // First IP
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.RemoteAddr = "1.1.1.1:1234"
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("ip1 first code=%d", rr.Code) }
    // Second IP should also allow first request
    rr = httptest.NewRecorder(); req = httptest.NewRequest("GET", "/", nil); req.RemoteAddr = "2.2.2.2:1234"
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("ip2 first code=%d", rr.Code) }
    // First IP second request should be 429
    rr = httptest.NewRecorder(); req = httptest.NewRequest("GET", "/", nil); req.RemoteAddr = "1.1.1.1:1234"
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusTooManyRequests { t.Fatalf("ip1 second code=%d", rr.Code) }
}

func TestRateLimiter_Refill(t *testing.T) {
    rl := NewRateLimiter(1, 50) // 50 tokens/sec
    if !rl.Allow("3.3.3.3", time.Now()) { t.Fatal("should allow first") }
    if rl.Allow("3.3.3.3", time.Now()) { t.Fatal("should block second immediately") }
    time.Sleep(30 * time.Millisecond) // ~1.5 tokens
    if !rl.Allow("3.3.3.3", time.Now()) { t.Fatal("should allow after refill") }
}

