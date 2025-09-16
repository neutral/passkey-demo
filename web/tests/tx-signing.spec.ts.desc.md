# Purpose
Happy-path dashboard signing flow under Playwright. Ensures Node-style `/tx/signing/options` payloads (with nested `options`) are handled correctly and that finish requests carry the expected base64url fields.

# Key Logic
- Stubs `navigator.credentials.get` in `addInitScript` to avoid real prompts. Uses the dashboard's default next-nonce value (input is read-only).
- Mocks `GET /me/account_key`, `GET /tx/list`, `POST /tx/signing/options` (Node JSON with nested `options`), and `POST /tx/signing/finish`.
- Asserts request body includes `bundle_cbor_b64`; asserts finish payload includes `tx_session_id` and base64url fields; verifies preview is cleared after success.

# Interactions
- Exercises `web/src/pages/Dashboard.tsx` and `web/src/lib/webauthn.ts` through the browser. Uses absolute API URLs and CORS per app config.

# Refs
Refs: requirement R-FLOW-SIGN; spec account-binding-explainer; spec bundle-shape-and-client-production-explainer; decision webauthn-corrections-and-standardizations
