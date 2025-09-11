# R-ERR — Error Responses and Limits

## Metadata
- State: Approved
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Provide clear, standardized error responses and apply router-level rate/size limits.

## Depends On
- R-PLAT-2 (backend)

## Scope
- In-scope: HTTP status conventions; JSON error envelope; request body size limits; per-group rate limits.
- Out-of-scope: full-featured API gateway or WAF.

## Acceptance Criteria
- Uses: 400 invalid input/verification fail; 401/403 auth/origin issues; 409 session/nonce violations; 413 payload too large; 429 rate limited; 5xx server errors.
- Error envelope includes `{ code, error, correlation_id? }`; messages avoid leaking sensitive internals.
- Enforce TTL 5 minutes for registration/login/tx sessions; body limit 1 MiB on `/authn/*` and `/tx/*`; burst 20, ~10 rps token bucket per IP.

## Flows
- Applies across all endpoints.

## Interfaces
- Standard JSON error envelope.

## Risks
- Overly terse diagnostics; inconsistent mapping of driver errors.

Refs: goal simple-ui-and-storage; decision encoding-and-ceremony-guardrails; decision http-error-envelope; decision router-builder-wiring; requirement R-ERR
