# Purpose
Enable credentialed CORS for the SPA dev origin and any explicitly allowed origins. Answers preflight requests and sets headers on actual requests so the browser can include the session cookie.

# Key Logic
- Allowlist: exact match against `{cfg.Origin} ∪ cfg.OriginAllowlist`.
- Preflight (OPTIONS): `204 No Content` with `Access-Control-Allow-Origin: <origin>`, `Allow-Methods: GET, POST, OPTIONS`, `Allow-Headers: Content-Type` (or validated subset), `Allow-Credentials: true`, `Max-Age: 600`, and `Vary: Origin`.
- Actual requests: add `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials: true`, and `Vary: Origin`.
- Disallowed origin: preflight → `403` (no CORS headers); actual → pass-through without CORS headers.

# Interactions
- Applied as the outermost middleware in `cmd/api/main.go` so preflights short-circuit before other middlewares.
- Works with session cookies (`sid`) and `credentials: 'include'` from the SPA.

# Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-2; requirement R-PLAT-1; spec cors-usage-explainer

