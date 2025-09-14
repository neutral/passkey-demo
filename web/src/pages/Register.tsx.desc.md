# Purpose
Register page implements the WebAuthn create() flow end‑to‑end using `@simplewebauthn/browser`: fetch options, convert to `PublicKeyCredentialCreationOptionsJSON` via adapter, call `startRegistration`, and POST the returned JSON to the backend.

# Key Logic
- Fetch `POST /authn/passkey/registration/options`, build `PublicKeyCredentialCreationOptionsJSON` via `toCreationOptionsJSON`, call `startRegistration`, then POST `{ ...attestationJSON, reg_session_id }` to `/authn/passkey/registration/finish`.
- Errors: parses non‑OK responses via `parseHttpError` and normalizes thrown errors via `normalizeError`; renders via `ErrorToast` (dismissible) instead of inline paragraphs.
- Success: displays `account_thumb_hex` (thumb) and offers a “Go to Login” button that navigates via `location.hash`.

# Interactions
- Called by `App.tsx` when user selects Register from Home. Uses absolute URLs via `web/src/config.ts` and relies on CORS (Step 26).
- Uses `web/src/lib/http.ts` for error shaping and `web/src/components/ErrorToast.tsx` for display.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-ERR; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
