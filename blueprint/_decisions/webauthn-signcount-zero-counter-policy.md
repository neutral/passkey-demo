Title: WebAuthn signCount==0 Policy (Counter-Not-Supported)
Status: Accepted
Date: 2025-09-10

Context:
- Some platform authenticators (notably iCloud Keychain/Brave) report an authenticator sign counter of 0 for all assertions. This indicates the device does not support, or does not expose, a monotonic counter for clone detection.
- Our original goal and ADR required a strictly increasing `signCount` during login as a replay/clone defense. In practice, enforcing monotonicity when the device reports 0 leads to unnecessary login failures (HTTP 409) for legitimate users.

Decision:
1) Treat `signCount == 0` in authenticatorData as “counter not supported”.
2) When `signCount == 0`, do not enforce monotonicity and do not update the stored counter.
3) When `signCount > 0`, require strictly increasing vs the stored value; reject equal or lower with HTTP 409 and do not update the counter.
4) Maintain UV=required, rpIdHash, origin policy, signature verification, and all other checks unchanged.

Consequences:
- Improves compatibility with platform passkeys that report `signCount == 0`, eliminating false 409 conflicts during login.
- Preserves clone/replay detection where supported by keeping strict monotonic enforcement for non-zero counters.
- Stored counters may remain 0 indefinitely for such authenticators; metrics should distinguish “counter-unsupported” from “monotonic” devices.
- Slightly weakens clone detection for devices without counters; other defenses (UV required, origin/rpId policy, session management) still apply.

Alternatives:
- Continue enforcing monotonicity for all values, causing persistent 409 errors on affected devices (rejected: poor UX, prevents login).
- Always ignore signCount and never store/update it (rejected: loses clone detection where it is available).
- Heuristic bump: treat 0→1 as first-use and then enforce (rejected: speculative and may still fail when devices keep reporting 0).

References:
- Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails

