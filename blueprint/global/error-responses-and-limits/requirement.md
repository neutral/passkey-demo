# R-ERR — Error Responses and Limits

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Provide clear error responses and apply basic rate/size limits for inputs and signing payloads.

## Depends On
- R-PLAT-2 (backend)

## Scope
- In-scope: HTTP status conventions; JSON error body; request body size limits; basic rate limit hooks.
- Out-of-scope: full-featured API gateway or WAF.

## Acceptance Criteria
- Uses: 400 invalid input/verification fail; 401/403 auth/origin issues; 409 session/nonce violations; 413 payload too large; 429 rate limited; 5xx server errors.
- Error messages are concise and avoid leaking sensitive internals.
 - Enforce TTL 5 minutes for registration/login/tx sessions; body ≤ 64 KB; message ≤ 1 KB; rate ~10 req/min per IP for auth endpoints.

## Flows
- Applies across all endpoints.

## Interfaces
- Standard JSON error envelope.

## Risks
- Overly terse diagnostics; inconsistent mapping of driver errors.

Refs: goal simple-ui-and-storage; decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; requirement R-ERR
