package http

import (
    "net"
    "net/http"
    "sync"
    "time"
)

// TokenBucket implements a simple token bucket for a single key (e.g., an IP).
type TokenBucket struct {
    capacity int
    tokens   float64
    rate     float64 // tokens per second
    last     time.Time
}

func (b *TokenBucket) allow(now time.Time) bool {
    if b.last.IsZero() {
        b.tokens = float64(b.capacity)
        b.last = now
    }
    elapsed := now.Sub(b.last).Seconds()
    if elapsed > 0 {
        b.tokens += elapsed * b.rate
        if b.tokens > float64(b.capacity) {
            b.tokens = float64(b.capacity)
        }
        b.last = now
    }
    if b.tokens >= 1 {
        b.tokens -= 1
        return true
    }
    return false
}

// RateLimiter holds per-key token buckets.
type RateLimiter struct {
    mu       sync.Mutex
    buckets  map[string]*TokenBucket
    capacity int
    rate     float64
}

// NewRateLimiter creates a limiter with the given bucket capacity and refill rate.
func NewRateLimiter(capacity int, rate float64) *RateLimiter {
    return &RateLimiter{buckets: make(map[string]*TokenBucket), capacity: capacity, rate: rate}
}

// Allow returns true if a request from key should be allowed at given time.
func (rl *RateLimiter) Allow(key string, now time.Time) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    b := rl.buckets[key]
    if b == nil {
        b = &TokenBucket{capacity: rl.capacity, rate: rl.rate, last: now, tokens: float64(rl.capacity)}
        rl.buckets[key] = b
    }
    return b.allow(now)
}

// RemoteIP derives a best-effort remote IP from the request. It does not trust headers.
func RemoteIP(r *http.Request) string {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err == nil && host != "" {
        return host
    }
    return r.RemoteAddr
}

// RateLimitMiddleware wraps a handler and enforces a simple per-key token bucket.
// ipfn is used to extract the key (typically an IP address).
func RateLimitMiddleware(rl *RateLimiter, ipfn func(*http.Request) string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := ipfn(r)
            if !rl.Allow(key, time.Now()) {
                http.Error(w, "too many requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

