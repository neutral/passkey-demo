# Purpose
Dashboard fetches and displays the authenticated account's transactions, and lets the user build a canonical CBOR bundle and sign it via WebAuthn to persist a new transaction.

# Key Logic
- List: On mount and on Refresh, `GET /tx/list` using absolute API URL via `apiUrl()` with `mode: 'cors'` and `credentials: 'include'` so the `sid` cookie is sent. Renders loading/unauthorized/empty/list states. Does not read cookies (HttpOnly); relies on HTTP 200/401.
- Build: Fetch `sender_key` via `GET /me/account_key` (authenticated), collect `message` and `nonce`, construct integer-key bundle `{0: sender_key, 1: nonce, 2: message, 3?: valid_until}`, and encode canonical CBOR using `cbor-x`. Preview `bundle_cbor_b64` and hex(B).
- Sign: `POST /tx/signing/options` with `{ bundle_cbor_b64 }`, transform to `PublicKeyCredentialRequestOptions`, call `navigator.credentials.get`, then `POST /tx/signing/finish`; on success, clear inputs and refresh the list.

# Interactions
- Rendered by `App.tsx` when route is `dashboard` (typically after login). Provides an `onBack` handler. Uses backend CORS from Step 26.
- Relies on `web/src/lib/bundle.ts` and `web/src/lib/cbor.ts` for encoding; `web/src/lib/webauthn.ts` for request option transform and finish payload builders.

# Refs
Refs: requirement R-FLOW-SIGN; requirement R-PLAT-1; requirement R-UI-2BTN; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer
