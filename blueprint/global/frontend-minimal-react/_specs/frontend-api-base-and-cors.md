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

## What to Avoid
- Do not use a Vite dev proxy. Do not set `VITE_API_BASE` to a relative path like `/api` (breaks `new URL(path, API_BASE)`).
- Do not rely on localStorage/sessionStorage for auth; use the server session cookie.

## Best Practices
- Centralize API base in `web/src/config.ts`; avoid scattering base URLs.
- Keep ports stable (`strictPort: true` in `vite.config.ts`).
- Prefer small helpers for base64url/UTF‑8 conversions and WebAuthn builders; avoid heavy deps.
- In tests, start both Vite and the Go backend using Playwright `webServer` entries.

## Expectations by Environment
- Dev: `API_BASE=http://localhost:8080`, frontend at `http://localhost:5173`. CORS permitted, cookies accepted.
- CI (Playwright): same ports; Playwright launches both servers.
- Prod: not covered here; proxy/CDN decisions belong to deployment.

## Refs
Refs: requirement R-PLAT-1; requirement R-OPS-DEV; decision webauthn-corrections-and-standardizations; spec cors-usage-explainer; spec session-cookies-usage-explainer

