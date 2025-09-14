### Step 33 — Dashboard: transaction signing (Done: 2025-09-10)

Verification notes
- Backend: `/tx/signing/options` and `/tx/signing/finish` mounted with a shared `TxSessionStore`; server build OK.
- Frontend: Dashboard Sign flow wired (options → get → finish), clears inputs and refreshes list; build OK.
- Tests: Added `web/tests/tx-signing.spec.ts` (mocked happy path), `web/tests/tx-webauthn.spec.ts` (builder unit for `buildTxFinish`), and `web/tests/tx-signing-negative.spec.ts` (401/409/400 coverage). All pass locally.

33. **Transaction signing options**

    - Scope:

      - Use the built `bundle_cbor_b64` (Step 32) to initiate a signing session: POST `/tx/signing/options` (authenticated), receive WebAuthn request options, call `navigator.credentials.get`, then POST `/tx/signing/finish` to persist the transaction and refresh the Dashboard list.
      - Ensure absolute URLs and `credentials: 'include'` are used so cookies are sent/accepted under CORS.

    - Source to add/modify:

      - Modify: `web/src/pages/Dashboard.tsx` — Add a “Sign” button that:
        1. Validates `bundle_cbor_b64` is present (built in Step 32)
        2. `POST /tx/signing/options` with `{ bundle_cbor_b64 }` using `mode: 'cors'`, `credentials: 'include'`
        3. Transforms response to `PublicKeyCredentialRequestOptions` via existing `toRequestOptions` (Step 30)
        4. Calls `navigator.credentials.get({ publicKey })`
        5. Builds and sends finish payload to `/tx/signing/finish` using `credentials: 'include'`, then clears inputs and calls `loadList()` to refresh
      - Add (frontend lib): `web/src/lib/webauthn.ts` — `buildTxFinish(cred: PublicKeyCredential, txSessionId: string)` to mirror login finish but with `tx_session_id` name.
      - Modify (backend wiring): `server/cmd/api/main.go` — Ensure `/tx/signing/options` and `/tx/signing/finish` are mounted with a shared `txStore := tx.NewTxSessionStore(10000)`.

    - Description files (create/update alongside code changes):

      - Update: `web/src/pages/Dashboard.tsx.desc.md` — Add signing flow details (options→get→finish, refresh behavior, errors). Refs included.
      - Update: `web/src/lib/webauthn.ts.desc.md` — Document `buildTxFinish` behavior and how it differs only in the session id field name.
      - Update: `server/cmd/api/main.go.desc.md` — Note that `/tx/signing/*` are mounted and share an in-memory `TxSessionStore`.

    - Request/response shape:

      - Request (options):
        - `POST {API_BASE}/tx/signing/options`
        - Headers: `Content-Type: application/json`
        - Body: `{ "bundle_cbor_b64": string }`
        - Cookie: `sid` (via `credentials: 'include'`)
      - Response (options):
        - JSON:
          - `tx_session_id: string`
          - `challenge: string` (base64url)
          - `tx_id_hex: string`
          - `options: { rp_id: string; origin: string; uv_required: boolean; allow_credentials: string[] }`
          - `expires_at: number`
      - WebAuthn input (browser):
        - `PublicKeyCredentialRequestOptions` with `challenge`, optional `allowCredentials`, and `userVerification: 'required'`.
      - Request (finish):
        - `POST {API_BASE}/tx/signing/finish`
        - Body JSON:
          - `tx_session_id: string`
          - `id: string`, `rawId: string` (base64url)
          - `type: 'public-key'`
          - `response: { authenticatorData: string (base64url), clientDataJSON: string (base64url), signature: string (base64url), userHandle?: string (base64url) }`
        - Cookie: `sid` (via `credentials: 'include'`)
      - Response (finish):
        - `200 OK` JSON `{ tx_id_hex: string, stored: boolean }`

    - Algorithm:

      - Sign action (Dashboard):
        - Guard: if `bundle_cbor_b64` is empty → show validation message (ask user to Build in Step 32 first).
        - Options: `const r = await fetch(apiUrl('/tx/signing/options'), { method: 'POST', mode: 'cors', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ bundle_cbor_b64 }) })`.
        - If 401 → show “Not logged in. Please Login.”. If 409 → show error (nonce policy or no credentials). If 400 → show “Invalid bundle”.
        - On 200: `const data = await r.json()`; `const publicKey = toRequestOptions(data)`; call `navigator.credentials.get({ publicKey })`.
        - Finish payload: `const in = buildTxFinish(cred, data.tx_session_id)`; POST to `/tx/signing/finish` with `mode: 'cors'`, `credentials: 'include'`, `Content-Type: 'application/json'`.
        - On 200: clear `msg`, `nonce`, `bundle_b64`, `bundle_hex`; call `loadList()` to refresh the table; optionally display `tx_id_hex` returned.
        - On errors: render inline status (unauthorized/forbidden/conflict/bad request) and do not clear inputs.

    - Database interactions:

      - None on client. Server validates and persists the transaction (nonce monotonicity, account binding, signature verification) and updates `credentials.sign_count`.

    - Policies & limits:

      - `allowCredentials` must include the account’s credential ids (server supplies); client must pass it through; UV required.
      - Bundle determinism is assumed from Step 32; if server returns 400 invalid bundle, the client should prompt a rebuild.
      - Use absolute URLs + `credentials: 'include'` for both options and finish.

    - Sequencing:

      - Depends on Step 32 (bundle build) and Step 30 (session cookie).
      - After success, Step 31’s list refresh shows the new entry.

    - Tests (happy path required, plus negatives):

      - Files to add: `web/tests/tx-signing.spec.ts`
        - Happy path (mocked):
          - Mock `/tx/signing/options` to return a valid response with `tx_session_id`, `challenge`, and `allow_credentials`.
          - Stub `navigator.credentials.get` to return a fake PublicKeyCredential object.
          - Assert that `buildTxFinish` constructs payload with `tx_session_id` and base64url fields, and that the UI calls finish and then refreshes the list.
        - Negative:
          - Options: 401 Unauthorized → UI shows prompt; 409 Conflict (nonce) → error banner; 400 Invalid bundle → validation prompt to rebuild.
          - Finish: 409 (signCount) or 401 (allowlist/cred mismatch) → error banner.
      - Commands:
        - `npm -C web run test:ui -- --project=chromium tests/tx-signing.spec.ts`

    - Verification:

      - Manual end-to-end:

        - Build bundle (Step 32) with nonce N; click Sign; approve WebAuthn prompt; expect 200 finish with `tx_id_hex` and the list updates showing the new row.
        - Build again with nonce N (same as previous) → options succeeds, finish returns 409 on signCount or server rejects nonce policy, and the UI shows an error.

      - User verification commands:

        ```bash
        # 1) Start backend and web
        RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C server run ./cmd/api &
        SERVER_PID=$!
        npm -C web run dev &
        WEB_PID=$!

        # 2) Register → Login → Dashboard → Build bundle → Sign → Confirm new entry in list

        # 3) Cleanup
        kill $WEB_PID || true
        kill $SERVER_PID || true
        ```

    - Acceptance criteria:

      - Dashboard sends `bundle_cbor_b64` to `/tx/signing/options`, calls `navigator.credentials.get` with decoded `challenge` and allowlist, and sends finish to `/tx/signing/finish`.
      - On success, the list refreshes and shows the new transaction with nonce and timestamp.
      - On policy failures (non-logged-in, nonce conflict, allowlist/cred mismatch), the UI renders an informative inline status without crashing.

    - Notes:

      - We reuse existing WebAuthn helpers: `toRequestOptions` for request options and a new `buildTxFinish` mirroring the login finish payload with `tx_session_id`.
      - Avoid reading cookies in JS; rely on HTTP statuses and responses.

    - Appendix:

    - Refs:

      - Refs: requirement R-FLOW-SIGN; requirement R-PLAT-1; requirement R-PLAT-2; requirement R-PORTABLE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer

