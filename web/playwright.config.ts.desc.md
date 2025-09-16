# Overview
Defines Playwright test defaults and shared webServer processes. Launches Vite dev server and the Node backend (`node-server/src/server.js`) so UI tests exercise the new API shapes end-to-end.

# Relations
Used by `npm run test:ui` / CI to spin up Vite + Node (`RP_ID=localhost`, `ORIGIN=http://localhost:5173`, `DB_PATH=../node-server/playwright.db`) before executing specs in `web/tests/`. Ensures tests hit the same origin and credentials settings configured in `web/src/config.ts`.

# Interfaces & Models
Configures Chromium project, shared baseURL `http://localhost:5173`, trace-on-retry, and two `webServer` entries (Vite + Node).

# Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors
