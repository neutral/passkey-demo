# Purpose
Login page implements the WebAuthn get() flow: fetch options, call `navigator.credentials.get`, and POST finish with `credentials: 'include'` to establish a session (cookie).

# Key Logic
- Fetch `POST /authn/passkey/login/options`, build `PublicKeyCredentialRequestOptions` using `toRequestOptions`, call `navigator.credentials.get`, then `buildLoginFinish` and `POST /authn/passkey/login/finish` with `credentials: 'include'`.
- Errors: parses non‑OK responses via `parseHttpError` (shows status like `HTTP 401`) and normalizes thrown errors via `normalizeError`; renders via `ErrorToast` (dismissible) instead of inline paragraphs.
- Success: displays `account_thumb_hex` and routes to Dashboard.

# Interactions
- Rendered by `App.tsx` when user selects Login from Home. Uses absolute URLs via `web/src/config.ts` and relies on CORS (Step 26). Receives HttpOnly cookie (not readable by JS).
- Uses `web/src/lib/http.ts` for error shaping and `web/src/components/ErrorToast.tsx` for display; keeps error strings compatible with UI tests (e.g., includes `HTTP 401`).

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-ERR; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
