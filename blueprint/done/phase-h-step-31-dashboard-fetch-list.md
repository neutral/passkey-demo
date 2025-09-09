### Step 31 — Dashboard: fetch list (Done: 2025-09-09)

Verification notes
- UI: Dashboard fetches `/tx/list` and renders unauthorized/empty/list states; Refresh works.
- Playwright: `tests/dashboard-list.spec.ts` passes (unauthorized prompt when not logged in).
- Manual: After login, Dashboard shows empty state or seeded rows; table updates on Refresh.

31. **Dashboard: fetch list**

    - Scope:

      - Wire the Dashboard page to fetch the authenticated account's transactions from the backend and render them in a simple table.
      - Use absolute API URLs and `credentials: 'include'` so the browser sends the `sid` cookie (set during login). Handle 401 (unauthorized) by prompting the user to Login.
      - Provide a “Refresh” button that re-fetches and updates the list.

    - Source to add/modify:

      - Modify: `web/src/pages/Dashboard.tsx` — On mount and on “Refresh”, `GET /tx/list` with `credentials: 'include'`; render states: loading, unauthorized, empty, list.
      - Optional (minor util): `web/src/lib/http.ts` — `jsonGet(path: string, init?: RequestInit)` thin wrapper using `apiUrl(path)` and defaulting `mode: 'cors'` and `credentials: 'include'`.
      - No backend changes are required; endpoint already exists.

    - Description files (create/update alongside code changes):

      - Update: `web/src/pages/Dashboard.tsx.desc.md` — Document list fetching, states (loading/unauthorized/empty/list), refresh behavior, and cookie usage under CORS. Refs included.
      - Update: `web/web.desc.md` — Note that Dashboard uses authenticated list endpoint with `credentials: 'include'`.

    - Request/response shape:

      - Request: `GET {API_BASE}/tx/list` with cookie `sid` (sent automatically when `credentials: 'include'`).
      - Success (200) JSON:
        - `{ items: Array<{ tx_id_hex: string, nonce: number, message: string, created_at: number }> }`
      - Unauthorized (401): no body required (UI shows a prompt to Login).

    - Algorithm:

      - On Dashboard mount:
        - Set `loading=true`; `error=null`.
        - `fetch(apiUrl('/tx/list'), { method: 'GET', mode: 'cors', credentials: 'include' })`.
        - If `401`: set state `unauthorized=true`, show “Not logged in. Please Login.” with a link/button to go to Login.
        - If `200`: parse JSON; set `items` to `resp.items` (default to empty array if missing); show table with columns: Time, Nonce, Message, Tx ID (hex, truncated UI-only).
        - If other status or parse error: set `error` message and show inline error banner.
        - Set `loading=false` in finally.
      - On “Refresh” button:
        - Re-run the same fetch logic and update the list.
      - Presentation:
        - Empty state: “No transactions yet.”
        - Time: display `new Date(created_at*1000).toLocaleString()` (UI-only; data remains numeric seconds).

    - Database interactions:

      - None on client. Backend resolves the session and returns the account's transactions ordered by `created_at DESC`.

    - Policies & limits:

      - Absolute API URL + `credentials: 'include'` are required for the cookie to be sent under CORS.
      - HttpOnly cookie: not readable in JS; rely on HTTP 200/401 to drive UI state.
      - Keep the table minimal; avoid heavy client-side formatting or pagination in this step (small demo-scale lists).

    - Sequencing:

      - Depends on Step 30 (login flow) to establish a session.
      - Precedes Step 32–33 (build bundle + signing), which will refresh the list on success.

    - Tests (happy path required, plus negatives):

      - Files to add: `web/tests/dashboard-list.spec.ts` (UI test via Playwright).
      - Happy path (unauthorized state):
        - Launch Vite and backend with no prior login; navigate to Dashboard view; assert an unauthorized prompt appears instead of a table.
      - Authorized state (optional, Chromium + Virtual Authenticator):
        - If a prior login exists (e.g., run registration+login manually or seed DB), navigating to Dashboard shows either an empty state or a table with rows; clicking Refresh re-issues the GET and leaves state consistent.
      - Negative cases:
        - Simulate network failure (optional by intercepting request) → error banner visible.
      - Commands:
        - `npm -C web run test:ui -- --project=chromium tests/dashboard-list.spec.ts`

    - Verification:

      - Manual (without login):
        - Start backend and web; open Dashboard directly. Expect “Not logged in. Please Login.”
      - Manual (after login):
        - Register, then Login; navigate to Dashboard; expect either an empty list (“No transactions yet.”) or previous entries; click Refresh and see the list persist/update.

      - User verification commands:

        ```bash
        # 1) Install deps
        npm -C web ci || npm -C web install

        # 2) Run backend and web
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C server run ./cmd/api &
        SERVER_PID=$!
        npm -C web run dev &
        WEB_PID=$!

        # 3) Visit http://localhost:5173 → Login (complete assertion) → go to Dashboard
        #    Verify list renders (empty or with entries); click Refresh and see the list persist/update

        # 4) Cleanup
        kill $WEB_PID || true
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - Dashboard fetches `/tx/list` using absolute URL and `credentials: 'include'`.
      - Unauthorized users see a clear prompt to Login; no crash.
      - Authorized users see a table for non-empty lists and an empty state when there are no transactions.
      - “Refresh” re-fetches and updates the list.

    - Notes:

      - We do not introduce global state or complex routing. Keep logic localized to the Dashboard component.
      - Table rendering remains minimal; any styling refinements are out of scope for this step.

    - Refs:
      - Refs: requirement R-FLOW-SIGN; requirement R-PLAT-1; requirement R-PORTABLE; decision webauthn-corrections-and-standardizations; goal simple-ui-and-storage; spec frontend-api-base-and-cors

