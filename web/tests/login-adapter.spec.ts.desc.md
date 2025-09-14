# Purpose
Unit tests for `toRequestOptionsJSON` adapter: verifies mapping to `PublicKeyCredentialRequestOptionsJSON`, enforcing `userVerification: 'required'`, setting `rpId`, and omitting `allowCredentials` when empty.

# Key Logic
- Confirms `rpId` and `userVerification` values and the presence/absence of `allowCredentials` depending on input.

# Interactions
- Runs in Playwright importing the module via Vite; no backend calls.

# Refs
Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
