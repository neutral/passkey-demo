# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Done (One liner)

Completed steps are stored as individual files under `blueprint/done/`, with phase information prefixed in each filename (e.g., `phase-a-step-1-...md`). Simply add a single line summary here for all done steps with their number and link to the done file.

- [Step 1 — Initialize repo](blueprint/done/phase-a-step-1-initialize-repo.md)
- [Step 2 — Server Go module](blueprint/done/phase-a-step-2-server-go-module.md)
- [Step 3 — Add deps](blueprint/done/phase-a-step-3-add-deps.md)
- [Step 4 — Web app scaffold](blueprint/done/phase-a-step-4-web-app-scaffold.md)
- [Step 5 — Server config struct](blueprint/done/phase-b-step-5-server-config-struct.md)
- [Step 6 — DB init & migrations](blueprint/done/phase-b-step-6-db-init-and-migrations.md)
- [Step 7 — Types: COSE, WebAuthn, Bundle](blueprint/done/phase-b-step-7-types-cose-webauthn-bundle.md)
- [Step 8 — COSE → ECDSA helper](blueprint/done/phase-b-step-8-cose-to-ecdsa-helper.md)
- [Step 9 — Base64url utilities](blueprint/done/phase-b-step-9-base64url-utilities.md)
- [Step 10 — CBOR canonical codec](blueprint/done/phase-b-step-10-cbor-canonical-codec.md)
- [Step 11 — Parse authenticatorData](blueprint/done/phase-c-step-11-parse-authenticatordata.md)
- [Step 12 — ClientDataJSON validation](blueprint/done/phase-c-step-12-clientdatajson-validation.md)
- [Step 13 — Signature verify utility](blueprint/done/phase-c-step-13-signature-verify-utility.md)
- [Step 14 — RP ID & Origin checks](blueprint/done/phase-c-step-14-rp-origin-checks.md)
- [Step 15 — /authn/passkey/registration/options handler](blueprint/done/phase-d-step-15-registration-options-handler.md)
- [Step 16 — Attestation parsing (minimal)](blueprint/done/phase-d-step-16-attestation-parsing.md)
- [Step 17 — /authn/passkey/registration/finish handler](blueprint/done/phase-d-step-17-registration-finish-handler.md)
- [Step 18 — /authn/passkey/login/options handler](blueprint/done/phase-e-step-18-login-options-handler.md)
- [Step 19 — /authn/passkey/login/finish handler](blueprint/done/phase-e-step-19-login-finish-handler.md)
- [Step 20 — Bundle validation helper](blueprint/done/phase-f-step-20-bundle-validation-helper.md)
- [Step 21 — /tx/signing/options handler](blueprint/done/phase-f-step-21-signing-options-handler.md)
- [Step 22 — /tx/signing/finish handler](blueprint/done/phase-f-step-22-signing-finish-handler.md)
- [Step 23 — /tx/list handler](blueprint/done/phase-f-step-23-tx-list-handler.md)
- [Step 24 — Session middleware](blueprint/done/phase-g-step-24-session-middleware.md)
- [Step 25 — Rate limiting & limits](blueprint/done/phase-g-step-25-rate-limiting-limits.md)
- [Step 26 — CORS & cookies](blueprint/done/phase-g-step-26-cors-and-cookies.md)
- [Step 27 — Basic pages](blueprint/done/phase-h-step-27-basic-pages.md)
- [Step 28 — Base64url helpers (web)](blueprint/done/phase-h-step-28-base64url-helpers-web.md)
- [Step 29 — WebAuthn create() flow (Register)](blueprint/done/phase-h-step-29-webauthn-register.md)
- [Step 30 — WebAuthn get() flow (Login)](blueprint/done/phase-h-step-30-webauthn-login.md)
- [Step 31 — Dashboard: fetch list](blueprint/done/phase-h-step-31-dashboard-fetch-list.md)
- [Step 30 — WebAuthn get() flow (Login)](blueprint/done/phase-h-step-30-webauthn-login.md)

## Next

### Phase H — Frontend (React) UI

---

31. **Dashboard: fetch list** — Moved to Done

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
        #    Verify list renders (empty or with entries); click Refresh and see state persist/update

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

---

32. **Dashboard: build bundle**

    - UI collects `message` (string) and a user-entered `nonce` (demo-only; no helper endpoint).
    - Build `Bundle` object in JS compatible with the CDDL.
    - Install a small CBOR lib (e.g., `cbor-x`) and encode canonical CBOR in the browser; ensure deterministic/canonical mode is enabled to match server hashing.
    - _Verify_: preview CBOR (hex) in console; repeated encodes of the same input produce the same bytes.

---

33. **Transaction signing options**

    - `POST /tx/signing/options` with `bundle_cbor_b64` using absolute API URL and `credentials: 'include'`.
    - Receive `publicKey` options; call `navigator.credentials.get(...)`.
    - Send `/tx/signing/finish`.
    - On success: refresh list.
    - _Verify_: message appears in table with nonce and timestamp.

---

34. **Error toasts**

    - Best-effort now: render HTTP status and any JSON `{ error, code? }` if present; fall back to status text. Upgrade to standardized envelopes after Step 38.
    - _Verify_: cause a failure (e.g., bad nonce or origin) and see a visible inline alert with informative text.

### Phase I — Testing & Fixtures

---

35. **Unit tests (Go)**

    - COSE→ECDSA, CDJ parse, AD parse, low‑S check, canonical encoding, hash anchors.
    - _Verify_: `go test ./...` passes.

---

36. **Golden vectors**

    - Create a tiny CLI in `server/cmd/vectors` that:

      - Loads a known COSE key, builds a sample bundle, prints hex(B), challenge, tx_id.

    - _Verify_: consistent outputs between runs.

---

37. **Manual E2E**

    - Run server + web; perform registration, login, and sign two messages with nonces 1, 2 in Safari and Chrome on macOS.
    - _Verify_: `/tx/list` shows two entries; DB reflects persisted rows; flows succeed in both browsers with platform authenticators.

### Phase J — Hardening (Demo-grade)

---

38. **Input validation & errors**

    - Enforce message length ≤ 1024; nonce ≤ `2^53 - 1`; origin/rpId in allowlist; body size ≤ 64 KB.
    - Implement standardized error envelope `{ code, error, correlation_id? }` and map to HTTP 400/401/403/409/413/429/5xx per R-ERR.
    - _Verify_: oversize blocked with 400/413; invalid/replay/UV/origin issues map to correct codes; frontend displays `code`.

---

39. **Session security**

    - Random session IDs (≥128 bits), expiry 1h, renewal on activity.
    - _Verify_: expired session returns 401; re-login works.

---

40. **Logging**

    - Structured logs with event names: `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`.
    - _Verify_: logs show account thumb and tx_id (hex).

---

41. **Build scripts**

    - Root Makefile: `make server`, `make web`, `make run`, `make clean`.
    - _Verify_: one command runs both.

### Phase K — Developer Experience

---

42. **API examples**

    - Add `docs/` with curl examples for each endpoint (sans WebAuthn ceremony).
    - _Verify_: docs render in repo.

---

43. **Postman / REST Client file**

    - Provide a collection with placeholders; helpful for observing JSON shapes.
    - _Verify_: collection can be imported.

---

44. **Env sample**

    - `.env.example` with `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`.
    - _Verify_: loads correctly.

---

45. **README**

    - Quickstart, limitations (no attestation trust), and demo notes (Touch ID prompts).
    - _Verify_: teammate can bootstrap in <10 minutes.

---

## Future

