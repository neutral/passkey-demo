# Purpose
Validates Dashboard signing error handling when the backend returns failures. Ensures user-facing messages render without crashes.

# Key Logic
- Stubs `navigator.credentials.get` to avoid platform prompts.
- Mocks `GET /tx/list` to keep the UI reachable. Uses dashboard default next-nonce (input is read-only).
- For the 401 case, overrides `GET /me/account_key` to return 401 so the UI shows the unauthorized prompt; other cases keep a 200 response.
- Cases covered:
  - Options 401 → unauthorized prompt shown (account key probe returns 401).
  - Options 409 → conflict error banner.
  - Options 400 → invalid bundle error.
  - Finish 409/400 → finish HTTP error surfaced.

# Interactions
- Drives `web/src/pages/Dashboard.tsx` through the UI using Playwright. Uses absolute API URLs and CORS consistent with app config.

# Refs
Refs: requirement R-FLOW-SIGN; global/error-responses-and-limits; decision webauthn-corrections-and-standardizations
