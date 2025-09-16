# Purpose
Dashboard fetches and displays the authenticated account's transactions, and lets the user build a canonical CBOR bundle and sign it via WebAuthn to persist a new transaction.

# Key Logic
- List: `GET /tx/list` (credentials included) hydrates the table, drives unauthorized messaging, and pre-computes the next nonce (`max + 1`) to keep the input read-only.
- Build: `GET /me/account_key` supplies the COSE sender key; the component decodes/normalizes it, encodes canonical CBOR with `bundle.ts`, and previews both base64url + hex for debugging.
- Sign: `postJson` to `/tx/signing/options` posts `{ bundle_cbor_b64 }`, `toRequestOptions` unwraps the Node JSON (nested `options`), `navigator.credentials.get` collects the assertion, and `postJson` sends finish payload from `buildTxFinish`. Success clears bundle state and refreshes the list/nonce.
- Errors: 401 options trigger a probe of `/me/account_key` to differentiate sender-key mismatch from missing session; other `ApiError`s route through `formatApiError` so toasts show `HTTP <status> — <message>`.

# Interactions
- Rendered by `App.tsx` for the `#dashboard` route (post-login). Uses `apiUrl`/`postJson` like registration/login and leans on the same helpers (`bundle`, `cbor`, `webauthn`).
- Exercised by Playwright (`tx-signing*.spec.ts`) which stub Node-style responses to ensure adapters and error handling stay in sync.

# Refs
Refs: requirement R-FLOW-SIGN; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-ERR; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer
