# Purpose
Validates Dashboard signing error handling when the backend returns failures. Ensures user-facing messages render without crashes.

- Stubs `navigator.credentials.get` and seeds Node-style options payloads.
- Mocks `GET /tx/list` (authorized) and conditionally overrides `/me/account_key` to trigger the unauthorized branch.
- Exercises error cases:
  - Options 401 → unauthorized prompt (account-key probe also 401).
  - Options 409/400 → toast shows `HTTP <status> — <message>` from `formatApiError`.
  - Finish 409/400 → finish errors surface the formatted message.

# Interactions
- Drives `web/src/pages/Dashboard.tsx` through the UI using Playwright. Uses absolute API URLs and CORS consistent with app config.

# Refs
Refs: requirement R-FLOW-SIGN; requirement R-ERR; decision webauthn-corrections-and-standardizations
