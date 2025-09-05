# Overview
Container for SPA page components: Register, Login, and Dashboard. These components render the primary UI states and are orchestrated by `App.tsx` via a minimal in-app router.

# Relations
- Used by `web/src/App.tsx` to render views based on local route state or hash fragment.
- Will call backend endpoints (registration/login/signing) in later steps using absolute URLs from `web/src/config.ts`.

# Interfaces & Models
- `Register`, `Login`, `Dashboard` React components. Presentational in this step; no props besides `onBack` navigation.

# Refs
Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b

