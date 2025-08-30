# R-OPS-DEV — Easy Local Development

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Support local development via localhost origins with a single backend binary + SPA dev server. Align rp.id and origin checks for modern browsers.

## Depends On
- R-PLAT-1 (frontend)
- R-PLAT-2 (backend)

## Scope
- In-scope: dev origin (e.g., http://localhost:5173), rp.id "localhost"; simple build/run commands.
- Out-of-scope: production deployment automation.

## Acceptance Criteria
- Registration, login, and signing succeed on localhost with platform authenticators.
- rp.id and origin exactly match allowlist; CORS configured if needed.

## Flows
- Applies to all flows.

## Interfaces
- Dev server config and backend origin checks.

## Risks
- Origin mismatches (scheme/host/port); CORS misconfig.

Refs: goal simple-ui-and-storage; goal webauthn-policy-defaults; spec spec-a; spec spec-b; requirement R-OPS-DEV
