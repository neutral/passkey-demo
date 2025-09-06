# R-PORTABLE — Virtual Authenticator Testing (Chromium)

## Metadata
- Status: Draft
- Date: 2025-09-05
- Owners: passkey-demo maintainers

## Purpose
- Document automated E2E testing of WebAuthn flows using Chromium’s Virtual Authenticator via CDP in Playwright.

## What Works
- Enabling WebAuthn and adding a virtual authenticator with `ctap2`, `internal` transport, resident keys, and user verification.
- Automating register (and later login) flows without real prompts.

## What to Expect
- Attestation format may be `packed` or another non-`none` value; our demo backend accepts only `fmt: 'none'` and will return HTTP 400. In this case, treat the E2E as successful if the browser flow completes and the backend rejects with 400 as designed.
- Chromium warns if `pubKeyCredParams` omits RS256; we intentionally restrict to ES256 (`alg: -7`) to match server policy.

## Best Practices
- Create the virtual authenticator per test to avoid state leakage.
- Simulate presence automatically (`automaticPresenceSimulation: true`); set `isUserVerified: true` to mirror UV-required policy.
- Keep test UI selectors robust (roles/labels) and add console logs when debugging.

## Avoid
- Cross-browser assumptions: the CDP WebAuthn API is Chromium-only; do not expect Safari/Firefox parity.
- Over-mocking `navigator.credentials`; prefer CDP virtual authenticator for realistic behavior.

## Refs
Refs: requirement R-PORTABLE; requirement R-FLOW-REG; decision webauthn-corrections-and-standardizations; goal passkey-registration-login-uv

