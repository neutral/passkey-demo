Title: Accept High‑S ECDSA Signatures for Login Only
Status: Superseded by "Accept High‑S ECDSA Signatures for Transaction Signing"
Date: 2025-09-10

Context:
- Our verifier enforces strict ES256 semantics: P‑256 curve, strict DER, and low‑S signatures. This is desirable for robust replay/forgery resistance.
- Some platform authenticators (e.g., Brave/iCloud Keychain) return assertions with high‑S signatures. With a strict low‑S requirement, otherwise valid logins fail with 401 (ErrHighS), degrading UX.
- Transaction signing in this demo is content‑binding and benefits from the stronger, strict low‑S requirement; we do not want to weaken that path.

Decision:
1) For login assertions only, if verification fails solely due to high‑S, normalize S (S' = N − S) and verify again; accept the result if it verifies.
2) Keep strict low‑S enforcement for transaction signing (no normalization on /tx/signing/finish).
3) Retain all other requirements: P‑256 curve, strict DER (no trailing bytes), UV required, rpIdHash and origin policy checks.
4) Emit structured logs indicating the error kind for observability (e.g., ErrHighS vs ErrBadSignature) without logging raw materials.

Consequences:
- Improves compatibility with platform authenticators that emit high‑S signatures for login while preserving stronger guarantees on transaction signing.
- Slightly reduces the uniformity of signature policy across endpoints; documentation and tests have been updated to clarify the split behavior.
- Telemetry can now distinguish high‑S occurrences to inform future policy adjustments.

Alternatives:
- Reject high‑S everywhere (status quo ante). Rejected due to poor interoperability and user experience.
- Accept high‑S everywhere (login and signing). Rejected to preserve stronger guarantees for content‑binding signatures.
- Gate via config flag. Deferred; we may add a toggle if needed, but default policy is as stated.

References:
- Refs: goal passkey-registration-login-uv; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
