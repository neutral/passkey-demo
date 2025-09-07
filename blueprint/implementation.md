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

## Next

### Phase H — Frontend (React) UI

---

30. **WebAuthn get() flow (Login)** — Moved to Done

    - Scope:

      - Implement the browser login (assertion) flow on the Login page using absolute API URLs and CORS.
      - Transform server JSON into `PublicKeyCredentialRequestOptions` (b64url → ArrayBuffer; allowCredentials decoding) and submit `navigator.credentials.get({ publicKey })`.
      - Send `login/finish` payload (ArrayBuffers → base64url) to the backend with `credentials: 'include'` so the browser accepts the `Set-Cookie` response header.
      - On success: show `account_thumb_hex`, set a local “logged in” UI state, and route to Dashboard (no global store).

    - Source to add/modify:

      - Add: `web/src/lib/webauthn.ts` — Extend with assertion helpers:
        - `toRequestOptions(resp: { login_session_id: string; challenge: string; options: { rp_id: string; origin: string; uv_required: boolean; allow_credentials: string[] } }): PublicKeyCredentialRequestOptions` — builds `publicKey` options with `challenge`, optional `allowCredentials` list of `{ type:'public-key', id: ArrayBuffer }`, and `userVerification: 'required'`. Optionally set `rpId = resp.options.rp_id`. If the decoded allow list is empty, omit the `allowCredentials` property entirely to allow discoverable credentials.
        - `buildLoginFinish(cred: PublicKeyCredential, loginSessionId: string)` — returns JSON with base64url-encoded `rawId`, `response.authenticatorData`, `response.clientDataJSON`, `response.signature`, and optional `response.userHandle`, plus `login_session_id`.
      - Modify: `web/src/pages/Login.tsx` — Wire fetch options → build request options → `navigator.credentials.get` → POST finish with `credentials: 'include'` → on success show `account_thumb_hex` and route to Dashboard (e.g., set `location.hash = '#/dashboard'`). Render inline error on failure.
      - Add (tests): `web/tests/login-webauthn.spec.ts` — Pure builder tests for `toRequestOptions` and `buildLoginFinish` (no real WebAuthn); run under Playwright.
      - Add (optional chromium e2e): `web/tests/login-e2e.chromium.spec.ts` — Uses Virtual Authenticator. Note: Without a pre‑registered credential in the DB, finish returns 401 (credential not recognized); treat that as acceptable for this environment unless the DB is seeded via a prior manual registration.

    - Description files (create/update alongside code changes):

      - Update: `web/src/lib/webauthn.ts.desc.md` — Document `toRequestOptions` and `buildLoginFinish` semantics (UV required; allowCredentials handling; encoding rules). Refs included.
      - Update: `web/src/pages/Login.tsx.desc.md` — Add flow details, network calls, transformations, cookie receipt via CORS, and success UI. Refs included.
      - Update: `web/web.desc.md` — Mention login flow wiring and `credentials: 'include'` for cookies.

    - Request/response shape:

      - Request (options):
        - `POST {API_BASE}/authn/passkey/login/options`
        - Headers: `Content-Type: application/json` (no body)
      - Response (options):
        - JSON:
          - `login_session_id: string`
          - `challenge: string` (base64url)
          - `options: { rp_id: string; origin: string; uv_required: boolean; allow_credentials: string[] }`
          - `expires_at: number` (unix seconds)
      - WebAuthn input (browser):
        - `PublicKeyCredentialRequestOptions` with `challenge: BufferSource`, optional `allowCredentials: { type:'public-key', id: BufferSource }[]`, and `userVerification: 'required'`. Optionally include `rpId`.
      - Request (finish):
        - `POST {API_BASE}/authn/passkey/login/finish`
        - Body JSON:
          - `login_session_id: string`
          - `id: string` (from `credential.id`)
          - `rawId: string` (base64url)
          - `type: 'public-key'`
          - `response: { authenticatorData: string (base64url), clientDataJSON: string (base64url), signature: string (base64url), userHandle?: string (base64url) }`
      - Response (finish):
        - `200 OK` JSON `{ account_thumb_hex: string, credential_id_b64: string }`
        - Plus `Set-Cookie: sid=...; HttpOnly; SameSite=Lax; Secure?` (Secure=false on `http://localhost`). Note: HttpOnly cookies are not readable via `document.cookie`; verify in DevTools Application → Cookies or via subsequent authenticated calls.

    - Algorithm:

      - Fetch options:
        - `const r = await fetch(apiUrl('/authn/passkey/login/options'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, mode: 'cors' })`.
        - Parse JSON; optionally validate `origin === location.origin` and `options.rp_id` expected in dev.
      - Build request options:
        - `challenge = base64urlToBytes(resp.challenge)`.
        - If `allow_credentials` present: map each to `{ type:'public-key', id: base64urlToBytes(b64) }`.
        - If the mapped list is empty, omit `allowCredentials` from the options object (do not send an empty array) to enable discoverable credentials.
        - `userVerification = 'required'`; optionally set `rpId = resp.options.rp_id`.
      - Call WebAuthn get:
        - `const cred = await navigator.credentials.get({ publicKey }) as PublicKeyCredential`.
      - Build finish payload:
        - `rawId = bytesToBase64url(new Uint8Array(cred.rawId as ArrayBuffer))`.
        - `authenticatorData = bytesToBase64url(cred.response.authenticatorData)`.
        - `clientDataJSON = bytesToBase64url(cred.response.clientDataJSON)`.
        - `signature = bytesToBase64url(cred.response.signature)`.
        - `userHandle = cred.response.userHandle ? bytesToBase64url(cred.response.userHandle) : ''`.
        - Include `id`, `type`, and `login_session_id` from options.
      - POST finish (with cookie acceptance):
        - `await fetch(apiUrl('/authn/passkey/login/finish'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, mode: 'cors', credentials: 'include', body: JSON.stringify(payload) })`.
      - Success/UI:
        - Show `account_thumb_hex` and route to Dashboard (`location.hash = '#/dashboard'`).
      - Errors:
        - Inline error text on HTTP failure or thrown WebAuthn errors (AbortError/NotAllowedError). Full toasts in Step 34.

    - Database interactions:

      - None on client; server verifies assertion, enforces signCount increasing, updates DB, and sets a `sid` session cookie.

    - Policies & limits:

      - UV required across options and verification.
      - `allowCredentials` may be empty for discoverable credentials; builder must handle both cases.
      - Absolute API URLs only; no proxy. Ensure `API_BASE` is an absolute URL (do NOT set `/api`). Use `mode: 'cors'` and `credentials: 'include'` on finish so cookies are accepted.
      - HttpOnly cookie: do not attempt to read `sid` in JS; validate via DevTools or by calling an authenticated endpoint (Step 31).
      - Do not persist auth state in localStorage; rely on `sid` cookie.

    - Sequencing:

      - Depends on Steps 18 (login/options), 19 (login/finish), 26 (CORS & cookies), and 28 (encoding helpers).
      - Follows Step 29 (registration); precedes Step 31 (Dashboard list) and later signing flows.

    - Tests (happy path required, plus negatives):

      - Files to add:
        - `web/tests/login-webauthn.spec.ts`
        - `web/tests/login-e2e.chromium.spec.ts` (optional; see note below)
      - Happy path (pure builder tests):
        - `toRequestOptions` sets `userVerification='required'`, decodes `challenge` to bytes, and maps `allow_credentials` to `allowCredentials[].id: ArrayBuffer`.
        - When `allow_credentials` is empty, the returned options object omits `allowCredentials`.
        - `buildLoginFinish` correctly base64url-encodes `rawId`, `authenticatorData`, `clientDataJSON`, and `signature`, includes `login_session_id`.
      - Negative cases:
        - Invalid base64url in `challenge` or `allow_credentials` → builder throws `TypeError`.
        - Missing required fields in response → builder throws clearly.
      - Automated E2E (Chromium + Virtual Authenticator; optional):
        - Without a pre-registered credential in the DB, finish returns 401 (credential not recognized) by design — treat as acceptable signal the flow executed end-to-end client-side.
        - If you first complete registration manually (or seed the DB), the same test can assert 200 and cookie presence via `document.cookie` checks.
        - Command: `npm -C web run test:ui -- --project=chromium tests/login-e2e.chromium.spec.ts`.

    - Verification:

      - Manual E2E:
        - Register first using the Register page. Then go to Login, click “Start Login”, approve the prompt.
        - Expect `200 OK` with `{ account_thumb_hex, credential_id_b64 }`. Confirm cookie in DevTools Application tab → Cookies for `http://localhost:5173` (HttpOnly cookie not visible to `document.cookie`).
        - App routes to Dashboard.
      - Builder tests pass in Playwright.

      - User verification commands:

        ```bash
        # 1) Install deps
        npm -C web ci || npm -C web install

        # 2) Run builder tests
        npm -C web run test:ui -- --project=chromium tests/login-webauthn.spec.ts

        # 3) Optional E2E without registration seed (expects 401 acceptable)
        npm -C web run test:ui -- --project=chromium tests/login-e2e.chromium.spec.ts || true

        # 4) Manual E2E
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C server run ./cmd/api &
        SERVER_PID=$!
        npm -C web run dev &
        WEB_PID=$!
        # Visit http://localhost:5173 → Register, then Login. Check cookie.
        kill $WEB_PID || true
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - Login page performs options → get → finish with absolute URLs; success displays `account_thumb_hex` and routes to Dashboard.
      - Finish request uses `credentials: 'include'` and cookie is set on success (SameSite=Lax, HttpOnly; Secure=false for `http://localhost`).
      - `toRequestOptions` and `buildLoginFinish` builders pass Playwright tests.

    - Notes:

      - `allowCredentials` filtering may be empty for discoverable credentials; this is expected for platform authenticators.
      - Some virtual authenticator configurations might not reflect production prompts; keep manual E2E as source of truth for UX.

    - Refs:
      - Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-1; requirement R-PORTABLE; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec spec-a; spec spec-b

---

31. **Dashboard: fetch list**

    - Call `GET /tx/list` with `credentials: 'include'` and absolute API URL; render in table; handle empty state gracefully.
    - _Verify_: list renders (empty or with prior entries) and updates on refresh.

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
