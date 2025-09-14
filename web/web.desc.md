# Overview
Frontend single-page application (React + Vite) providing a minimal UI to Register and Login with a passkey, and a post-login Dashboard to sign messages and view transactions. It calls backend endpoints and handles WebAuthn ceremonies via `@simplewebauthn/browser`, with lightweight adapters that map current Go server option shapes to the library’s JSON shapes. All API calls use absolute URLs from `web/src/config.ts` (no Vite proxy) and rely on CORS.

# Relations
- Uses backend REST endpoints for registration, login, and signing; depends on server-supplied options and cookies.
- Performs base64url ⇄ ArrayBuffer conversions and builds canonical CBOR bundles for transaction signing using a small CBOR library.

# Interfaces & Models
- Views: Register, Login, Dashboard (two primary actions on home; dashboard shows list and sign form).
- API calls: `POST /authn/passkey/registration/options|finish`, `POST /authn/passkey/login/options|finish`, `POST /tx/signing/options|finish`, `GET /tx/list`.
- Absolute URL policy: calls constructed via `apiUrl()` against `API_BASE`.
- Encoding utilities: base64url and UTF‑8 helpers live in `web/src/lib/encoding.ts` and are used across flows.
- WebAuthn helpers: `web/src/lib/webauthn.ts` provides adapters to `@simplewebauthn/browser` JSON shapes for registration/login and keeps a `buildTxFinish` helper for Dashboard signing until server parity is reached.
- WebAuthn: `startRegistration`/`startAuthentication` with `userVerification: "required"` and discoverable credentials (residentKey "required").

# Refs
Refs: goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
