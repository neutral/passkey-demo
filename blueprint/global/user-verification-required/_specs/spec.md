# R-SEC-UV — UV Policy Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Enforce `userVerification: "required"` and verify UV flag across all ceremonies.

## Interfaces
- Registration/login/signing options; assertion verification routines.

## Data / Models
- Flags byte in authenticatorData; verification state stored only implicitly (no separate column).

## Algorithms
- Set UV required in options; check UV bit server-side; reject otherwise.

## Security / Privacy
- Prevents unverified assertions; complements rpId/origin checks.

## Errors / Observability
- 400/401 with clear message when UV missing; minimal logging.

## Testing Strategy
- Attempt ceremonies with UV disabled (if possible) and ensure rejection; normal path succeeds.

## Open Questions
- None.

Refs: decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; goal passkey-registration-login-uv; requirement R-SEC-UV
