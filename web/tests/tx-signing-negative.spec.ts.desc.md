# Purpose
Validates Dashboard signing error handling when the backend returns failures. Ensures user-facing messages render without crashes.

# Key Logic
- Stubs `navigator.credentials.get` to avoid platform prompts.
- Mocks `GET /me/account_key` and `GET /tx/list` to keep the UI reachable.
- Cases covered:
  - Options 401 → unauthorized prompt shown.
  - Options 409 → conflict error banner.
  - Options 400 → invalid bundle error.
  - Finish 409/400 → finish HTTP error surfaced.

# Interactions
- Drives `web/src/pages/Dashboard.tsx` through the UI using Playwright. Uses absolute API URLs and CORS consistent with app config.

# Refs
Refs: requirement R-FLOW-SIGN; global/error-responses-and-limits; decision webauthn-corrections-and-standardizations

