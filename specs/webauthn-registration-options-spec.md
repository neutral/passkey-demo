# WebAuthn Registration Options — Technical Spec

Date: 2025-09-01
Status: Draft (analyze-first)
Owners: passkey-demo maintainers

Scope
- Server endpoint that initiates WebAuthn registration by returning a fresh challenge and minimal policy options.
- Session semantics that bind the challenge to the relying party and origin with a short TTL.

Endpoint
- Method/Path: `POST /authn/passkey/registration/options`
- Request: empty body; server derives all values from configuration and secure randomness.
- Response (JSON):
  - `reg_session_id` (string): base64url, opaque, ≥128 bits of entropy.
  - `challenge` (string): base64url, 32 bytes of crypto randomness.
  - `options` (object):
    - `rp_id` (string): equals configured `RP_ID`.
    - `origin` (string): equals configured `Origin`.
    - `uv_required` (bool): true.
    - `attestation` (string): "none".
  - `expires_at` (number): UNIX seconds, now + 300s.

Session Model
- Storage: in-memory, concurrency-safe map keyed by `reg_session_id`.
- Value: `{challenge []byte, rp_id string, origin string, expires_at time}`.
- TTL: 5 minutes. Entries are removed on expiry or first successful consumption (single-use).
- Capacity: optionally cap total live sessions (e.g., 10k) to prevent unbounded memory growth.

Generation Rules
- `reg_session_id`: random bytes (≥16 bytes), base64url (unpadded). Must be unique across live sessions; collision → retry.
- `challenge`: exactly 32 random bytes, base64url (unpadded). Must be unique across a short window (best-effort randomness suffices).
- Randomness: `crypto/rand` only.

Policy Defaults
- `userVerification: required` — mandates biometric/PIN on authenticators that support UV.
- `residentKey: required` — discoverable credential for username-less flows.
- `attestation: none` — demo avoids trust chain processing.

Normalization and Echo
- Echo server config values for `rp_id` and `origin` exactly as enforced by Step 14 checks.
- Do not derive or modify RP ID based on incoming request.

Error Cases
- 500 on internal errors (randomness failure, store put failure).
- 429 if capacity limit exceeded (optional defensive limit).

Security Considerations
- The challenge is secret-ish but not sensitive at rest; avoid logging raw values. If logging, hash with SHA-256.
- Ensure single-use: the same `reg_session_id` cannot be used twice in finish step.
- Use monotonic `expires_at` and reject finish after TTL.

Interoperability
- Use base64url without padding to align with browser expectations and our existing encoding utils.
- Frontend will transform minimal `options` into `PublicKeyCredentialCreationOptions`.

Open Questions
- Session persistence: remain in-memory for demo; optionally support SQLite-backed sessions later.
- Capacity default: propose 10k live sessions; tune as needed.

References
- WebAuthn Level 2: Creation options and challenge handling.
- Internal specs: Step 14 (RP/Origin verification), encoding/base64url utilities.

