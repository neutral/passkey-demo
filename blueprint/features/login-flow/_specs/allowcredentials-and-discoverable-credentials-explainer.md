# R-FLOW-LOGIN — Discoverable Credentials and `allowCredentials` (Non‑Developer)

## Purpose & Audience
- Explain what “discoverable credentials” (a.k.a. resident credentials/passkeys) are, how the `allowCredentials` option affects browser UX and authenticator behavior during login, and why this app leaves it empty by default.

## What `allowCredentials` Is
- A list of credential IDs the browser should allow during `navigator.credentials.get(...)`.
- Omit the property entirely to enable “account discovery” using client‑side discoverable credentials for the RP ID (no pre‑selected credential from the server).
- If the list contains IDs, browsers filter authenticators to those specific credentials and typically skip the account chooser.
- Important: Do not send an empty array to mean “no filter.” Some implementations interpret an empty list as “no credentials allowed,” leading to immediate failure or “No passkeys available.”

## Discoverable vs Non‑Discoverable Credentials
- Discoverable (resident) credentials: Stored on the authenticator/OS; modern “passkeys”. The authenticator can find these by RP ID alone.
- Non‑discoverable (non‑resident) credentials: The server must provide the credential IDs in `allowCredentials` so the authenticator knows which one to use.

## Our Demo Approach (Step 18)
- We intentionally return `allow_credentials: []` in assertion options to leverage discoverable credentials.
- Benefits:
  - Passkey UX: The OS/native account/passkey chooser appears and handles account discovery.
  - Simpler flow: No username/email pre‑step required to narrow credential IDs.
  - Privacy: We do not enumerate or hint at specific credential IDs from the server side.
- Server behavior (Step 19): The authenticator returns the selected credential ID; the server uses it to identify the account and verify the assertion.

## When To Provide `allowCredentials`
- Security keys or older authenticators that only support non‑discoverable credentials.
- Targeted flows after user identification (e.g., user entered email), to reduce prompts and constrain which credentials can be used.
- Recovery or migration scenarios where a specific credential must be exercised.

## UX & Browser Behavior Notes
- Empty `allowCredentials` prompts the platform’s passkey picker for the configured RP ID.
- If the user has no discoverable credentials for the RP ID, the browser may show “No passkeys available.” In such cases, provide guidance or a username step that enables filtered options.
- `userVerification: "required"` still applies; the authenticator must confirm biometric/PIN before returning an assertion.

## Security & Privacy Considerations
- RP scoping still applies via `rpId` and the `rpIdHash` check during finish.
- Empty `allowCredentials` avoids sending credential IDs that could correlate users across requests; however, the presence/absence of passkeys on a device can still be inferred by the user’s UI.
- The server validates UV, origin, and rpIdHash during finish; this choice does not weaken cryptographic verification.

## Limitations & Future Options
- Environments with non‑resident credentials require `allowCredentials` to be populated to make those credentials usable.
- Later steps may add a username pre‑step or a mode that populates `allowCredentials` from the user’s known credentials when appropriate.

## Refs
- Refs: requirement R-FLOW-LOGIN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
