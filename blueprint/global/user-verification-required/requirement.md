# R-SEC-UV — User Verification Required

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Enforce user verification (biometric/PIN) for all WebAuthn ceremonies (registration, login, signing). Server checks UV flag in authenticatorData and rejects if not set.

## Depends On
- R-PLAT-2 (backend)
- R-PLAT-1 (frontend)

## Scope
- In-scope: set `userVerification: "required"` in options; verify UV flag server-side.
- Out-of-scope: policy exceptions.

## Acceptance Criteria
- Options for all ceremonies include `userVerification: "required"`.
- Server-side verification rejects assertions lacking UV flag.

## Flows
- Applies to registration, login, signing.

## Interfaces
- Creation/Request options builders and assertion verifiers.

## Risks
- Platform differences in prompts; dev confusion if UV blocked by OS settings.

Refs: goal webauthn-policy-defaults; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-SEC-UV
