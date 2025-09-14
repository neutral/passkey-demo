# Purpose
Attach a stable correlation id to each request and response via the `X-Request-ID` header; expose it on `req.id` for logging.

# Key Logic
- Generates a 16-char URL-safe id with `nanoid` when no header is provided; otherwise uses the incoming header value.
- Sets `X-Request-ID` response header for downstream tracing.

# Interactions
- Used by `src/server.js` before HTTP logging to include `req.id` in logs.

# Refs
Refs: decision request-id-and-slog-json

