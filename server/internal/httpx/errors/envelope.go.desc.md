# Purpose
Provide a standard JSON error envelope for API responses, including a stable `code` and a human-readable message. This enables consistent frontend handling and improves observability.

# Key Logic
- Defines `Envelope{code,error,correlation_id}`.
- `Write(w,status,code,message)` sets `Content-Type` JSON, status code, and serializes the envelope.
- `WriteReq(w,r,status,code,message)` additionally includes `correlation_id` from request context when present.
- Predefines canonical error codes (e.g., `ERR_UNAUTHORIZED`, `ERR_BAD_REQUEST`, `ERR_CONFLICT`).
- Includes `ERR_RATE_LIMIT` (429) and `ERR_PAYLOAD_TOO_LARGE` (413) for middleware responses.

# Interactions
- Used across handlers (`/authn/*`, `/tx/*`) for consistent error responses.
- Includes `correlation_id` sourced from Request ID middleware when present.

# Refs
Refs: requirement R-ERR; decision http-error-envelope; decision encoding-and-ceremony-guardrails; decision request-id-and-slog-json; decision structured-logging-with-slog-guidelines
