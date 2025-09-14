# Purpose
Write a consistent JSON error envelope `{ code, error, correlation_id? }` with a given HTTP status.

# Key Logic
- `writeError(res, status, code, message, correlationId?)` sets status and `Content-Type: application/json`, and writes the envelope.

# Interactions
- Planned for use by future routes (authn/tx) to standardize error responses.

# Refs
Refs: decision http-error-envelope

