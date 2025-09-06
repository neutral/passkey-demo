# Purpose
Helpers to build WebAuthn registration options from server JSON and to construct the finish payload by encoding binary fields to base64url.

# Key Logic
- `toCreationOptions` maps server `rp_id` to `rp.id`, decodes `challenge` (base64url→bytes), sets `attestation: 'none'`, `authenticatorSelection.residentKey: 'required'`, `userVerification: 'required'`, and `pubKeyCredParams=[{ type:'public-key', alg:-7 }]`.
- `buildRegFinish` reads `PublicKeyCredential` and encodes `rawId`, `attestationObject`, and `clientDataJSON` to base64url; includes `reg_session_id`.

# Interactions
- Used by `Register.tsx` to fetch options, call `navigator.credentials.create`, and POST finish using absolute URLs via `web/src/config.ts` under CORS (Step 26).
- Tested via Playwright: pure builder tests; optional Chromium E2E with Virtual Authenticator.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec spec-a; spec spec-b

