# Purpose
Implement credentialed CORS with an exact origin allowlist for the SPA dev origin(s). Handles preflight and actual requests and sets `Vary: Origin`.

# Key Logic
- Allowlist = `{ ORIGIN } ∪ ORIGIN_ALLOWLIST` (exact match).
- Preflight (OPTIONS): if origin not allowed → 403 (no `Access-Control-*`). If allowed and requested method ∈ {GET, POST} and requested headers ⊆ {Content-Type} → 204 and headers: `Access-Control-Allow-Origin: <origin>`, `Access-Control-Allow-Methods: GET, POST, OPTIONS`, `Access-Control-Allow-Headers: Content-Type`, `Access-Control-Allow-Credentials: true`, `Access-Control-Max-Age: 600`, and `Vary: Origin`.
- Non-preflight: if allowed, set `Access-Control-Allow-Origin` echo and `Access-Control-Allow-Credentials: true`; set `Vary: Origin` when `Origin` present.

# Interactions
- Applied globally in `src/server.js` before session-protected routes; `/health` remains public but still receives CORS headers when `Origin` is allowed.

# Refs
Refs: requirement R-OPS-DEV; spec cors-usage-explainer

