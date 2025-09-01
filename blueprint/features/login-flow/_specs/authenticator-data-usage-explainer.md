# R-FLOW-LOGIN — Authenticator Data (AD) Usage Explainer (Non‑Developer)

## Purpose & Audience
- Explain what WebAuthn “authenticatorData” is and which fields we read during login/signing, in practical terms for technical collaborators.

## What Authenticator Data Is
- A byte sequence returned by the authenticator (e.g., Touch ID) that includes:
  - `rpIdHash` (32 bytes): a SHA‑256 hash of the relying party ID (our site identifier).
  - `flags` (bitfield): state bits such as “user present” and “user verified”.
  - `signCount` (counter): a usage counter maintained by the authenticator.
- In some cases it also contains “attested credential data” or “extensions” (registration‑focused fields), but those are not used for login checks.

## Why We Use It Here
- Security checks during login/signing:
  - The `rpIdHash` must match our configured site, preventing cross‑origin assertions.
  - The `flags` must indicate “user verified” (biometric/PIN), enforcing our UV policy.
  - The `signCount` must increase over time, helping detect authenticator cloning/downgrades.

## How It Works in This App
- The backend parses the fixed header of authenticatorData (32‑byte `rpIdHash`, 1‑byte `flags`, 4‑byte big‑endian `signCount`).
- The remainder may contain registration‑specific data (ignored for login).
- We combine AD with other inputs (clientDataJSON, signature, and the public key) to verify assertions.

## Performance & Safety Settings
- Fast, fixed‑size parsing for the header used in login; minimal overhead.
- Clear errors on malformed or short inputs; no panics.

## Security & Privacy
- The `rpIdHash` ensures the assertion is scoped to our site.
- UV is required for all ceremonies; assertions without UV are rejected.
- Only counters and policy bits are read and logged in aggregate (no raw byte dumps).

## Operating It Day‑to‑Day
- No configuration needed beyond correct RP ID and origin settings.
- If login errors occur, the server returns a precise error (e.g., wrong RP ID hash) and logs a concise reason.

## Limitations & When to Upgrade
- The header excludes attested credential data and extensions, which are parsed separately during registration.
- Some authenticators do not strictly increase counters; the app defends accordingly but may require policy review for edge devices.

## Errors & Observability
- Short/invalid inputs produce “bad request” style errors; logs include the failure cause.
- Counter and UV policy enforcement are observable in tests and logs.

## Glossary
- rpIdHash: SHA‑256 of the relying party ID (site identifier).
- flags: bitfield indicating user presence (UP), user verification (UV), and optional sections.
- signCount: a usage counter tied to the authenticator; should increase over time.

## Refs
- Refs: requirement R-FLOW-LOGIN; requirement R-SEC-UV; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations
