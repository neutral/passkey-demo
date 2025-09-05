# R-OPS-DEV — Dev Servers and Origins Spec

## Metadata
- Status: Draft
- Date: 2025-09-05
- Owners: passkey-demo maintainers

## Purpose
- Document how the local development environment serves the frontend and backend, how requests flow between them, and how this aligns with our CORS and cookie policies. Provide a canonical reference for ports, URLs, and test tooling.

## Overview
- Frontend is served by a Vite dev server at `http://localhost:5173` in development. Backend runs at `http://localhost:8080`.
- The frontend never uses a Vite proxy. All backend calls are made via absolute URLs built from `API_BASE` defined in `web/src/config.ts`.
- Cross-origin requests are allowed by the backend CORS middleware (Step 26) for the dev origin. Session cookies use `HttpOnly` and `SameSite=Lax`; `Secure` is omitted on `http://localhost` for dev.
- Playwright UI tests launch the Vite dev server to serve static assets/modules and do not interact with the backend unless an E2E spec requires it.

## Interfaces
- Frontend:
  - Dev server: `http://localhost:5173` (no proxy). Config: `web/vite.config.ts` with `strictPort: true`.
  - API base: `API_BASE` in `web/src/config.ts` (default `http://localhost:8080`). Helper `apiUrl(path)` builds absolute URLs.
- Backend:
  - API server: `http://localhost:8080`.
  - CORS allowlist: `ORIGIN=http://localhost:5173` (and allowlist union if configured) — see `server/internal/http/cors.go`.
  - Health: `GET /health` for checks and tests.
- Tests:
  - Playwright config: `web/playwright.config.ts` starts the dev server and runs Chromium headless.
  - Test files live under `web/tests/`.

## Data / Models
- None beyond the application’s flows. Dev config is environmental: `RP_ID`, `ORIGIN`, `PORT`, and Vite server port.

## Process
- Start backend:
  - `RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./server/cmd/api`
- Start frontend:
  - `npm -C web run dev` (serves assets at `http://localhost:5173`)
- Frontend makes API calls as `fetch(apiUrl('/authn/passkey/login/options'), { credentials: 'include' })`.
- Cookies:
  - `sid` cookie is `HttpOnly; SameSite=Lax`. `Secure=true` only for HTTPS origins; for `http://localhost` dev, `Secure=false`.
- CORS:
  - Backend sets `Access-Control-Allow-Origin: http://localhost:5173` and `Access-Control-Allow-Credentials: true` for allowed origins; `Vary: Origin` is set. Preflight handled by middleware.
- Playwright:
  - Uses Vite dev server to load frontend modules; does not use any proxy. E2E specs that hit the backend must use absolute URLs and may start the backend within the test environment or expect it running.

## Security / Privacy
- Dev environment reduces cookie `Secure` requirement only for `http://localhost` to enable testing. UV and RP/Origin validation policies remain unchanged.
- No wildcard CORS; exact-origin allowlist only.

## Errors / Observability
- On CORS misconfig: preflight will return 403; normal requests will miss `Access-Control-Allow-Origin` and be blocked by the browser.
- On origin mismatch: server returns 403 with error envelope in hardening steps.

## Testing Strategy
- Smoke checks:
  - Backend `/health` responds 200 under correct env.
  - Frontend loads at `:5173` and Playwright test `ui.spec.ts` passes.
- Optional E2E: write Playwright specs that orchestrate register→login→sign using absolute URLs and `credentials: 'include'`.

## Open Questions
- Whether to provide a `make run` target that concurrently runs backend and Vite dev server with env injection.

## Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-1; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; spec session-cookies-usage-explainer; spec cors-usage-explainer; goal simple-ui-and-storage

