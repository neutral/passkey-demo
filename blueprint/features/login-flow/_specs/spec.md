# R-FLOW-LOGIN — Login Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Server-supplied assertion options; client performs WebAuthn get; server verifies UV, rpIdHash, origin, and `signCount` strictly increases.

## Interfaces
- POST /authn/passkey/login/options → `PublicKeyCredentialRequestOptionsJSON` plus metadata `{ login_session_id, expires_at }` (consumed by `@simplewebauthn/browser`)
- POST /authn/passkey/login/finish → submit `AuthenticationResponseJSON` (fetch with `credentials: 'include'` to accept cookie)

## Data / Models
- credentials (`credential_id`, `acct_cbor_fk`, `sign_count`), accounts.

## Algorithms
- Challenge issuance and session binding; `allowCredentials` may be empty for discoverable credentials; session IDs are ≥128-bit entropy base64url strings, TTL 5 minutes.
- Server now returns the SimpleWebAuthn JSON directly (no legacy wrapping), so the frontend forwards `challenge`, `rpId`, `allowCredentials`, etc., without reshaping.
- Verify assertion over `authenticatorData || SHA-256(clientDataJSON)` with account key; enforce low‑S; update `sign_count`.
- Finish handler persists authenticated sessions: validate login session (single-use), call `verifyAuthenticationResponse`, require UV, ensure `newCounter` strictly increases, insert `sessions` row with 1 hour TTL, set `sid` cookie (`HttpOnly; Path=/; SameSite=Lax; Secure` when https), and log `login_finish` with hashed identifiers.
- Accept optional `userHandle` in finish request; ignore for identity (Model 2).
 - Client uses `@simplewebauthn/browser` via adapters that map server shapes to `*OptionsJSON` and returns `AuthenticationResponseJSON`.

## Security / Privacy
- UV required; origin allowlist; rpIdHash check; secure session cookie.

## Errors / Observability
- 400/401 on verification failures; 409 on session/nonce conflicts (if any); log thumbprint + credential id length.

## Testing Strategy
- E2E login with platform authenticator using `@simplewebauthn/browser`; simulate sign_count progression; page tests validate posted `AuthenticationResponseJSON` + `login_session_id` and cookie handling.

## Open Questions
- None for demo scope.

Refs: decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-FLOW-LOGIN
