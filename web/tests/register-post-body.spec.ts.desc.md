# Purpose
Page test to validate the registration finish POST body when using `@simplewebauthn/browser`. Ensures the payload includes `reg_session_id` and base64url string fields per the server contract.

# Key Logic
- Stubs `navigator.credentials.create` and intercepts both options and finish endpoints.
- Asserts that the posted JSON contains expected fields and string encodings.

# Interactions
- Exercises `web/src/pages/Register.tsx` through the UI. No real backend is called; responses are routed via Playwright.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
