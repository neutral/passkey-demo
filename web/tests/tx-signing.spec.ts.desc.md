# Purpose
Covers the happy-path UI signing flow end-to-end using Playwright with network stubs. Validates that Dashboard posts options with `bundle_cbor_b64`, transforms to WebAuthn request options, performs a stubbed `navigator.credentials.get`, builds a finish payload via `buildTxFinish`, posts `/tx/signing/finish`, and clears the preview while refreshing the list.

# Key Logic
- Stubs `navigator.credentials.get` in `addInitScript` to avoid real prompts. Uses the dashboard's default next-nonce value (input is read-only).
- Mocks `GET /me/account_key`, `GET /tx/list`, `POST /tx/signing/options`, and `POST /tx/signing/finish`.
- Asserts request body includes `bundle_cbor_b64`; asserts finish payload includes `tx_session_id` and base64url fields; verifies preview is cleared after success.

# Interactions
- Exercises `web/src/pages/Dashboard.tsx` and `web/src/lib/webauthn.ts` through the browser. Uses absolute API URLs and CORS per app config.

# Refs
Refs: requirement R-FLOW-SIGN; spec account-binding-explainer; spec bundle-shape-and-client-production-explainer; decision webauthn-corrections-and-standardizations
