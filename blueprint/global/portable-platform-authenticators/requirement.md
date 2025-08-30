# R-PORTABLE — Portable Platform Authenticators (Modern Browsers)

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Support running in modern browsers (e.g., Safari/Chrome on macOS) with platform authenticators (Touch ID/Face ID). Ensure flows work via WebAuthn with required policies.

## Depends On
- R-PLAT-1 (frontend), R-PLAT-2 (backend)
- R-SEC-UV, R-FLOW-REG, R-FLOW-LOGIN, R-FLOW-SIGN

## Scope
- In-scope: platform authenticator compatibility; discoverable credentials to enable username-less login; origin/rpId correctness.
- Out-of-scope: cross-platform hardware key quirks beyond demo testing.

## Acceptance Criteria
- Registration, login, and signing complete successfully on at least Safari and Chrome with platform authenticators on macOS.
- Resident credential (`residentKey: "required"`) works for username-less login.

## Flows
- Applies to all ceremonies.

## Interfaces
- Standard WebAuthn calls; no vendor-specific APIs.

## Risks
- Browser-specific behaviors and prompts; platform restrictions in enterprise devices.

Refs: goal passkey-registration-login-uv; goal webauthn-policy-defaults; goal ui-simplicity-two-buttons; spec spec-a; spec spec-b; requirement R-PORTABLE
