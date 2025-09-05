# Overview
Frontend single-page application (React + Vite) providing a minimal UI to Register and Login with a passkey, and a post-login Dashboard to sign messages and view transactions. It calls backend endpoints and handles WebAuthn ceremonies and binary conversions. All API calls use absolute URLs from `web/src/config.ts` (no Vite proxy) and rely on CORS as configured in Step 26.

# Relations
- Uses backend REST endpoints for registration, login, and signing; depends on server-supplied options and cookies.
- Performs base64url ⇄ ArrayBuffer conversions and builds canonical CBOR bundles for transaction signing using a small CBOR library.

# Interfaces & Models
- Views: Register, Login, Dashboard (two primary actions on home; dashboard shows list and sign form).
- API calls: `POST /authn/passkey/registration/options|finish`, `POST /authn/passkey/login/options|finish`, `POST /tx/signing/options|finish`, `GET /tx/list`.
- Absolute URL policy: calls constructed via `apiUrl()` against `API_BASE`.
- WebAuthn: `navigator.credentials.create` and `navigator.credentials.get` with `userVerification: "required"`.

# Refs
Refs: goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN
