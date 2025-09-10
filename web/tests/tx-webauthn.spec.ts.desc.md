# Purpose
Unit-style test for `buildTxFinish` to ensure assertion fields are base64url-encoded and `tx_session_id` is included. Runs in the browser context served by Vite.

# Key Logic
- Constructs a fake `PublicKeyCredential` object shape with ArrayBuffer fields.
- Calls `buildTxFinish(cred, 'txsess-1')`; asserts presence/types of encoded fields and session id.

# Interactions
- Imports `web/src/lib/webauthn.ts` via Vite dev server path; no network calls.

# Refs
Refs: requirement R-FLOW-SIGN; spec account-binding-explainer; decision webauthn-corrections-and-standardizations

