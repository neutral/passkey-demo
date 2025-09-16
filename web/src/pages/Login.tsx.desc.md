# Purpose
Login page drives the WebAuthn get() ceremony via `@simplewebauthn/browser`: consume the Node backend login JSON, execute `startAuthentication`, and submit the assertion with `login_session_id` so the server can mint an HttpOnly session cookie.

# Key Logic
- `postJson` pulls `/authn/passkey/login/options`, `toRequestOptionsJSON` normalizes it, and `startAuthentication` handles the authenticator prompt.
- Successful responses are forwarded to `/authn/passkey/login/finish` using `postJson` (credentials included) to establish the session and capture the returned `account_thumb_hex` for display.
- `ApiError` instances are rendered via `formatApiError` so toasts show `HTTP <status> — <message>`; authenticator DOM exceptions fall back to `mapDomException` + `normalizeError` for readable text.

# Interactions
- Mounted from `App.tsx`; relies on `apiUrl`/`postJson` to hit the Node server origin and share code with Playwright adapter/post-body tests.
- After success, routes to Dashboard via `location.hash` to expose signing features.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-ERR; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
