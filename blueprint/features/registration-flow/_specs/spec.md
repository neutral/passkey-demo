# R-FLOW-REG — Registration Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Server-supplied options for WebAuthn registration; client calls `navigator.credentials.create` and submits result for verification and storage.

## Interfaces
- POST /authn/passkey/registration/options → `PublicKeyCredentialCreationOptions`
- POST /authn/passkey/registration/finish → verifies and creates account+credential

## Data / Models
- accounts, credentials tables as per DB schema.

## Algorithms
- Generate random challenge; set `attestation: "none"`, `residentKey: "required"`, `userVerification: "required"`; verify attestation; extract COSE EC2; canonicalize CBOR; store account + credential with `sign_count`.

## Security / Privacy
- Enforce UV; origin and rpId alignment; session binding for options.

## Errors / Observability
- Clear 400/401/409 on invalid inputs, origin/rpId mismatch, dup accounts; log minimal details.

## Testing Strategy
- E2E create passkey on platform authenticator; validate DB rows and sign_count initialization.

## Open Questions
- None for demo scope.

Refs: decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; goal passkey-registration-login-uv; requirement R-FLOW-REG
