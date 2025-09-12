# Purpose
Attach a stable request correlation ID to each incoming HTTP request. The ID is exposed in `X-Request-ID` and added to context for downstream logging and error envelopes. It also prepends a `correlation_id` attribute to the context via a context-aware slog handler.

# Key Logic
- Accepts `X-Request-ID` if provided; otherwise generates base64url ID from 12 random bytes.
- Stores the ID in context under an unexported key; exposes `FromContext(ctx)` to retrieve it.

# Interactions
- Error envelope can include `correlation_id` by reading request ID from context.
- The logger is wrapped with `slog-context`, and the middleware prepends `correlation_id` so handlers that use `InfoContext` automatically include it.
- Intended to be wrapped outermost so all handlers share the same ID.

# Refs
Refs: requirement R-ERR; decision request-id-and-slog-json
