# Purpose
App shell with a minimal, dependency-free router that renders Home, Register, Login, and Dashboard views. Enforces the two-button Home invariant.

# Key Logic
- Maintains `route` state (`'home'|'register'|'login'|'dashboard'`).
- Optionally syncs route with `location.hash` and updates hash on state changes.
- Home shows exactly two primary buttons: Register and Login.

# Interactions
- Renders components from `web/src/pages/`.
- Future network calls will use absolute URLs from `web/src/config.ts` and rely on CORS per Step 26.

# Refs
Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b

