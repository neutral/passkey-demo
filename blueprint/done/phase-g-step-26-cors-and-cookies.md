### Step 26 — CORS & cookies (Done: 2025-09-05)

Verification notes
- Unit tests: `cd server && go test ./...` → all green.
- Manual checks: allowed origin preflight and simple requests return expected CORS headers; disallowed origin preflight returns 403.

26. **CORS & cookies**

    - Scope: Enable cross-origin requests from the SPA dev origin and ensure session cookies behave correctly in dev and https.

      - Allow `http://localhost:5173` via explicit CORS headers; set `Access-Control-Allow-Credentials: true`.
      - Cookies: `sid` uses `HttpOnly`, `SameSite=Lax`, and `Secure=true` only when `ORIGIN` is https; for localhost http dev, `Secure=false`.

    - Source to add/modify:

      - Add: `server/internal/http/cors.go` — CORS middleware with allowlist using `cfg.Origin` ∪ `cfg.OriginAllowlist`; handles preflight.
      - Add: `server/internal/http/cors.go.desc.md` — High-level description of CORS middleware behavior and policies.
      - Modify: `server/cmd/api/main.go` — Wrap mux with CORS middleware (import as `httpx "github.com/neutral/passkey-demo/internal/http"` to avoid stdlib conflict); short-circuit `OPTIONS` using middleware; apply to all routes.
      - Update: `server/cmd/api/main.go.desc.md` — Mention CORS, middleware layering, and preflight behavior.
      - No change expected (confirm in tests): `server/internal/webauthn/login_finish.go` already sets `sid` cookie with `HttpOnly`, `SameSite=Lax`, `Secure` based on scheme.
      - Update if needed: `server/internal/webauthn/login_finish.go.desc.md` — Ensure cookie attributes and dev `Secure` exception are documented.
      - Documentation: link non‑dev explainers from relevant desc files and this step’s Refs.

    - Description files:

      - Create: `server/internal/http/cors.go.desc.md` — Purpose, headers set, preflight logic, and allowlist usage. Refs included.
      - Update: `server/cmd/api/main.go.desc.md` — Add CORS middleware wiring details and note `Vary: Origin`; include Refs to CORS explainer.
      - Update (if text missing): `server/internal/webauthn/login_finish.go.desc.md` — Clarify `SameSite=Lax` and `Secure` conditional on https origin; include Refs to Session Cookies explainer.

    - Request/response shape (CORS):

      - Preflight request: `OPTIONS /<path>` with headers `Origin`, `Access-Control-Request-Method`, `Access-Control-Request-Headers` (e.g., `content-type`).
      - Preflight response (allowed origin):
        - `Access-Control-Allow-Origin: <exact origin>`
        - `Access-Control-Allow-Methods: GET, POST, OPTIONS`
        - `Access-Control-Allow-Headers: Content-Type` (or echo a validated subset of `Access-Control-Request-Headers` if present)
        - `Access-Control-Allow-Credentials: true`
        - `Access-Control-Max-Age: 600`
        - `Vary: Origin`
        - Status: 204 No Content
      - Simple/actual response (allowed origin):
        - `Access-Control-Allow-Origin: <exact origin>`; `Access-Control-Allow-Credentials: true`; `Vary: Origin`.
      - Disallowed origin: omit CORS headers; preflight returns 403.

    - Algorithm:

      - Parse `Origin` from request; if blank, pass-through (no headers).
      - Consider allowed set: `{cfg.Origin} ∪ cfg.OriginAllowlist` (exact match only).
      - If origin is allowed:
        - For `OPTIONS` with `Access-Control-Request-Method`: write preflight headers and `204`; return.
        - For all other methods: set `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials: true`, and `Vary: Origin`; call next.
      - If origin is not allowed:
        - For `OPTIONS` preflight: return `403` without `Access-Control-*`.
        - For normal requests: pass-through without CORS headers (browser blocks).
      - Cookie policy is enforced by existing `setSessionCookie`: compute `Secure = (url.Parse(cfg.Origin).Scheme == "https")`; always `SameSite=Lax`, `HttpOnly=true`.

    - Database interactions: None (CORS is stateless). Session cookie issuance remains in login finish handler; DB unchanged.

    - Policies & limits:

      - Allowlist: exact match only; no wildcards; derive from config (`ORIGIN`, `ORIGIN_ALLOWLIST`).
      - Credentials: always set `Access-Control-Allow-Credentials: true` for allowed origins.
      - Methods: `GET, POST, OPTIONS`. Headers: at minimum `Content-Type` (JSON); additional requested headers may be echoed if validated.
      - Preflight cache: `Access-Control-Max-Age=600` seconds.
      - Headers hygiene: add `Vary: Origin` on allowed responses to avoid cache poisoning.

    - Sequencing:

      - Depends on Step 24 (Session middleware) being available to keep layering consistent; CORS should wrap the entire mux before session/rate middlewares or after? Apply CORS as the outermost wrapper so preflight can short-circuit early.
      - Step 25 (Rate limiting) remains in place; ensure preflight is not rate-limited (handled entirely by CORS middleware before rate limiting).
      - Frontend in later steps must use `fetch(..., { mode: 'cors', credentials: 'include' })` for cookie receipt and sending.

    - Tests:

      - Files to add:
        - `server/internal/http/cors_test.go`
      - Happy path (required):
        - Allowed origin `http://localhost:5173` on `OPTIONS /authn/passkey/login/finish` returns 204 and headers: `Allow-Origin` echo, `Allow-Methods` includes `POST`, `Allow-Headers` includes `Content-Type`, `Allow-Credentials=true`, `Max-Age=600`, `Vary: Origin`.
        - Allowed origin on `GET /health` sets `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials: true`.
      - Cookie attributes (existing code; add assertions):
        - In `server/internal/webauthn/login_finish_test.go`, add subtests asserting `SameSite=Lax` present and `Secure` absent for `ORIGIN=http://localhost:5173`; and `Secure` present for `ORIGIN=https://example.com`.
      - Negative cases:
        - Disallowed origin preflight returns 403 and no `Access-Control-*` headers.
        - Simple request with disallowed origin does not include `Access-Control-Allow-Origin`.
      - Invariants/mapping checks:
        - `Vary: Origin` always present for allowed origins.
        - `Access-Control-Allow-Origin` never `*` when `Allow-Credentials=true`.
      - Commands:
        - `go test ./server/internal/http -run CORS`
        - `go test ./server/internal/webauthn -run LoginFinish` (to include cookie attribute subtests)
        - `go test ./...` should pass after implementation.

    - Verification:

      - Build and run server: `RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./server/cmd/api`.
      - Preflight from allowed origin:
        - Expect 204 and headers as specified.
      - Simple request from allowed origin:
        - `GET /health` returns 200 with `Access-Control-Allow-Origin: http://localhost:5173` and `Access-Control-Allow-Credentials: true`.
      - Fix-forward loop: if tests fail or headers missing, adjust middleware; re-run `go test ./...` until green.
      - User verification commands:

        ```bash
        # 1) Run server
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./server/cmd/api &
        SERVER_PID=$!

        # 2) Preflight check (allowed)
        curl -i -X OPTIONS \
          -H "Origin: http://localhost:5173" \
          -H "Access-Control-Request-Method: POST" \
          -H "Access-Control-Request-Headers: content-type" \
          http://localhost:8080/authn/passkey/login/finish

        # 3) Simple request check (allowed)
        curl -i -H "Origin: http://localhost:5173" http://localhost:8080/health

        # 4) Preflight check (disallowed)
        curl -i -X OPTIONS \
          -H "Origin: http://evil.local" \
          -H "Access-Control-Request-Method: POST" \
          http://localhost:8080/authn/passkey/login/finish

        # 5) Cleanup
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - CORS middleware exists and is wired as outermost layer; preflight for `http://localhost:5173` returns 204 with correct headers; disallowed origin returns 403.
      - All allowed-origin responses include `Access-Control-Allow-Origin: <origin>`, `Access-Control-Allow-Credentials: true`, and `Vary: Origin`.
      - `sid` cookie attributes: `HttpOnly`, `SameSite=Lax`, `Secure=false` for `ORIGIN=http://localhost:5173`; `Secure=true` for https origins. Tests assert this.
      - `go test ./...` passes.

    - Notes:

      - This is a demo-grade CORS policy: exact-origin allowlist only; no wildcards. Keep scope minimal.
      - Preflight is short-circuited and should not be rate-limited; place CORS middleware outermost.
      - SPA must set `credentials: 'include'` on fetch; cookie flow relies on SameSite semantics (localhost is same-site across ports).

    - Refs:
      - Refs: goal simple-ui-and-storage; requirement R-OPS-DEV; requirement R-PLAT-2; requirement R-PLAT-1; requirement R-ERR; decision webauthn-corrections-and-standardizations; spec cors-usage-explainer; spec session-cookies-usage-explainer

