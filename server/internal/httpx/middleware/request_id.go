package middleware

import (
    "context"
    "net/http"
    randutil "github.com/neutral/passkey-demo/internal/util/randutil"
    b64 "github.com/neutral/passkey-demo/internal/encoding"
)

type ctxKey string

const requestIDKey ctxKey = "req_id"

// WithRequestID returns a copy of the context with the request id.
func WithRequestID(ctx context.Context, id string) context.Context { return context.WithValue(ctx, requestIDKey, id) }

// FromContext extracts the request id if present.
func FromContext(ctx context.Context) (string, bool) {
    v, ok := ctx.Value(requestIDKey).(string)
    return v, ok
}

// RequestID middleware attaches a correlation id to the request context and response header.
// Accepts an incoming X-Request-ID header; otherwise generates a base64url id.
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            b, _ := randutil.BytesE(12)
            id = b64.Encode(b)
        }
        // Add a weak form of time component to aid log correlation across systems without clocks skewing ordering
        w.Header().Set("X-Request-ID", id)
        r = r.WithContext(WithRequestID(r.Context(), id))
        next.ServeHTTP(w, r)
    })
}
