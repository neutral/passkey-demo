# Purpose
Limit request body sizes for JSON endpoints to prevent resource exhaustion.

# Key Logic
- Wraps handler with `http.MaxBytesReader` at a configured byte size.
- If `Content-Length` exceeds the limit, returns 413 with a JSON error envelope `{code:"ERR_PAYLOAD_TOO_LARGE",error}`.
- If the body exceeds the limit while reading, `MaxBytesReader` triggers a 413; response body may be non-enveloped in that edge path.

# Interactions
- Applied to `/authn/*` and `/tx/*` endpoints via the centralized router.

# Refs
Refs: requirement R-ERR; requirement R-PLAT-2
