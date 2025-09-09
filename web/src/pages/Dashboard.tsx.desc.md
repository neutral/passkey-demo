# Purpose
Dashboard fetches and displays the authenticated account's transactions and provides a simple Refresh control. Signing UI remains disabled until later steps.

# Key Logic
- On mount and on Refresh, `GET /tx/list` using absolute API URL via `apiUrl()` with `mode: 'cors'` and `credentials: 'include'` so the `sid` cookie is sent.
- Renders states: loading, unauthorized (prompt to Login), empty list, or a table of `{ time, nonce, message, tx_id_hex }`.
- Does not read cookies (HttpOnly); relies on HTTP 200/401 to drive UI state.

# Interactions
- Rendered by `App.tsx` when route is `dashboard` (typically after login). Provides an `onBack` handler. Uses backend CORS from Step 26.

# Refs
Refs: requirement R-FLOW-SIGN; requirement R-PLAT-1; requirement R-UI-2BTN; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
