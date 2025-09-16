# Purpose
Provide centralized helpers for writing standardized JSON error envelopes (`{code,error,correlation_id?,details?}`) and an Express error handler that maps parser/size failures into those envelopes.

# Key Logic
- `ERROR_CODES` enumerates the canonical codes (`ERR_BAD_REQUEST`, `ERR_UNAUTHORIZED`, `ERR_FORBIDDEN`, `ERR_CONFLICT`, `ERR_PAYLOAD_TOO_LARGE`, `ERR_RATE_LIMIT`, `ERR_INTERNAL`).
- Convenience responders (`respondBadRequest`, `respondUnauthorized`, `respondForbidden`, `respondConflict`, `respondPayloadTooLarge`, `respondRateLimit`, `respondInternalError`) set HTTP status, populate the envelope, and ensure correlation ids/headers (e.g., `Retry-After`).
- `buildExpressErrorHandler()` converts `body-parser` parse/limit errors into the correct envelope so middleware earlier in the stack can throw and still preserve consistency.

# Interactions
- Used across WebAuthn and transaction routes to return policy-aligned error codes.
- Rate limiter/body limit middleware invoke `respondPayloadTooLarge`/`respondRateLimit` so clients always see the same envelope.
- Error handler appended in `server.js` to catch JSON parsing and unexpected errors at the edge.

# Refs
Refs: requirement R-ERR; decision http-error-envelope; decision router-builder-wiring
