# Purpose
Validate CORS middleware behavior for allowed/disallowed preflight and allowed actual requests.

# Key Logic
- Preflight allowed: expect 204 and appropriate allow headers (origin echo, credentials true, vary origin).
- Preflight disallowed: expect 403 and no CORS headers.
- Actual allowed: GET /health includes `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials`.

# Refs
Refs: spec cors-usage-explainer; requirement R-OPS-DEV

