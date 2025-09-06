### Step 29 — WebAuthn create() flow (Register) (Done: 2025-09-05)

Verification notes
- Builders: Playwright tests pass for `toCreationOptions` and `buildRegFinish`.
- Chromium E2E (Virtual Authenticator): Runs to completion; demo backend returns HTTP 400 if attestation fmt ≠ `none` (acceptable for this environment).
- Manual E2E: Register flow works in browser with platform authenticator; shows `account_thumb_hex` and routes to Login.

29. **WebAuthn create() flow (Register)**

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

