# Purpose
Centralizes the browser-facing API origin so the React app always targets the Node server (`ORIGIN`) without relying on a dev proxy. Keeps CORS credentials aligned with the backend policy.

# Key Logic
- Reads `import.meta.env.VITE_API_BASE` and falls back to the local Node server `http://localhost:8080` (matching `ORIGIN`).
- `apiUrl(path)` promotes relative paths to absolute URLs via `new URL(path, API_BASE)`.

# Interactions
- Used by `lib/api.ts` helpers and every fetch/`postJson` call so cookies are sent to the Node backend and Playwright can reuse the same config in CI.

# Refs
Refs: goal simple-ui-and-storage; goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-OPS-DEV; spec frontend-api-base-and-cors
