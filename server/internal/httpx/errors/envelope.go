package errors

import (
    "encoding/json"
    "net/http"
    mid "github.com/neutral/passkey-demo/internal/httpx/middleware"
)

// Envelope is the standard JSON error payload for API responses.
// It intentionally avoids leaking sensitive values, while providing a stable code
// and a human-readable message suitable for logs or developer tools.
type Envelope struct {
    Code          string `json:"code"`
    Error         string `json:"error"`
    CorrelationID string `json:"correlation_id,omitempty"`
}

// Predefined error codes used across the API.
const (
    CodeMethodNotAllowed = "ERR_METHOD_NOT_ALLOWED"
    CodeUnauthorized     = "ERR_UNAUTHORIZED"
    CodeForbidden        = "ERR_FORBIDDEN"
    CodeBadRequest       = "ERR_BAD_REQUEST"
    CodeConflict         = "ERR_CONFLICT"
    CodeRateLimit        = "ERR_RATE_LIMIT"
    CodeInternal         = "ERR_INTERNAL"
)

// Write writes a JSON error envelope with the given HTTP status and code.
// The message is included for developer clarity; clients should key off `code`.
func Write(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(Envelope{Code: code, Error: message})
}

// WriteReq is like Write but attempts to include `correlation_id` from
// request context when present.
func WriteReq(w http.ResponseWriter, r *http.Request, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    env := Envelope{Code: code, Error: message}
    if id, ok := mid.FromContext(r.Context()); ok { env.CorrelationID = id }
    _ = json.NewEncoder(w).Encode(env)
}
