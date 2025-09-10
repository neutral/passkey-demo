# Purpose
Helpers to build WebAuthn registration and login options from server JSON and to construct the finish payloads by encoding binary fields to base64url.

# Key Logic
- `toCreationOptions` maps server `rp_id` to `rp.id`, decodes `challenge` (base64url→bytes), sets `attestation: 'none'`, `authenticatorSelection.residentKey: 'required'`, `userVerification: 'required'`, and `pubKeyCredParams=[{ type:'public-key', alg:-7 }]`.
- `buildRegFinish` reads `PublicKeyCredential` and encodes `rawId`, `attestationObject`, and `clientDataJSON` to base64url; includes `reg_session_id`.
- `toRequestOptions` decodes `challenge`, sets `userVerification: 'required'`, optionally sets `rpId`, and maps `allow_credentials` (b64url→bytes) to `allowCredentials[]`. If the decoded list is empty, omits `allowCredentials` entirely to enable discoverable credentials.
- `buildLoginFinish` encodes `rawId`, `authenticatorData`, `clientDataJSON`, `signature`, and optional `userHandle`; includes `login_session_id`.
- `buildTxFinish` mirrors login finish but includes `tx_session_id` instead; encodes the same binary fields in base64url.

# Interactions
- Used by `Register.tsx` to fetch options, call `navigator.credentials.create`, and POST finish using absolute URLs via `web/src/config.ts` under CORS (Step 26).
- Used by `Login.tsx` to fetch options, call `navigator.credentials.get`, and POST finish with `credentials: 'include'` to accept cookies under CORS.
- Tested via Playwright: pure builder tests; optional Chromium E2E with Virtual Authenticator. Used by Dashboard for transaction signing flow (options→get→finish).

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec spec-a; spec spec-b
