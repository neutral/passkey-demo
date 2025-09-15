# R-FLOW-REG — Registration Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Server-supplied options for WebAuthn registration; client calls `navigator.credentials.create` and submits result for verification and storage.

## Interfaces
- POST /authn/passkey/registration/options → `PublicKeyCredentialCreationOptionsJSON` plus `reg_session_id` (24 char base64url) and `expires_at` (epoch seconds) returned by Node server for diagnostics; consumed by `@simplewebauthn/browser`.
- POST /authn/passkey/registration/finish → submit `RegistrationResponseJSON`; server verifies and creates account+credential

## Data / Models
- accounts, credentials tables as per DB schema.

## Algorithms
- Generate random challenge via `@simplewebauthn/server` with rp `{ id: RP_ID, name: 'Passkey Demo' }`, user placeholder (base64url 32 bytes), `pubKeyCredParams=[ES256]`, `authenticatorSelection={ residentKey: 'required', requireResidentKey: true, userVerification: 'required' }`, `attestation='none'`, `timeout=60000`.
- Persist registration options server-side in `RegistrationSessionStore` as `{ challenge, rpID, origin, expiresAt = now + 300s }`; expose `reg_session_id` to the client.
- Verify attestation response; extract COSE EC2; canonicalize CBOR; store account + credential with `sign_count`.

## Security / Privacy
- Enforce UV; origin and rpId alignment; session binding for options.

## Errors / Observability
- Clear 400/401/409 on invalid inputs, origin/rpId mismatch, dup accounts; log minimal details.
- Successful `/registration/options` requests emit `reg_options` log entries with `correlation_id`, `expires_at`, `rp_id`, `origin`, and `session_id_len`; failures map to JSON envelope `{ code: 'internal_error' }`.

## Testing Strategy
- Unit: `node-server/test/reg-options.test.js` stubs deterministic dependencies to assert policy flags, TTL persistence, collision handling, and error envelopes.
- E2E create passkey on platform authenticator via `@simplewebauthn/browser`; validate DB rows and sign_count initialization; page tests validate posted `RegistrationResponseJSON` + `reg_session_id`.

## Open Questions
- None for demo scope.

Refs: decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; goal passkey-registration-login-uv; requirement R-FLOW-REG
