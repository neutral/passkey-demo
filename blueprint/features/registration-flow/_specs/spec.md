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
- Finish: call `verifyRegistrationResponse` with stored challenge/origin/RP ID allowlists, require UV, ensure attestation fmt `none`, canonicalize COSE key (CBOR sorted) before persistence, compute `acct_thumb = SHA-256('ACCTK1' || acct_cbor)`, store account + credential (with `sign_count`, `aaguid`), delete session, and emit `reg_finish` log with hashed identifiers.
- SimpleWebAuthn v10 expects binary `userID` values when generating registration options; the server returns the library-native JSON (including `user.id` as base64url) without reshaping so the browser can forward it directly. See https://simplewebauthn.dev/docs/advanced/server/custom-user-ids for rationale.

## Security / Privacy
- Enforce UV; origin and rpId alignment; session binding for options and finish; treat sessions as single-use with 5-minute TTL.

## Errors / Observability
- Clear 400/401/409 on invalid inputs, origin/rpId mismatch, dup accounts; log minimal details.
- Successful `/registration/options` requests emit `reg_options` log entries with `correlation_id`, `expires_at`, `rp_id`, `origin`, and `session_id_len`; failures map to JSON envelope `{ code: 'internal_error' }`.
- `/registration/finish` emits `reg_finish` logs with `account_thumb_hex`, `credential_id_hash`, `sign_count`; duplicate credentials return 409 `conflict`; expired/invalid sessions return 401 `unauthorized`; policy failures map to 403 `policy_violation` while malformed inputs map to 400 `bad_request`.

## Testing Strategy
- Unit: `node-server/test/reg-options.test.js` stubs deterministic dependencies to assert policy flags, TTL persistence, collision handling, and error envelopes.
- Unit: `node-server/test/reg-finish.test.js` covers verification success, UV/attestation failures, duplicate credential conflicts, session expiry, and malformed requests using in-memory SQLite.
- E2E create passkey on platform authenticator via `@simplewebauthn/browser`; validate DB rows and sign_count initialization; page tests validate posted `RegistrationResponseJSON` + `reg_session_id`.

## Open Questions
- None for demo scope.

Refs: decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; goal passkey-registration-login-uv; requirement R-FLOW-REG
