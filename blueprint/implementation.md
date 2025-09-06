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

## Next

### Phase H — Frontend (React) UI

---

29. **WebAuthn create() flow (Register)** — Moved to Done

    - Scope:

      - Implement the browser registration flow end-to-end on the Register page using absolute API URLs and CORS (no Vite proxy).
      - Transform server options into `PublicKeyCredentialCreationOptions` (b64url → ArrayBuffer) and submit `navigator.credentials.create({ publicKey })`.
      - Send `registration/finish` payload (ArrayBuffers → base64url) to the backend. On success, display the `account_thumb_hex` and route to Login.

    - Source to add/modify:

      - Add: `web/src/lib/webauthn.ts` — Helpers for registration flow:
        - `toCreationOptions(resp: { reg_session_id: string; challenge: string; options: { rp_id: string; origin: string; uv_required: boolean; attestation: string } }): PublicKeyCredentialCreationOptions` — builds `publicKey` options; sets `rp.id`, `challenge`, `attestation`, `authenticatorSelection.residentKey='required'`, `authenticatorSelection.userVerification='required'`, `pubKeyCredParams=[{ type:'public-key', alg:-7 }]`, `rp.name='Passkey Demo'`, and ephemeral `user` (random 32-byte id, name/displayName 'demo').
        - `buildRegFinish(cred: PublicKeyCredential, regSessionId: string)` — returns JSON matching server’s expected shape with base64url strings (id/rawId/attestationObject/clientDataJSON) and `reg_session_id`.
        - Uses `bytesToBase64url`/`base64urlToBytes` from `web/src/lib/encoding.ts`.
      - Modify: `web/src/pages/Register.tsx` — Wire the UI:
        - “Start Registration” button triggers: fetch options → build options → `navigator.credentials.create` → POST finish → show `account_thumb_hex` and a button “Go to Login”.
        - Use absolute URLs via `apiUrl('/authn/passkey/registration/options|finish')`. No proxy. Handle loading and basic error text inline.
      - Add (tests): `web/tests/webauthn.spec.ts` — Pure tests for `toCreationOptions` and `buildRegFinish` (no real WebAuthn); run under Playwright.
      - Add (chromium e2e): `web/tests/webauthn-e2e.chromium.spec.ts` — Automated E2E using a Virtual Authenticator via CDP; provisions a CTAP2 authenticator with resident keys and UV enabled, then performs Register flow end-to-end.
      - Modify (tests config): `web/playwright.config.ts` — Add a second `webServer` entry to start the Go backend for E2E (`RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ../server/cmd/api`), alongside the existing Vite dev server.

    - Description files (create/update alongside code changes):

      - Create: `web/src/lib/webauthn.ts.desc.md` — Describe builders, invariants (UV/residentKey/attestation), and encoding rules. Refs included.
      - Update: `web/src/pages/Register.tsx.desc.md` — Add flow details, network calls, transformations, and success UI. Refs included.
      - Update: `web/web.desc.md` — Note WebAuthn helpers usage and absolute API under CORS.

    - Request/response shape:

      - Request (options):
        - `POST {API_BASE}/authn/passkey/registration/options`
        - Headers: `Content-Type: application/json` (no body)
      - Response (options):
        - JSON:
          - `reg_session_id: string`
          - `challenge: string` (base64url)
          - `options: { rp_id: string; origin: string; uv_required: boolean; attestation: 'none' }`
          - `expires_at: number` (unix seconds)
      - WebAuthn input (browser):
        - `PublicKeyCredentialCreationOptions` with binary fields as `ArrayBuffer`/`BufferSource`.
      - Request (finish):
        - `POST {API_BASE}/authn/passkey/registration/finish`
        - Body JSON:
          - `reg_session_id: string`
          - `id: string` (as received from `credential.id`)
          - `rawId: string` (base64url)
          - `type: 'public-key'`
          - `response: { attestationObject: string (base64url), clientDataJSON: string (base64url) }`
      - Response (finish):
        - `201 Created` JSON `{ account_thumb_hex: string, credential_id_b64: string }`

    - Algorithm:

      - Fetch options:
        - `const r = await fetch(apiUrl('/authn/passkey/registration/options'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, mode: 'cors' })`.
        - Parse JSON; optionally validate `origin === location.origin` and `options.rp_id === 'localhost'` when in dev; warn if mismatch.
      - Build `publicKey` options:
        - `challenge = base64urlToBytes(resp.challenge)`.
        - `rp = { id: resp.options.rp_id, name: 'Passkey Demo' }`.
        - `user = { id: crypto.getRandomValues(new Uint8Array(32)), name: 'demo', displayName: 'Demo' }` (demo-only; stable identity is server-side passkey key).
        - `pubKeyCredParams = [{ type: 'public-key', alg: -7 }]` (ES256).
        - `authenticatorSelection = { residentKey: 'required', userVerification: 'required' }`.
        - `attestation = 'none'`.
      - Call WebAuthn create:
        - `const cred = await navigator.credentials.create({ publicKey }) as PublicKeyCredential`.
      - Build finish payload:
        - `rawId = bytesToBase64url(new Uint8Array(cred.rawId as ArrayBuffer))`.
        - `attestationObject = bytesToBase64url(cred.response.attestationObject)`.
        - `clientDataJSON = bytesToBase64url(cred.response.clientDataJSON)`.
        - Include `id`, `type`, and `reg_session_id` from options.
      - POST finish:
        - `await fetch(apiUrl('/authn/passkey/registration/finish'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, mode: 'cors', body: JSON.stringify(payload) })`.
      - Success/UI:
        - Show `account_thumb_hex` with monospace styling; provide a “Go to Login” button that routes to Login.
      - Errors:
        - Show inline error text if HTTP not ok or JSON parse fails; if `navigator.credentials.create` rejects (AbortError, NotAllowedError), show the message. Full toasts come in Step 34.

    - Database interactions:

      - None on client; server persists account and credential as per Step 17.

    - Policies & limits:

      - Attestation: `'none'`; ResidentKey: `'required'`; UserVerification: `'required'` — must be reflected in the built options.
      - Algorithm: ES256 (`alg: -7`) only.
      - Absolute URLs only; do not use a proxy; `mode: 'cors'` on fetch.
      - Do not store any secrets or session state in local storage; rely on server session.

    - Sequencing:

      - Depends on Steps 15 (reg/options), 17 (reg/finish), 26 (CORS), and 28 (encoding helpers).
      - Precedes Step 30 (login/assertion) and later dashboard steps.

    - Tests (happy path required, plus negatives):

      - Files to add:
        - `web/tests/webauthn.spec.ts`
        - `web/tests/webauthn-e2e.chromium.spec.ts`
      - Happy path (pure builder tests):
        - Verify `toCreationOptions` maps `rp_id` to `rp.id`, sets `attestation='none'`, and `authenticatorSelection.residentKey='required'`, `userVerification='required'`.
        - Challenge is an `ArrayBuffer` with expected byte length (32) after decoding the provided base64url.
        - `pubKeyCredParams` contains `{ type:'public-key', alg:-7 }`.
        - `buildRegFinish` correctly base64url-encodes provided byte fields and includes `reg_session_id`.
      - Negative cases:
        - Invalid base64url challenge → builder throws `TypeError`.
        - Missing required fields in response → builder throws with a clear message.
      - Automated E2E (Chromium + Virtual Authenticator):
        - Enable CDP WebAuthn and add a virtual authenticator with options:
          - `protocol: 'ctap2'`, `transport: 'internal'`, `hasResidentKey: true`, `hasUserVerification: true`, `isUserVerified: true`, `automaticPresenceSimulation: true`.
        - Start backend (Playwright `webServer` entry) and Vite dev server; navigate to Home → Register → click “Start Registration”.
        - Expect 201 finish with `{ account_thumb_hex, credential_id_b64 }` and success UI.
        - Command: `npm -C web run test:ui -- --project=chromium tests/webauthn-e2e.chromium.spec.ts`.
      - Commands:
        - `npm -C web run test:ui` (Playwright runner executes `webauthn.spec.ts`).

    - Verification:

      - Manual E2E (for platform authenticators):
        - Start backend and dev server; open Register; click “Start Registration”; approve platform authenticator prompt.
        - Expect server to return 201 with `{ account_thumb_hex, credential_id_b64 }`; UI shows thumb and a button to go to Login; clicking it routes to the Login page.
      - Automated builder tests pass in Playwright.
      - Automated Chromium E2E passes with Virtual Authenticator (no real prompts). Note: if the virtual authenticator returns an attestation format other than `none` (e.g., `packed`), the demo backend returns HTTP 400 by design; treat that as acceptable for the E2E in this environment.

      - User verification commands:

        ```bash
        # 1) Install deps
        npm -C web ci || npm -C web install

        # 2) Run builder tests
        npm -C web run test:ui --silent

        # 2b) (Optional) Run Chromium E2E with Virtual Authenticator
        # Ensure Playwright config starts both backend and Vite servers
        npm -C web run test:ui -- --project=chromium tests/webauthn-e2e.chromium.spec.ts

        # 3) Run backend and web (manual E2E)
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go run ./server/cmd/api &
        SERVER_PID=$!
        npm -C web run dev &
        WEB_PID=$!

        # 4) Visit http://localhost:5173 → Register, perform create(), observe success thumb

        # 5) Cleanup
        kill $WEB_PID || true
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - Register page performs options → create → finish with absolute URLs; success shows `account_thumb_hex` and allows routing to Login.
      - `toCreationOptions` sets correct flags and transforms base64url challenge to `ArrayBuffer`.
      - `buildRegFinish` encodes all binary fields to base64url and includes `reg_session_id`.
      - Playwright builder tests pass.

    - Notes:

      - User object is demo-only: random 32-byte `user.id`, and simple `name/displayName`; server identity is passkey-first via COSE key (Step 17 + R-ID-KEY).
      - Keep UI minimal; full error toast styling comes in Step 34.

    - Refs:
      - Refs: requirement R-FLOW-REG; requirement R-PLAT-1; requirement R-PORTABLE; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec spec-a; spec spec-b

---

30. **WebAuthn get() flow (Login)**

    - `POST /authn/passkey/login/options`; convert to `PublicKeyCredentialRequestOptions`.
    - Call `navigator.credentials.get(...)`.
    - Send `login/finish` (use absolute API URL and `credentials: 'include'` on fetch to accept `Set-Cookie` under CORS); on success: set “logged in” UI state (no global store, just local state) and route to Dashboard.
    - _Verify_: end-to-end login completes; cookie present.

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
