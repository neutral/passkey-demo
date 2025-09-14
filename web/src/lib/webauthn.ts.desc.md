# Purpose
Adapters and helpers for WebAuthn flows using `@simplewebauthn/browser`. Converts the Go server’s option shapes into the library’s `*OptionsJSON` and normalizes browser errors for UI toasts. Also retains `buildTxFinish` for Dashboard signing until the Node server emits library JSON directly.

# Key Logic
- `toCreationOptionsJSON` maps server fields into `PublicKeyCredentialCreationOptionsJSON` (challenge stays base64url string), sets `attestation: 'none'`, `residentKey: 'required'`, `userVerification: 'required'`, and `pubKeyCredParams=[{ type:'public-key', alg:-7 }]`. Generates a random 32‑byte base64url `user.id`.
- `toRequestOptionsJSON` maps into `PublicKeyCredentialRequestOptionsJSON` with `userVerification: 'required'`, optional `rpId`, and `allowCredentials[]` only when non‑empty (discoverable credentials otherwise).
- `mapDomException` converts thrown `DOMException` (e.g., `NotAllowedError`) into a simple `{title,detail}` for `normalizeError`.
- `buildTxFinish` mirrors login finish but includes `tx_session_id` instead; encodes the same binary fields in base64url.

# Interactions
- `Register.tsx` and `Login.tsx` use adapters + `@simplewebauthn/browser` (`startRegistration`/`startAuthentication`) and then POST the library’s JSON responses with the server session ids.
- Tested via Playwright: adapter unit tests, POST‑body checks, and Chromium E2E with Virtual Authenticator. Dashboard still uses `buildTxFinish` for signing.

# Refs
Refs: requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec frontend-api-base-and-cors
