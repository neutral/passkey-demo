### Step 30 — WebAuthn get() flow (Login) (Done: 2025-09-05)

Verification notes
- Builders: Playwright tests pass for `toRequestOptions` and `buildLoginFinish` (including omission of allowCredentials when empty).
- Optional Chromium E2E (Virtual Authenticator): passes; accepts NotAllowed/401 when DB unseeded; success path shows thumb.
- Manual E2E: After registration, login succeeds (200), `sid` cookie is set (HttpOnly; verify in DevTools), routes to Dashboard.

30. **WebAuthn get() flow (Login)**

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

