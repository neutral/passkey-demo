# R-FLOW-LOGIN — ClientDataJSON (CDJ) Usage Explainer (Non‑Developer)

## Purpose & Audience
- Explain what ClientDataJSON is and which fields we parse during login/signing, in practical terms for technical collaborators.

## What ClientDataJSON Is
- A small JSON object built by the browser for WebAuthn ceremonies that includes:
  - `type`: the ceremony (`webauthn.get` for login/assertion, `webauthn.create` for registration/attestation).
  - `challenge`: a server‑provided random value (base64url string).
  - `origin`: the page origin that initiated the ceremony (scheme + host + port).

## Why We Use It Here
- Security and intent binding:
  - `type` confirms the ceremony matches what the server expects.
  - `challenge` binds the assertion to a server‑issued nonce (prevents replay).
  - `origin` is checked separately to ensure the ceremony happened on our approved site.

## How It Works in This App
- The browser fetches options from the server and calls WebAuthn; the browser then returns ClientDataJSON.
- The backend parses ClientDataJSON and decodes the base64url `challenge` back into bytes.
- We use the decoded challenge (and authenticatorData + signature) to verify the assertion.

## Performance & Safety Settings
- Lightweight JSON parse and base64url decode; negligible overhead.
- Tolerant base64url decode accepts padded/unpadded encoding.

## Security & Privacy
- The `challenge` should be unpredictable and specific to a short‑lived server session.
- We do not log raw challenges; logs include only non‑sensitive identifiers.
- Origin checking is performed as a separate policy enforcement step.

## Operating It Day‑to‑Day
- No operator changes needed. If verification fails, the server returns a clear reason (e.g., malformed JSON or bad base64).

## Limitations & When to Upgrade
- We accept only documented `type` values; future ceremonies would require explicit support.
- If additional fields are necessary (e.g., token binding), the parser can be extended with tests.

## Errors & Observability
- Malformed/missing fields or bad base64 produce “bad request” style errors.
- The decoded `challenge` must match the server’s stored value for the session.

## Glossary
- ClientDataJSON: Browser‑produced JSON describing the WebAuthn ceremony.
- challenge: Server‑issued nonce (base64url string) that binds the assertion to an options session.

## Refs
- Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
