# Purpose
Dev server configuration for the SPA. Explicitly avoids using a proxy so the frontend always calls the backend via absolute URLs and relies on CORS (Step 26).

# Key Logic
- Port `5173` with `strictPort: true` to avoid port drift.
- No `server.proxy` block is defined.

# Interactions
- Frontend calls backend at `API_BASE` from `web/src/config.ts`.

# Refs
Refs: goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-OPS-DEV; requirement R-UI-2BTN; spec spec-a; spec spec-b

