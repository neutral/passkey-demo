# R-SEC-UV — UV vs UP in WebAuthn (Explainer)

## Purpose
- Clarify the difference between User Presence (UP) and User Verification (UV), how they surface in WebAuthn, what the RP can require via browser options, and what the server must validate. Provide quick, actionable guidance for this demo’s UV-required policy.

## Quick Summary
- UP = User Presence: proves “someone touched/confirmed”.
- UV = User Verification: proves “the right user unlocked/verified”.

## Definitions
- User Presence (UP): Minimal signal the user is physically present (e.g., button touch, dialog confirm). No identity check.
- User Verification (UV): Local authenticator verification of the user (biometric, PIN, or device screen‑lock). Confirms who is using it.

## Security Impact
- Assurance: UP is low; UV is high.
- Use cases: UP pairs well as a 2nd factor with passwords/OTP. UV enables true passwordless and strong phishing resistance with passkeys.
- Threats: With UP only, anyone with the key/device can authenticate if the server accepts UP. With UV required, stolen devices are blocked unless unlocked.

## Where It Appears (Flags)
- WebAuthn `authenticatorData.flags` bits:
  - `up` (bit 0, 0x01): user presence.
  - `uv` (bit 2, 0x04): user verification.
- Modern platform authenticators commonly set both during passkey use when the device is unlocked.

## RP Controls (Browser API)
- `navigator.credentials.create/get` options:
  - `userVerification`: `"required" | "preferred" | "discouraged"`
    - `required`: authenticator must set `uv=1` or fail.
    - `preferred`: try UV; fall back to UP if not possible.
    - `discouraged`: avoid UV; expect only UP.
  - Related for passkeys: `authenticatorSelection.residentKey` (`"required"|"preferred"|"discouraged"`) and `authenticatorAttachment` (`"platform"|"cross-platform"`).

### Example Policies
- Passwordless (this demo):
  - Registration: `userVerification: "required"`, `residentKey: "required"`, `attestation: "none"`.
  - Authentication: `userVerification: "required"` (empty `allowCredentials` to enable account discovery with discoverable creds).
- 2FA security key with passwords:
  - Registration/Auth: `userVerification: "discouraged"` (no PIN/biometric prompts), typically non‑discoverable credentials.

## Server-Side Checks (Required)
- Always verify `up == 1` on attestation and assertion.
- If policy requires UV (this project), also verify `uv == 1`; reject otherwise.
- With `preferred`, accept when `uv==1`; optionally step‑up if absent.

```text
if !flags.up => reject
if policy.requiresUV && !flags.uv => reject
```

## UX Differences
- UP: quick tap/click confirmation. No unlock required.
- UV: device unlock, biometric, or PIN prompt required.

## Common Pitfalls
- Requiring `userVerification: "required"` with a cross‑platform key that has no PIN/biometric set will fail. Instruct users to set a PIN or use `"preferred"` during migration.
- Expecting both bits on all devices: Some cross‑platform keys may only provide UP unless configured for UV.
- Relying on browser UI alone: Enforce UP/UV by validating flags server‑side.

## Cheat Sheet
- Want passwordless: Require UV, use discoverable credentials (residentKey required).
- Want 2FA with a key: Discourage UV; rely on UP + your primary factor.
- Always check `up`; check `uv` per your policy.

## Refs
- Refs: requirement R-SEC-UV; goal passkey-registration-login-uv; goal webauthn-policy-defaults; decision webauthn-corrections-and-standardizations

