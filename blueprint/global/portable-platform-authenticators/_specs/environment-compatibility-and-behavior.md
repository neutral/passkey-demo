# R-PORTABLE — Environment Compatibility and Behavior Notes

## Metadata
- Status: Draft
- Date: 2025-09-05
- Owners: passkey-demo maintainers

## Purpose
- Capture practical environment behaviors observed during registration/login with platform authenticators and virtual authenticators across browsers. Provide actionable guidance to avoid common pitfalls in local dev and CI.

## Key Learnings (Summary)
- Attestation formats vary by browser/authenticator. Brave/Chromium may return `fmt: "packed"` even when conveyance is "none"; demo accepts both `none` and `packed` (no trust-chain).
- Credential public key (COSE_Key) encoding can be wrapped (CBOR bstr, tag 24) or use variant integer key types. Robust COSE parsing is required to extract `{1:kty, 3:alg, -1:crv, -2:x, -3:y}`.
- ES256-only: We advertise and verify ES256 (P-256). Chromium warns about missing RS256; this is expected. Allowing RS256 can cause selection of RSA that the server won’t verify.
- Discoverable credentials: Omit `allowCredentials` to enable account discovery. Do not send an empty array — some implementations treat it as “no credentials allowed”.
- Dev networking: Absolute API URLs plus CORS; no Vite proxy. Cookies are HttpOnly + SameSite=Lax; `Secure=false` for `http://localhost`.
- Virtual Authenticator vs platform: CI/E2E flows may return 400 (registration) or 401 (login without seeded DB); acceptable for automation. Manual platform flows validate real UX.

## Browser/Authenticator Differences
- Attestation `fmt`:
  - Safari, Chrome (platform): commonly `none` with empty attStmt for basic registration.
  - Brave/Chromium Virtual Authenticator: may use `packed`; demo should accept this while still extracting attested credential data.
- COSE_Key encoding:
  - Direct map works on many platforms.
  - Some implementations wrap the map as a CBOR byte string or as Tag 24 (encoded CBOR). Parser should unwrap and decode.
  - Map keys may decode as int/int64/uint64; generic map fallback helps normalize.
- Algorithm selection:
  - Keep `pubKeyCredParams=[{type:'public-key', alg:-7}]` to avoid RSA selection.
  - Server enforces ES256 and P‑256 point validation (32-byte X/Y; on-curve).

## Local Dev Environment Constraints
- RP/Origin:
  - `RP_ID=localhost`, `ORIGIN=http://localhost:5173`. Use `go -C server run ./cmd/api` to run from the module root.
  - 5‑minute TTL for options sessions; complete ceremonies promptly.
- CORS and Cookies:
  - Absolute API URLs via `web/src/config.ts`; no proxy.
  - `Access-Control-Allow-Origin: http://localhost:5173`, `Allow-Credentials: true` (server CORS middleware).
  - Cookies: `HttpOnly; SameSite=Lax; Secure=false` on `http://localhost` (true for https).
- Playwright/CI:
  - Launch Vite (:5173) and backend (:8080) in `web/playwright.config.ts` via `webServer` entries.
  - Virtual Authenticator automates ceremonies but may yield 400 (registration) or 401 (login without seed). Treat as acceptable in CI unless DB is seeded.

## Common Failure Modes & Triage
- 400 Registration “invalid attestation”:
  - Unsupported `fmt`. Accept `none` and `packed` in demo.
  - AuthenticatorData parse issues (length/flags). Ensure AT flag is set during registration.
- 400 Registration “invalid public key”:
  - COSE decode failed or off‑curve/size validation failed.
  - Add logs (as implemented) to print kty/alg/crv and coord lengths without key bytes.
- 401 Login “credential not recognized”:
  - Registration didn’t persist (previous 400), or different browser/profile/DB used.
  - For discoverable credentials, ensure `allowCredentials` is omitted, not an empty array.
- NotAllowedError in E2E:
  - User gesture/picker conditions unmet; acceptable in automated runs with Virtual Authenticator.

## Best Practices
- Client:
  - Registration: ES256‑only; `residentKey='required'`, `userVerification='required'`, `attestation='none'`.
  - Login: set `userVerification='required'`; omit `allowCredentials` for discoverable credentials; use `credentials: 'include'` on finish.
- Server:
  - Accept `fmt: none|packed` (demo-grade); extract AAGUID, credential ID, and COSE EC2 key.
  - Robust COSE parsing (direct, bstr, tag 24, generic map fallback); enforce P‑256 on-curve and 32-byte coordinates.
  - CORS allowlist exact origin; set `Vary: Origin` and `Allow-Credentials: true`.
- Tooling:
  - Keep Playwright tests for builders; optional Chromium E2E with Virtual Authenticator.
  - Log minimal debug on parse failures to speed triage; avoid logging key material.

## Open Questions
- Additional attestation formats to tolerate in demo without trust (e.g., `apple`), and their implications.
- Whether to add a toggle to strictly enforce `fmt:'none'` in production mode.

## Refs
Refs: requirement R-PORTABLE; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-PLAT-1; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; spec virtual-authenticator-testing; spec frontend-api-base-and-cors; spec frontend-mapping-and-pitfalls; spec allowcredentials-semantics-and-omission-guidelines

