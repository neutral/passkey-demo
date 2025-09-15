# R-FLOW-LOGIN — Login Flow

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: fr

## Description
- Client fetches assertion options from server, invokes `navigator.credentials.get`, submits assertion for verification. Server checks UV, rpIdHash, origin, and strictly increasing `signCount`.

## Depends On
- R-PLAT-1, R-PLAT-2, R-PLAT-3
- R-ID-KEY, R-SEC-UV

## Scope
- In-scope: login/options + finish endpoints; challenge/session tracking; session establishment on success.
- Out-of-scope: MFA, recovery flows.

## Acceptance Criteria
- `/authn/passkey/login/options` returns SimpleWebAuthn-compatible JSON with `userVerification: "required"`, includes `login_session_id` (≥128-bit entropy) and `expires_at` (5 minute TTL), and logs the issuance.
- `/authn/passkey/login/finish` verifies assertion, ensures UV flag and increasing `signCount`, and establishes session.

## Flows
- Login: blueprint/_user-flows/login.md

## Interfaces
- POST /authn/passkey/login/options
- POST /authn/passkey/login/finish

## Risks
- SignCount rollback; origin misconfiguration; allowCredentials filtering.

Refs: goal passkey-registration-login-uv; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-FLOW-LOGIN
