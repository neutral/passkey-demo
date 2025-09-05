# Purpose
Minimal Register page rendering a heading and a primary action button. Actual WebAuthn registration logic is added in Step 29.

# Key Logic
- Purely presentational in this step. Exposes an `onBack` handler for navigation back to Home.

# Interactions
- Called by `App.tsx` when user selects Register from Home.
- No network or WebAuthn calls in this step; to be added later with absolute API URLs and CORS support.

# Refs
Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b

