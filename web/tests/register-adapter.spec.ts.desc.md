# Purpose
Unit tests for `toCreationOptionsJSON` adapter: verifies correct mapping from Go server option shape to `PublicKeyCredentialCreationOptionsJSON` for `@simplewebauthn/browser`.

# Key Logic
- Asserts `rp.id`, `attestation: 'none'`, `residentKey: 'required'`, `userVerification: 'required'`, presence of ES256 in `pubKeyCredParams`, and that `challenge` remains a base64url string and `user.id` is a string.

# Interactions
- Runs in Playwright by importing the module from Vite dev server; no backend calls.

# Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
