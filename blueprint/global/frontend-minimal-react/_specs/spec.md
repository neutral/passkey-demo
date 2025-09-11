# R-PLAT-1 Frontend Minimal React — Technical Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Minimal React SPA implementing register, login, and post-login transaction signing. No external state management; use React hooks and backend session. Binary exchanges use base64url. Client constructs canonical CBOR bundle for signing; server supplies WebAuthn options for all ceremonies.

## Interfaces
- Web UI: SPA with primary actions Register and Login; dashboard shows signed messages list and input box to add a signed message.
- HTTP API (JSON; base64url for binary) consumed by the frontend:
  - POST /authn/passkey/registration/options
  - POST /authn/passkey/registration/finish
  - POST /authn/passkey/login/options
  - POST /authn/passkey/login/finish
  - POST /tx/signing/options
  - POST /tx/signing/finish
  - GET  /tx/list

## Data / Models
- Account identity: COSE EC2 public key (canonical CBOR) from registration.
- Transaction bundle (client-built): { sender_key (COSE EC2 pub), nonce (uint), message (tstr), optional valid_until (uint) } — encoded to canonical CBOR (B).
- Binary fields are base64url in JSON requests/responses.

## Algorithms / Client Logic
- Registration: fetch server options → navigator.credentials.create → submit finish.
- Login: fetch server options → navigator.credentials.get → submit finish.
- Signing: build canonical CBOR bundle (B) → POST /tx/signing/options (server derives challenge = SHA-256("CHALv1" || B)) → navigator.credentials.get → POST /tx/signing/finish.

## Security / Privacy
- User Verification required in all WebAuthn ceremonies; verify UV on server.
- Dev environment: rp.id "localhost"; origin must exactly match, e.g., http://localhost:5173.
- No usernames; passkey-first identity (COSE key) for account model.
- Avoid persisting sensitive data client-side; rely on server session cookie.

## Errors / Observability
- Surface clear error messages for WebAuthn and HTTP failures using the backend error envelope `{code,error,correlation_id?}`.
- Use `web/src/lib/api.ts` to centralize `credentials: 'include'` and envelope parsing into a typed `ApiError`.
- Log minimal diagnostic info in dev console; no PII in logs.

## Testing Strategy
- Manual E2E verification: register → login → sign message → verify message appears in dashboard.
- Spot-check canonical CBOR generation determinism for simple samples.
- Validate origin/rpId alignment and UV enforcement during flows.

## Open Questions
- Confirm bundler/dev server setup for http://localhost:5173 (e.g., Vite) and CORS settings as needed.
- Identify the CBOR and base64url utility libraries used in the frontend.

Refs: requirement frontend-minimal-react; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails; decision http-error-envelope; goal simple-ui-and-storage; goal ui-simplicity-two-buttons
