# Purpose
Minimal Register page rendering a heading and a primary action button. Actual WebAuthn registration logic is added in Step 29.

# Key Logic
- Purely presentational in this step. Exposes an `onBack` handler for navigation back to Home.

# Interactions
- Called by `App.tsx` when user selects Register from Home.
- No network or WebAuthn calls in this step; to be added later with absolute API URLs and CORS support.

# Refs
Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b
# Purpose
Register page implements the WebAuthn create() flow end-to-end: fetch options, call `navigator.credentials.create`, and POST finish to the backend.

# Key Logic
- Fetch `POST /authn/passkey/registration/options`, build `PublicKeyCredentialCreationOptions` using `toCreationOptions`, call `navigator.credentials.create`, then `buildRegFinish` and `POST /authn/passkey/registration/finish`.
- Shows inline errors; on success displays `account_thumb_hex` and provides a “Go to Login” button (routes via hash).

# Interactions
- Called by `App.tsx` when user selects Register from Home. Uses absolute URLs via `web/src/config.ts` and relies on CORS (Step 26).

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; requirement R-UI-2BTN; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b
