# Purpose
Page test that the login UI posts `{ login_session_id, ...AuthenticationResponseJSON }` while consuming the Node backend’s camelCase options JSON.

# Key Logic
- Stubs `navigator.credentials.get`, serves Node-style options (top-level `rpId`, `userVerification`, `allowCredentials`), and intercepts both options/finish endpoints.
- Asserts that the posted JSON contains expected fields and string encodings.

# Interactions
- Exercises `web/src/pages/Login.tsx` through the UI. No real backend is called; responses are routed via Playwright.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
