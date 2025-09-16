# Purpose
End-to-end UI checks for error toasts across Register, Login, and Dashboard using Playwright with network stubs.

# Key Logic
- Stubs WebAuthn globals so `@simplewebauthn/browser` works in the headless browser.
- Serves Node-style options payloads and targeted error envelopes (`message`/`code`) to drive `formatApiError` output.
- Validates toast rendering (`role="alert"`) and visible strings (status, message, codes) for register/login/signing flows.

# Interactions
- Exercises `ErrorToast`, `parseHttpError`, and page-level wiring in Register, Login, and Dashboard.

# Refs
Refs: requirement R-ERR; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
