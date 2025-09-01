# Registration Options — Test Vectors

Purpose: Validate correctness, policy, and boundary conditions for Step 15.

Assumptions
- Config: `RP_ID=example.com`, `Origin=https://example.com`.
- TTL: 300 seconds.

Response Shape
- Must contain keys: `reg_session_id` (string), `challenge` (string), `options` (object), `expires_at` (number).
- `options` contains: `rp_id` (== `example.com`), `origin` (== `https://example.com`), `uv_required=true`, `attestation="none"`.

Entropy & Lengths
- `challenge`: base64url string decodes to exactly 32 bytes.
- `reg_session_id`: base64url decodes to ≥16 bytes (128 bits) of randomness; uniqueness across N=10k issued sessions.

TTL Window
- `expires_at` ∈ [now+290s, now+310s].

Store Semantics
- Inserted session value equals decoded `challenge`, and echoes `rp_id`/`origin` from config.
- Issuing N sessions increases store count by N.
- After TTL, session is considered expired; finish handler (later step) must reject.

Failure Cases
- Randomness failure → 500.
- Capacity exceeded (if limit set) → 429.

Encoding
- All base64url values are unpadded; tolerant decoding on the consumer side should succeed.

