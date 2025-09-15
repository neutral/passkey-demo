# R-FLOW-REG — Frontend Mapping and Pitfalls

## Metadata
- Status: Draft
- Date: 2025-09-05
- Owners: passkey-demo maintainers

## Purpose
- Explain how server registration options map to `PublicKeyCredentialCreationOptionsJSON` (consumed by `@simplewebauthn/browser`) and capture pitfalls observed during implementation and testing.

## Mapping
- `options.rp_id` → `rp.id`; set `rp.name` (display only).
- `challenge` remains a base64url string in `*OptionsJSON`; the library handles encoding for the browser.
- `authenticatorSelection.residentKey='required'`; `userVerification='required'`.
- `attestation='none'`.
- `pubKeyCredParams=[{ type: 'public-key', alg: -7 }]` (ES256 only to match server).
- `user`: ephemeral (32 random bytes) encoded base64url in `user.id` with `name/displayName` placeholders. Identity is server-side (passkey-first via COSE key).
- Node server also returns `reg_session_id` (24-char base64url) and `expires_at` (epoch seconds); the frontend stores both and posts them back in the finish request alongside `RegistrationResponseJSON`.

## Pitfalls & Gotchas
- Binary conversions must be exact for any remaining helpers (base64url no padding, URL-safe alphabet); prefer centralized helpers. Using the library avoids most manual ArrayBuffer transforms.
- `new URL(path, API_BASE)`: `API_BASE` must be absolute; relative values like `/api` cause failures.
- Chromium warns if RS256 is omitted; we keep ES256-only per policy. Some authenticators may be incompatible with ES256-only — acceptable for demo scope.
- Manual E2E requires secure context; `http://localhost` counts as secure in modern browsers.
- `navigator.credentials.create` requires a user gesture and may reject with `NotAllowedError` if the prompt is dismissed.

## Testing
- Unit: adapter tests that map server option shapes to `PublicKeyCredentialCreationOptionsJSON`.
- Page tests: verify `RegistrationResponseJSON` is posted including `reg_session_id` and base64url fields.
- E2E: Virtual Authenticator in Chromium via CDP; accept HTTP 400 for non-`none` attestation.

## Refs
Refs: requirement R-FLOW-REG; requirement R-PLAT-1; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv; spec spec-a; spec spec-b
