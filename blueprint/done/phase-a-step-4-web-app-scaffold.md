### Step 4 — Web app scaffold (Done: 2025-08-31)

- Context

  - Initialize a minimal React + Vite SPA to support Register/Login/Dashboard in later steps.

- Structure

  - Ensure `web/` exists. If not present, scaffold from repo root.
    - New project: `npm create vite@latest web -- --template react`.
    - Existing folder: `cd web && npm create vite@latest . -- --template react`.

- Source to add (instructions only)

  - Ensure `package.json` contains scripts: `dev`, `build`, `preview` (Vite defaults).
  - Confirm `vite.config.ts` exists; keep defaults (proxy optional; CORS handled in Step 26).
  - Keep `src/App.tsx` as minimal root component; no additional pages yet (arrive in Step 27+).
  - Optional: add `.env.development` with `VITE_API_BASE=http://localhost:8080` for later use (not required yet).

- Description files to add (instructions only)

  - None; high-level `web/web.desc.md` already exists from Step 1. Add per-file descriptions when pages/components are implemented (Step 27+).

- Blueprint updates

  - Refs to include upon implementation: goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-UI-2BTN; requirement R-OPS-DEV.

- Verification (to run after implementation)

  - Install deps: `cd web && npm ci`.
  - Dev server: `cd web && npm run dev` (expect server on http://localhost:5173).
  - Build: `cd web && npm run build` (expect dist/ output with no errors).
  - Optional: `curl -I http://localhost:5173` after dev server starts (expect 200 OK).

- User verification commands (copy/paste)

  ```bash
  # Install deps (if node_modules absent)
  test -d web/node_modules || (cd web && npm ci)

  # Start dev server in background and verify
  cd web
  npm run dev > /tmp/vite.log 2>&1 & echo $! > /tmp/vite.pid
  sleep 1
  curl -I http://localhost:5173
  kill $(cat /tmp/vite.pid) && rm -f /tmp/vite.pid

  # Production build
  npm run build
  cd -
  ```

- Notes
  - Do not wire API calls or pages yet; keep the scaffold minimal and compilable.
  - CORS is configured in Step 26. A Vite proxy is optional for local convenience and can be added later.