- Handler-level logging and error mapping integration

  - What: Wire a minimal JSON logger using Go `log/slog` in `server/cmd/api` and apply the `MapVerifyError` and `LogAssertion` utilities from `server/internal/webauthn` in the assertion-finish handler. Keep client responses generic while emitting structured, privacy-preserving logs with stable `error_kind` values from sentinel errors.
  - Why: Improves observability, incident triage, and auditability without leaking sensitive data. Cleanly separates transport concerns (HTTP codes) from cryptographic failure semantics via sentinel errors, enabling accurate metrics and alerts (e.g., spikes in `ErrMalformedDER`).
  - How: Initialize `slog` with a JSON handler for dev; in the handler, call `VerifyAssertion(...)`, map the error with `MapVerifyError`, log once via `LogAssertion` using hashed identifiers (`HashID`), and return an appropriate status code with a standard error envelope. This remains compatible with the existing plan’s later steps for endpoints and error envelopes.

- IDNA (punycode) normalization support
  - What: Normalize internationalized domain names to ASCII (punycode) for RP ID and origin comparisons using `golang.org/x/net/idna`.
  - Why: Prevent mismatches and policy bypass due to Unicode vs punycode inconsistencies; ensure consistent hashing for rpIdHash and accurate origin validation across i18n domains.
  - How: Add a deterministic normalization helper that converts Unicode hostnames to punycode before validation/comparison; guard with unit tests and vectors (e.g., `bücher.ch` ⇄ `xn--bcher-kva.ch`), and document deployment guidance to keep config values consistent.

## Acceptance Checks

- **Build**: `go build ./server/...` and `npm run build` in `/web` both succeed.
- **Register/Login**: Using macOS with Touch ID, both ceremonies prompt for fingerprint; login sets cookie.
- **Sign**: Create 2 messages with nonces 1 and 2; both appear in `/tx/list` and DB.
- **Ephemeral TTLs**: Registration/login/tx option sessions expire after 5 minutes; expired attempts return 409.
- **Replay/Nonce**: Re-submit nonce 2 → **409** conflict.
- **UV check**: If browser returns an assertion without UV (simulate by forcing options incorrectly) → **403** forbidden.
- **Origin/RP guard**: Change `origin` in request body → **403**.
- **Low‑S enforced**: Hand-craft a signature with high‑S (unit test) → **400**.

## Coverage & Refs (Traceability)

- R-FLOW-REG: Steps 11–17, 26, 38. Refs: requirement R-FLOW-REG; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails.
- R-FLOW-LOGIN: Steps 11–14, 18–19, 24, 26, 38–39. Refs: requirement R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations.
- R-FLOW-SIGN: Steps 20–23, 32–33, 26, 38. Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails.
- R-ID-KEY: Steps 7–8, 16–17, 19, 22–23. Refs: requirement R-ID-KEY; decision webauthn-corrections-and-standardizations.
- R-SCHEMA-LITE: Steps 10, 20, 32. Refs: requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails.
- R-UI-2BTN: Steps 27, 31–34. Refs: requirement R-UI-2BTN.
- R-PLAT-1: Steps 4, 27–34. Refs: requirement R-PLAT-1.
- R-PLAT-2: Steps 5, 11–26, 38–40. Refs: requirement R-PLAT-2.
- R-PLAT-3: Steps 6, 17, 19, 22–23, 34–35. Refs: requirement R-PLAT-3.
- R-OPS-DEV: Steps 5, 26, 41, 44–45. Refs: requirement R-OPS-DEV.
- R-ERR: Steps 25, 34, 38, 40. Refs: requirement R-ERR; decision encoding-and-ceremony-guardrails.
- R-NO-BROKER: Entire plan avoids brokers; synchronous calls. Refs: requirement R-NO-BROKER.
- R-PORTABLE: Steps 26, 27–34, 37. Refs: requirement R-PORTABLE.
- R-SEC-UV: Steps 14–15, 17–19, 21–22. Refs: requirement R-SEC-UV; decision webauthn-corrections-and-standardizations.
