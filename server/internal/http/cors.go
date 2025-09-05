package http

import (
    "net/http"
    "strings"

    cfgpkg "github.com/neutral/passkey-demo/internal/config"
)

// CORSMiddleware returns a middleware that enables credentialed CORS for a
// configured allowlist of origins. It handles preflight requests and sets
// appropriate headers for actual requests when the Origin is allowed.
//
// Policy:
// - Exact-match allowlist: {cfg.Origin} ∪ cfg.OriginAllowlist
// - Methods: GET, POST, OPTIONS
// - Headers: at minimum Content-Type; if Access-Control-Request-Headers is
//   present, echo the intersection with a small allowlist (currently Content-Type).
// - Credentials allowed for allowed origins; Vary: Origin added on allowed responses.
func CORSMiddleware(cfg *cfgpkg.Config) func(http.Handler) http.Handler {
    // Build an exact-match allowlist
    allowed := map[string]struct{}{}
    allowed[cfg.Origin] = struct{}{}
    for _, o := range cfg.OriginAllowlist {
        allowed[o] = struct{}{}
    }

    allowHeadersDefault := "Content-Type"
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            if origin == "" {
                // Non-CORS request
                next.ServeHTTP(w, r)
                return
            }

            // Determine if origin is allowed (exact string match)
            _, ok := allowed[origin]
            if !ok {
                // For preflight, explicitly deny; for others, pass through without CORS headers
                if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
                    w.WriteHeader(http.StatusForbidden)
                    return
                }
                next.ServeHTTP(w, r)
                return
            }

            // Allowed origin
            // Always indicate the specific origin and that credentials are allowed
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Credentials", "true")
            // Vary by Origin to avoid cache poisoning across origins
            vary := w.Header().Get("Vary")
            if vary == "" {
                w.Header().Set("Vary", "Origin")
            } else if !strings.Contains(strings.ToLower(vary), "origin") {
                w.Header().Set("Vary", vary+", Origin")
            }

            // Handle preflight
            if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

                // Compute allowed headers
                reqHdrs := r.Header.Get("Access-Control-Request-Headers")
                allowHdrs := allowHeadersDefault
                if reqHdrs != "" {
                    // Only echo allow-listed headers (case-insensitive contains "content-type")
                    // Normalize and check if requester asked for content-type
                    asked := strings.Split(reqHdrs, ",")
                    var out []string
                    for _, h := range asked {
                        h = strings.TrimSpace(h)
                        if strings.EqualFold(h, "content-type") {
                            out = append(out, "Content-Type")
                        }
                    }
                    if len(out) > 0 {
                        allowHdrs = strings.Join(out, ", ")
                    }
                }
                w.Header().Set("Access-Control-Allow-Headers", allowHdrs)
                w.Header().Set("Access-Control-Max-Age", "600")
                w.WriteHeader(http.StatusNoContent)
                return
            }

            // Actual request; proceed
            next.ServeHTTP(w, r)
        })
    }
}

