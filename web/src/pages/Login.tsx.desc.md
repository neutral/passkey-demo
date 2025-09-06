# Purpose
Login page implements the WebAuthn get() flow: fetch options, call `navigator.credentials.get`, and POST finish with `credentials: 'include'` to establish a session (cookie).

# Key Logic
- Fetch `POST /authn/passkey/login/options`, build `PublicKeyCredentialRequestOptions` using `toRequestOptions`, call `navigator.credentials.get`, then `buildLoginFinish` and `POST /authn/passkey/login/finish` with `credentials: 'include'`.
- Shows inline errors; on success displays `account_thumb_hex` and routes to Dashboard.

# Interactions
- Rendered by `App.tsx` when user selects Login from Home. Uses absolute URLs via `web/src/config.ts` and relies on CORS (Step 26). Receives HttpOnly cookie (not readable by JS).

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; requirement R-UI-2BTN; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b
