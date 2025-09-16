# Purpose
Page test validating that the registration UI posts `{ reg_session_id, ...RegistrationResponseJSON }` while consuming the Node backend’s camelCase options JSON.

# Key Logic
- Stubs `navigator.credentials.create`, serves Node-style options (`rp`, `authenticatorSelection`, etc.), and intercepts both options/finish endpoints.
- Asserts that the posted JSON contains expected fields and string encodings.

# Interactions
- Exercises `web/src/pages/Register.tsx` through the UI. No real backend is called; responses are routed via Playwright.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
