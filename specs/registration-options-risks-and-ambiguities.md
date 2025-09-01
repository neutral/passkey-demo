# Registration Options — Risks and Ambiguities

Potential Pitfalls
- Weak randomness: using non-crypto RNG for challenge/session IDs reduces security; must use `crypto/rand`.
- Predictable session IDs: short or sequential IDs enable guessing; require ≥128 bits entropy.
- Logging secrets: avoid printing raw `challenge`/`reg_session_id` in logs; hash if needed.
- TTL drift: clock skew or test flakiness around `expires_at`; use tolerance windows in tests.
- Memory growth: no capacity limit may allow abuse; consider a sane cap and cleanup.
- Multi-issue collisions: extremely unlikely, but handle detected ID collision by regenerating.
- Policy drift: ensure `rp_id` and `origin` mirror Step 14 policy exactly.

Decisions (current plan)
- Use 32-byte challenge; 16–24 bytes session ID (≥128 bits) base64url.
- In-memory store for demo; optional capacity limit; single-use to be enforced in finish step.
- No attestation trust chain (attestation: none).
- UV required and residentKey required; frontend will reflect these in `create()` parameters.

Operational Notes
- For multi-env setups (staging/dev), ensure each environment uses distinct `reg_session_id` namespaces and separate origin/RP config.
- Consider rate limiting issuance to deter abuse (later hardening step).

Open Questions
- Should session ID be 24 bytes for extra margin? (Leaning yes, but 16+ is adequate.)
- Should we persist sessions to SQLite for restart tolerance? (Out of scope for demo; could add later.)

