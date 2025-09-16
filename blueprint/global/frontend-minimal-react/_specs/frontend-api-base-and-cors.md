# R-PLAT-1 — Frontend API Base and CORS Guidelines

## Metadata
- Status: Draft
- Date: 2025-09-05
- Owners: passkey-demo maintainers

## Purpose
- Capture frontend practices for talking to the backend in local dev and test environments without a proxy. Reduce CORS issues and prevent misconfiguration.

## What Works
- Absolute API URLs via `API_BASE` and `apiUrl(path)` in `web/src/config.ts`.
- CORS allowed origin is `http://localhost:5173`; fetch with `mode: 'cors'` and set `credentials: 'include'` when cookies are required.
- Cookies: `HttpOnly; SameSite=Lax; Secure=false` on `http://localhost` (Secure=true for https origins).
 - Error envelope: backend returns `{code,error,correlation_id?}` on non‑2xx; frontend should parse and surface concise messages.
 - WebAuthn client: use `@simplewebauthn/browser` with `PublicKeyCredential*OptionsJSON` and submit `RegistrationResponseJSON`/`AuthenticationResponseJSON`; adapters may be used to bridge server shapes.

## What to Avoid
- Do not use a Vite dev proxy. Do not set `VITE_API_BASE` to a relative path like `/api` (breaks `new URL(path, API_BASE)`).
- Do not rely on localStorage/sessionStorage for auth; use the server session cookie.

## Best Practices
- Centralize API base in `web/src/config.ts`; avoid scattering base URLs.
- Use the shared API client (`web/src/lib/api.ts`) / helpers (`postJson`, `formatApiError`) so every request sends credentials and surfaces structured errors.
- Keep ports stable (`strictPort: true` in `vite.config.ts`).
 - Prefer small helpers for base64url/UTF‑8 conversions; use `@simplewebauthn/browser` for WebAuthn flows to avoid manual ArrayBuffer transforms.
- In tests, start both Vite and the Node backend (`node-server/src/server.js`) using Playwright `webServer` entries.

## Expectations by Environment
- Dev: `API_BASE=http://localhost:8080`, frontend at `http://localhost:5173`. CORS permitted, cookies accepted.
- CI (Playwright): same ports; Playwright launches both servers.
- Prod: not covered here; proxy/CDN decisions belong to deployment.

## Refs
Refs: requirement R-PLAT-1; requirement R-OPS-DEV; decision webauthn-corrections-and-standardizations; decision http-error-envelope; spec cors-usage-explainer; spec session-cookies-usage-explainer
