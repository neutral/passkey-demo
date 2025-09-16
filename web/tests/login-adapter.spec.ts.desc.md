# Purpose
Unit tests for `toRequestOptionsJSON`: confirms Node camelCase login payloads map cleanly to `PublicKeyCredentialRequestOptionsJSON` and honour policy defaults.

# Key Logic
- Verifies `rpId`, `userVerification: 'required'`, and conditional inclusion of `allowCredentials` when IDs are provided.

# Interactions
- Runs in Playwright importing the module via Vite; no backend calls.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
