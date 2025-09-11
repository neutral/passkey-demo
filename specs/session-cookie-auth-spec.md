# Session Cookie Auth — Spec

## Purpose

Document how the backend uses an HttpOnly session cookie (`sid`) as the authentication mechanism across registration, login, and transaction-signing flows. Define behavior, lifecycle, security properties, and error handling to ensure consistent implementation and testing.

## Overview

- Authentication uses a server-managed session identified by an HttpOnly cookie `sid`.
- The session binds a browser to an account key (COSE EC2) for a short, fixed lifetime.
- All protected endpoints rely on middleware (`SessionMiddleware`) to establish auth context from `sid` and never read cookies directly in handlers.
- CORS is configured for credentialed requests; frontends send cookies via `credentials: 'include'`.

## Cookie Properties

- Name: `sid`
- Scope: path `/`
- HttpOnly: true (not accessible to JavaScript)
- SameSite: `Lax`
- Secure: true when `Origin` is HTTPS; false on `http://localhost` for development
- Transport: attached automatically by the browser with credentialed CORS requests

## Session Lifecycle

1. Issue: On successful login finish, the server inserts a row into `sessions(session_id, acct_cbor, expires_at, created_at)` and sets `sid`.
2. Validation: `SessionMiddleware` reads `sid`, loads `{acct_cbor, expires_at}`, verifies not expired, optionally refreshes expiry (see middleware docs), and places `acct_cbor` + `credential_id` allowlist in request context.
3. Use: Handlers for protected resources access session via context only.
4. Expiry: Once expired, requests receive `401` with error envelope `{code: "ERR_UNAUTHORIZED"}`.
5. Revocation: Not implemented per-session (demo); users can clear cookies to sign out.

## Flow Integration

- Registration
  - Options: no auth required.
  - Finish: no auth required; on success, persists account + credential. Does not set `sid`.
- Login
  - Options: no auth required.
  - Finish: verifies assertion, enforces signCount policy; creates session row and sets `sid` cookie.
- Transaction Signing
  - Options (POST `/tx/signing/options`): requires session; responds with allowlisted `credential_id`s for the account and creates a transient tx session.
  - Finish (POST `/tx/signing/finish`): requires session; verifies assertion and persists transaction.
- Account Key
  - GET `/me/account_key`: requires session; returns the account’s COSE public key.
- List Transactions
  - GET `/tx/list`: requires session; returns transactions for the authenticated account.

## CORS and Credentials

- CORS allowlist uses exact matches (`cfg.Origin` ∪ `OriginAllowlist`).
- `Access-Control-Allow-Credentials: true` for allowed origins.
- Frontend must call `fetch(..., { mode: 'cors', credentials: 'include' })` so the `sid` cookie is sent.

## Error Handling

- Missing/invalid/expired `sid` → `401` with JSON envelope: `{ code: "ERR_UNAUTHORIZED", error: "unauthorized", correlation_id?: string }`.
- Handlers never fall back to cookie→DB lookups themselves; the middleware is the sole auth source.

## Security Considerations

- XSS: `HttpOnly` mitigates token exfiltration; maintain strict CSP and input sanitization.
- CSRF: Credentialed CORS with strict origin allowlist + non-GET actions requiring JSON bodies reduces CSRF risk. For production, consider an anti-CSRF strategy if multi-origin access is needed.
- Session Fixation: On login, always mint a new `sid`.
- Transport Security: Set `Secure` when `Origin` is HTTPS; restrict origins in production.
- Logging: Never log raw `sid`; use hashed identifiers via `HashID` in logs.

## Observability

- Request ID middleware attaches `X-Request-ID`; error envelopes may include `correlation_id` for cross-tier tracing.

## Testing Notes

- Server tests wrap endpoints with `SessionMiddleware` when exercising protected routes.
- Frontend integration tests must set `credentials: 'include'` and use the configured `Origin`.

## Alternatives Considered (Summary)

- Bearer access tokens (no cookies): avoids CSRF but increases XSS exposure; would require significant client changes and token storage strategy. Cookies chosen for simplicity and stronger default protections.

## References

- Middleware: `server/internal/http/session.go`
- Router wiring: `server/internal/app/router.go`
- Cookie issuance: `server/internal/webauthn/login_finish.go` (`setSessionCookie`)
- Protected handlers using session context: `server/internal/tx/*`, `server/internal/me/account_key.go`

## Refs
Refs: goal simple-ui-and-storage; goal passkey-registration-login-uv; goal transaction-content-signing; requirement R-ERR; decision request-id-and-slog-json; decision encoding-and-ceremony-guardrails

