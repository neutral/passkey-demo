# Step 1 — Web: adopt `@simplewebauthn/browser` for Register/Login

Scope

- Replace local WebAuthn builders with `startRegistration` and `startAuthentication` from `@simplewebauthn/browser`.
- Keep base64url helpers; migrate to the library’s JSON option shapes to remove manual ArrayBuffer transforms.
- Update `Register.tsx` and `Login.tsx` to call the new functions; preserve UI and error toast behavior.

Source to add/modify

- Add: `web/package.json` dependency `@simplewebauthn/browser@^10`.
- Modify: `web/src/pages/Register.tsx` — fetch registration options (Go shape), convert to `PublicKeyCredentialCreationOptionsJSON`, call `startRegistration(optionsJSON)`, then POST finish with the returned `RegistrationResponseJSON` plus `reg_session_id`.
- Modify: `web/src/pages/Login.tsx` — fetch login options (Go shape), convert to `PublicKeyCredentialRequestOptionsJSON`, call `startAuthentication(optionsJSON)`, then POST finish with the returned `AuthenticationResponseJSON` plus `login_session_id`; include `credentials: 'include'`.
- Modify: `web/src/lib/webauthn.ts` — remove manual builders `toCreationOptions`, `toRequestOptions`, `buildRegFinish`, `buildLoginFinish`.
- Add: `web/src/lib/webauthn.ts` — thin compatibility adapters to bridge current Go options to the lib JSON shapes and normalize errors:
  - `toCreationOptionsJSON(resp: RegistrationOptionsResponse): PublicKeyCredentialCreationOptionsJSON`
  - `toRequestOptionsJSON(resp: LoginOptionsResponse): PublicKeyCredentialRequestOptionsJSON`
  - `mapDomException(e: unknown): { title: string; detail: string }` (wraps `DOMException`/`NotAllowedError` to align with `normalizeError`)
- Update tests: replace builder-focused tests with adapter/page tests that assert correct invocation and finish payloads.

Description files

- Update: `web/src/pages/Register.tsx.desc.md` — document use of `@simplewebauthn/browser` and finish body composition with `reg_session_id`. Refs updated.
- Update: `web/src/pages/Login.tsx.desc.md` — same, including cookie handling (`credentials: 'include'`). Refs updated.
- Update: `web/src/lib/webauthn.ts.desc.md` — remove builder roles; describe adapters to JSON option shapes and error normalization. Refs updated.
- Update: `web/web.desc.md` — note adoption of library and JSON shapes through adapters. Refs updated.
- Add: test descriptions for each new/renamed spec file (below), e.g., `web/tests/register-adapter.spec.ts.desc.md` and `web/tests/login-post-body.spec.ts.desc.md` summarizing purpose and Refs.

Request/response shape

- Options (server → browser):
  - Current Go options (input to adapters):
    - Registration: `{ reg_session_id: string, challenge: base64url, options: { rp_id: string, origin: string, uv_required: true, attestation: 'none' }, expires_at: number }`
    - Login: `{ login_session_id: string, challenge: base64url, options: { rp_id: string, origin: string, uv_required: true, allow_credentials: string[] }, expires_at: number }`
  - Library JSON (adapters output):
    - `PublicKeyCredentialCreationOptionsJSON`: `{ rp: { id: string, name: 'Passkey Demo' }, user: { id: base64url, name: 'demo', displayName: 'Demo' }, challenge: base64url, pubKeyCredParams: [{ type: 'public-key', alg: -7 }], authenticatorSelection: { residentKey: 'required', userVerification: 'required' }, attestation: 'none' }`
    - `PublicKeyCredentialRequestOptionsJSON`: `{ challenge: base64url, rpId?: string, userVerification: 'required', allowCredentials?: [{ type: 'public-key', id: base64url }] }`
- Finish (browser → server):
  - Registration: `RegistrationResponseJSON & { reg_session_id: string }` (fields `id`, `rawId`, `response.attestationObject`, `response.clientDataJSON` are base64url strings)
  - Login: `AuthenticationResponseJSON & { login_session_id: string }` (fields `id`, `rawId`, `response.authenticatorData`, `response.clientDataJSON`, `response.signature`, `response.userHandle` [optional] are base64url strings)

Algorithm

- Register:
  - POST `/authn/passkey/registration/options` (CORS, JSON) → parse Go response.
  - Build `optionsJSON = toCreationOptionsJSON(resp)` ensuring policy: residentKey: `required`, userVerification: `required`, attestation: `none`.
  - Await `att = startRegistration(optionsJSON)` from `@simplewebauthn/browser`.
  - POST `/authn/passkey/registration/finish` with `{ ...att, reg_session_id: resp.reg_session_id }`.
  - On `!ok`, parse error via `parseHttpError`; on thrown `DOMException`, normalize via `mapDomException` then `normalizeError` for toast.
- Login:
  - POST `/authn/passkey/login/options` → parse Go response.
  - Build `optionsJSON = toRequestOptionsJSON(resp)`; omit `allowCredentials` when empty; set `userVerification: 'required'` and `rpId` if provided.
  - Await `asg = startAuthentication(optionsJSON)`.
  - POST `/authn/passkey/login/finish` with `{ ...asg, login_session_id: resp.login_session_id }`, `credentials: 'include'`.
  - Handle errors as in Register.

Database interactions

- None in this step (frontend-only change).

Policies & limits

- Enforce in options JSON: `residentKey: 'required'`, `userVerification: 'required'`, `attestation: 'none'` (registration).
- Preserve CORS usage per `frontend-api-base-and-cors` (absolute API URLs; include credentials on login finish).
- Verify via adapter unit tests that policy flags are present; verify via E2E that flows succeed or surface expected policy/status errors.

Sequencing

- Compatible with existing Go backend: adapters convert current Go option shapes to library JSON; finish payloads match server expectations.
- Node server later (Steps 5–8) will emit library JSON directly; at that time remove the adapters and call the library with server-provided JSON.
- Replace builder-centric tests in this step to avoid breakage when removing `toCreationOptions`/`toRequestOptions`.

Tests

- Strategy: keep browser-level E2E, add focused adapter unit tests, and add page-level POST-body assertions to ensure finish payloads remain correct. Happy path is REQUIRED; include negative cases and mapping checks.
- Add/modify test files:
  - Add `web/tests/register-adapter.spec.ts`: unit-test `toCreationOptionsJSON`
    - Asserts: rp.id matches, attestation 'none', residentKey 'required', `challenge` remains base64url string, user.id base64url 32 bytes.
  - Add `web/tests/login-adapter.spec.ts`: unit-test `toRequestOptionsJSON`
    - Asserts: `rpId` set when provided, `userVerification: 'required'`, `allowCredentials` omitted when empty, present with 1 entry when provided.
  - Add `web/tests/register-post-body.spec.ts`: page test
    - Stub options endpoint with Go shape; stub `navigator.credentials.create` to return a fake `PublicKeyCredential` object; intercept `/registration/finish` and validate posted JSON includes `reg_session_id`, base64url strings for `rawId` and `response.attestationObject/clientDataJSON`, and `Content-Type: application/json`.
  - Add `web/tests/login-post-body.spec.ts`: page test
    - Similar to register; ensure body includes `login_session_id`, base64url strings, and request uses `credentials: 'include'`.
  - Update `web/tests/error-toasts.spec.ts`: keep stubbing of `navigator.credentials` and network; ensure `DOMException('NotAllowedError', ...)` bubbles into toast via `mapDomException`/`normalizeError`.
  - Remove/rewrite builder-centric tests:
    - Replace `web/tests/webauthn.spec.ts` → with `register-adapter.spec.ts`.
    - Replace `web/tests/login-webauthn.spec.ts` → with `login-adapter.spec.ts`.
- Commands:
  - `npm -C web i`
  - `npm -C web run test:ui`
- Expected outcomes:
  - All adapter tests green; POST-body tests assert shapes and session ids; existing Chromium Virtual Authenticator E2E remain passing or surface acceptable NotAllowed/401 errors as documented.

Verification

- Unit + page tests:
  - Run `npm -C web i && npm -C web run test:ui` until green.
- Manual (Go server compatibility):
  - Start Go server and web, then perform register/login manually to confirm flows and toasts remain correct.
- User verification commands

```bash
# Install web deps (adds @simplewebauthn/browser)
npm -C web i

# Run Playwright tests (unit + page + existing E2E)
npm -C web run test:ui

# Manual check against Go server (in separate terminals)
make server
npm -C web run dev
# In the browser: Register → Login. Expect success UI or policy/HTTP toasts as before.
```

Acceptance criteria

- Register/Login operate via `@simplewebauthn/browser` with no manual ArrayBuffer transforms.
- Adapters produce valid library JSON option shapes while Go backend is in use; finish payloads include the appropriate `*_session_id` and match server expectations.
- Tests added/updated as listed; happy path and negative mapping cases covered; all web tests pass locally.

Notes

- This step updates only the frontend and tests; the backend remains the existing Go server. Adapters are temporary and will be removed once the Node server emits library JSON directly (Steps 5–8).
- Keep description files in sync for all modified/added files; PR will be blocked otherwise.

Refs: goal passkey-registration-login-uv; requirement R-PLAT-1; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors

