# Purpose
Adapters and helpers for WebAuthn flows using `@simplewebauthn/browser`. Converts server option shapes into the library’s `*OptionsJSON` and normalizes browser errors for UI toasts. Registration/login now pass through the Node server’s SimpleWebAuthn-native JSON while keeping `buildTxFinish` for Dashboard signing until the API is migrated.

# Key Logic
- `toCreationOptionsJSON` accepts the Node server’s SimpleWebAuthn JSON, only generating a random fallback `user.id` when the payload omits one.
- `toRequestOptionsJSON` strips metadata (`login_session_id`, `expires_at`) and forwards the remaining SimpleWebAuthn login options verbatim; `toRequestOptions` converts the JSON into native browser types (base64url→ArrayBuffer) for legacy callers.
- `mapDomException` converts thrown `DOMException` (e.g., `NotAllowedError`) into a simple `{title,detail}` for `normalizeError`.
- `buildTxFinish` mirrors login finish but includes `tx_session_id` instead; encodes the same binary fields in base64url.

# Interactions
- `Register.tsx` and `Login.tsx` use adapters + `@simplewebauthn/browser` (`startRegistration`/`startAuthentication`) and then POST the library’s JSON responses with the server session ids.
- Tested via Playwright: adapter unit tests, POST‑body checks, and Chromium E2E with Virtual Authenticator. Dashboard still uses `buildTxFinish` for signing.

# Refs
Refs: requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec frontend-api-base-and-cors
