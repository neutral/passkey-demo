# Purpose
Minimal Login page rendering a heading and a primary action button. Actual WebAuthn assertion flow is added in Step 30.

# Key Logic
- Presentational only in this step; provides an `onBack` handler for navigation.

# Interactions
- Rendered by `App.tsx` when user selects Login from Home.
- No fetch or WebAuthn yet; later will use absolute API URLs and send credentials under CORS.

# Refs
Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b

