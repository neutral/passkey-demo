### Step 27 — Basic pages (Done: 2025-09-05)

Verification notes
- Build: `npm -C web run build` → OK
- Lint: `npm -C web run lint` → OK
- Playwright UI: `npm -C web run test:ui` → 1 passed (Home two buttons; navigation works)
- Static checks: no Vite proxy configured; no relative API fetch paths found
- Backend `/health`: OK (server started with RP_ID=localhost, ORIGIN=http://localhost:5173)
- Web dev server: OK on port 5173

27. **Basic pages**

    - Scope:

      - Scaffold a minimal SPA UI with three pages: `Register`, `Login`, `Dashboard`.
      - Replace Vite template content in `App.tsx` with a simple, dependency-free router (local state or hash-based) and a home view that shows exactly two primary actions: Register and Login. Do not implement WebAuthn or API calls yet.
      - Enforce absolute API usage by preparing a central `API_BASE` config and removing any Vite dev proxy. Calls will be wired in later steps.

    - Source to add/modify:

      - Add: `web/src/pages/Register.tsx` — Placeholder component with heading and a primary action button to start registration (no logic yet).
      - Add: `web/src/pages/Login.tsx` — Placeholder component with heading and a primary action button to start login (no logic yet).
      - Add: `web/src/pages/Dashboard.tsx` — Placeholder component with heading and placeholders for list and sign form (no data yet).
      - Add: `web/src/config.ts` — Central config exporting `API_BASE` (from `import.meta.env.VITE_API_BASE || 'http://localhost:8080'`) and a helper `apiUrl(path: string)` that always returns absolute URLs. This file will be used from Step 29 onward.
      - Modify: `web/src/App.tsx` — Replace template with a minimal SPA shell and route state: `'home' | 'register' | 'login' | 'dashboard'`. Home renders exactly two prominent buttons to navigate to Register and Login; add simple navigation back to Home from sub-pages. Do not add external routing libraries.
      - Modify: `web/vite.config.ts` — Remove the `server.proxy` block to ensure no dev proxy is used. Keep port at 5173; optionally set `strictPort: true` to avoid port drift.
      - Add (tests): `web/playwright.config.ts`, `web/tests/ui.spec.ts` — Minimal Playwright setup to verify two-button Home and basic navigation; add `test:ui` script and `@playwright/test` devDependency in `web/package.json`.

    - Description files (create/update alongside code changes):

      - Create: `web/src/pages/pages.desc.md` — Overview of pages folder responsibilities and relationships. Refs included.
      - Create: `web/src/pages/Register.tsx.desc.md` — High-level purpose and interactions (to be wired later). Refs included.
      - Create: `web/src/pages/Login.tsx.desc.md` — High-level purpose and interactions (to be wired later). Refs included.
      - Create: `web/src/pages/Dashboard.tsx.desc.md` — Purpose, placeholders for list/sign form, future interactions. Refs included.
      - Create: `web/src/config.ts.desc.md` — Purpose of absolute API config and no-proxy policy. Refs included.
      - Create/Update: `web/src/App.tsx.desc.md` — App shell, minimal router, home two-button invariant, and navigation behavior. Refs included.
      - Create/Update: `web/vite.config.ts.desc.md` — Dev server config with explicit statement: no proxy; CORS relied upon per Step 26. Refs included.
      - Update: `web/web.desc.md` — Note absolute API usage and that no Vite proxy is used, relying on CORS from Step 26.

    - Request/response shape:

      - N/A for this step (no API calls yet). Future steps will define WebAuthn and signing payload shapes.

    - Algorithm (UI behavior and invariants):

      - App boot:
        - Initialize a route state: default `'home'`.
        - Optionally read initial route from `location.hash` in dev; keep logic minimal and resilient (unknown hash → `'home'`).
      - Home view:
        - Render exactly two primary buttons: Register, Login.
        - Clicking Register → set route `'register'`; clicking Login → set route `'login'`.
      - Sub-pages:
        - `Register` and `Login` pages: render headings and a primary action button each (no handlers yet), with a “Back” link to home. Do not perform any fetch or WebAuthn call in this step.
        - `Dashboard` page: render a heading and placeholders for list and sign form; include “Back” link to home (dashboard will be shown post-login in Step 30).
      - Absolute API configuration:
        - `web/src/config.ts` defines `API_BASE` and `apiUrl()` returning `new URL(path, API_BASE).toString()`; do not export or use any relative `/api` paths.
      - Dev server policy:
        - Ensure `vite.config.ts` contains no `server.proxy` config; all API calls in later steps must use absolute URLs and `credentials: 'include'` (per Step 26 CORS & cookies).

    - Database interactions:

      - None in this step.

    - Policies & limits:

      - Two-button invariant: Home shows only two primary actions (Register, Login) and no other primary actions.
      - No Vite proxy: `vite.config.ts` must not define `server.proxy`.
      - Absolute API URLs only: Centralize base in `web/src/config.ts` and forbid relative `/api` paths.
      - No auth state in localStorage/sessionStorage; rely on server session cookie (to be exercised later).

    - Sequencing:

      - Depends on Step 26 (CORS & cookies) being in place to allow cross-origin requests in later steps.
      - Prepares scaffolding for Steps 28–34 (helpers, WebAuthn flows, dashboard list and signing).
      - Do not wire any network calls yet; keep the UI purely presentational.

    - Tests (plan):

      - Happy path (manual):
        - App loads without console errors; Home shows exactly two buttons: “Register” and “Login”.
        - Clicking each navigates to its page showing the correct heading and a way to return to Home.
      - Automated UI (Playwright):
        - Files: `web/playwright.config.ts`, `web/tests/ui.spec.ts`.
        - Script: `npm -C web run test:ui`.
        - Asserts: Home header visible; exactly two buttons (Register, Login); navigate to Register, Back; navigate to Login, Back.
      - Negative/guard checks (manual + static):
        - Verify `vite.config.ts` contains no `proxy` config.
        - Static search shows no usage of relative API paths (e.g., `fetch('/authn` or `fetch('/tx` or `'/api'`).
      - Commands:
        - Type check and build: `npm -C web run build`.
        - Lint: `npm -C web run lint`.
        - Static checks:
          - `rg -n "proxy\s*:" web/vite.config.ts || echo "No proxy config present"`
          - `rg -n "fetch\(\s*['\"]/(authn|tx|api)" web/src || echo "No relative API paths found"`

    - Verification:

      - Build passes; dev server runs on port 5173; the SPA renders Home with exactly two buttons and navigates to placeholder pages.
      - No proxy in dev config; config file for absolute API exists.
      - Automated UI: Playwright test passes and is kept for future steps (`npm -C web run test:ui`).
      - Fix-forward loop: If any check fails (extra buttons, missing pages, lingering proxy, or relative API paths), adjust files and re-run the checks until satisfied.

      - User verification commands:

        ```bash
        # 1) Install deps
        npm -C web ci || npm -C web install

        # 2) Type-check and build
        npm -C web run build

        # 3) Lint
        npm -C web run lint || true

        # 4) Run Playwright UI test
        npm -C web run test:ui --silent

        # 5) Confirm no Vite proxy configured
        rg -n "proxy\s*:" web/vite.config.ts || echo "OK: no proxy configured"

        # 6) Confirm no relative API paths in sources
        rg -n "fetch\(\s*['\"]/(authn|tx|api)" web/src || echo "OK: no relative API fetch paths found"

        # 7) Run backend (separate shell)
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./server/cmd/api &
        SERVER_PID=$!

        # 8) Run web dev server (observe output for port 5173)
        npm -C web run dev &
        WEB_PID=$!

        # 9) Manually open http://localhost:5173 in a browser
        #    Verify Home shows exactly two buttons: Register, Login.
        #    Click both and verify navigation to the respective placeholder pages.

        # 10) Cleanup
        kill $WEB_PID || true
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - `web/src/pages/{Register,Login,Dashboard}.tsx` exist and render headings.
      - `web/src/App.tsx` renders a Home view with exactly two primary buttons (Register, Login) and minimal navigation to each page.
      - `web/vite.config.ts` contains no `server.proxy` configuration.
      - `web/src/config.ts` exists and exposes `API_BASE` and `apiUrl()` for absolute URL usage.
      - Build and dev server run without errors; static checks show no relative API paths.

    - Notes:

      - Keep UI minimal and dependency-free (no router libraries or global stores). Hash-based routing is optional; route state is sufficient for the demo.
      - Absolute API URLs are required for all future network calls; do not reintroduce a dev proxy.

    - Refs:
      - Refs: goal ui-simplicity-two-buttons; goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-UI-2BTN; spec spec-a; spec spec-b

