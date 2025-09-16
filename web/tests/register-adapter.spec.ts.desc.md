# Purpose
Unit tests for `toCreationOptionsJSON`: ensure the adapter accepts the Node backend’s SimpleWebAuthn JSON and still enforces policy flags/ES256 expectations.

# Key Logic
- Confirms camelCase payloads (`rp`, `authenticatorSelection`, `pubKeyCredParams`) survive normalization with `attestation: 'none'`, `residentKey: 'required'`, `userVerification: 'required'`, and base64url `challenge`/`user.id`.

# Interactions
- Runs in Playwright by importing the module from Vite dev server; no backend calls.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
