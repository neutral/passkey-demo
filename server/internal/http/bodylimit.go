package http

import (
    stdhttp "net/http"
)

// BodyLimitMiddleware caps the readable size of the request body.
// If Content-Length exceeds maxBytes, it immediately returns 413.
// Otherwise it wraps the body with MaxBytesReader to enforce the limit during reads.
func BodyLimitMiddleware(maxBytes int64) func(stdhttp.Handler) stdhttp.Handler {
    return func(next stdhttp.Handler) stdhttp.Handler {
        return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
            if maxBytes > 0 && r.ContentLength > maxBytes {
                stdhttp.Error(w, "payload too large", stdhttp.StatusRequestEntityTooLarge)
                return
            }
            if r.Body != nil && maxBytes > 0 {
                r.Body = stdhttp.MaxBytesReader(w, r.Body, maxBytes)
            }
            next.ServeHTTP(w, r)
        })
    }
}
