# Purpose
Register page runs the WebAuthn create() ceremony with the Node backend: request options, build `PublicKeyCredentialCreationOptionsJSON`, invoke `startRegistration`, and forward the library JSON plus `reg_session_id` to `/finish`.

# Key Logic
- `postJson` requests `/authn/passkey/registration/options`, normalizes via `toCreationOptionsJSON`, and feeds it into `@simplewebauthn/browser`.
- On success submits `{ reg_session_id, ...RegistrationResponseJSON }` back to `/finish` and renders the returned `account_thumb_hex`.
- Catches `ApiError` to surface structured strings (`formatApiError`) and maps authenticator DOM exceptions through `mapDomException` for friendly toasts.

# Interactions
- Triggered from `App.tsx`; relies on `apiUrl`/`postJson` so every request hits the Node server with credentials.
- Shares adapters with the unit/Playwright tests (`web/tests/register-*`) and uses `ErrorToast` for UX parity with login/signing flows.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-ERR; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
