# R-PORTABLE — Platform Authenticator Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Ensure compatibility with platform authenticators across modern browsers; use discoverable credentials and UV required.

## Interfaces
- Same WebAuthn flows; adjust UX to prompt authenticator use.

## Data / Models
- N/A beyond normal flow data.

## Algorithms
- Registration with `residentKey: "required"`; UV required everywhere; allow username-less login via discoverable credentials.

## Security / Privacy
- Maintain origin/rpId integrity; no attestation trust chain required (attestation: "none").

## Errors / Observability
- Surface platform-specific errors cleanly in UI; provide retry guidance.

## Testing Strategy
- Manual runs on Safari and Chrome with platform authenticators; ensure consistent behavior.

## Open Questions
- None.

Refs: spec spec-a; spec spec-b; decision webauthn-corrections-and-standardizations; requirement R-PORTABLE
