# R-FLOW-REG — Registration Flow

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: fr

## Description
- Client obtains registration options from the server, calls `navigator.credentials.create`, and submits the result for verification and account/credential creation. Enforce `residentKey: "required"`, `userVerification: "required"`, and `attestation: "none"`.

## Depends On
- R-PLAT-1 (frontend minimal React)
- R-PLAT-2 (single Go backend)
- R-PLAT-3 (SQLite)
- R-ID-KEY (identity model)
- R-SEC-UV (UV required)

## Scope
- In-scope: endpoints, WebAuthn options, attestation none, discoverable credentials, storing account + credential, initial `sign_count`.
- Out-of-scope: enterprise attestation validation/trust chains.

## Acceptance Criteria
- `POST /authn/passkey/registration/options` returns valid `PublicKeyCredentialCreationOptions` with required flags.
- `POST /authn/passkey/registration/finish` verifies attestation, extracts COSE key, stores account and credential, persists `sign_count` and optional `aaguid`.
- UV required is enforced server-side.

## Flows
- Registration: blueprint/_user-flows/registration.md

## Interfaces
- POST /authn/passkey/registration/options
- POST /authn/passkey/registration/finish

## Risks
- RP/origin mismatches; attestation parsing; platform authenticator UX.

Refs: goal passkey-registration-login-uv; goal webauthn-policy-defaults; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-FLOW-REG
