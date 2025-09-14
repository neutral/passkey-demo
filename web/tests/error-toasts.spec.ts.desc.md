# Purpose
End-to-end UI checks for error toasts across Register, Login, and Dashboard using Playwright with network stubs.

# Key Logic
- Stubs WebAuthn methods (`navigator.credentials.create/get`) and minimal WebAuthn globals (`window.PublicKeyCredential`, `getClientExtensionResults`) to interoperate with `@simplewebauthn/browser`.
- Mocks backend routes to return specific error statuses and JSON payloads.
- Asserts toast visibility (role="alert") and key texts (HTTP statuses, messages, codes).

# Interactions
- Exercises `ErrorToast`, `parseHttpError`, and page-level wiring in Register, Login, and Dashboard.

# Refs
Refs: requirement R-ERR; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
