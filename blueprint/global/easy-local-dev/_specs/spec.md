# R-OPS-DEV — Local Dev Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Local run with SPA dev server (e.g., Vite on :5173) and Go backend; rp.id "localhost"; origin must match exactly.

## Interfaces
- Backend origin allowlist; CORS (if used) aligned to dev origin.

## Data / Models
- None specific beyond normal flow models.

## Algorithms
- Validate `clientDataJSON.origin` against allowlist; set rp.id accordingly in options.
- Cookies: set `HttpOnly; SameSite=Lax; Secure` on HTTPS; for localhost over HTTP, omit `Secure`.

## Security / Privacy
- Secure context in modern browsers permits http://localhost; avoid exposing additional origins in dev.

## Errors / Observability
- Clear messages on origin mismatches; console logs in dev.

## Testing Strategy
- Manual E2E on localhost with platform authenticator.

## Open Questions
- Bundler choice and exact port; CORS details.

Refs: spec spec-a; spec spec-b; goal simple-ui-and-storage; requirement R-OPS-DEV
