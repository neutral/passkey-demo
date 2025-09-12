Title: Accept High‑S ECDSA Signatures for Transaction Signing
Status: Accepted
Date: 2025-09-11

Context:
- Some platform authenticators produce high‑S ECDSA signatures. We already normalize and accept high‑S for login to improve interoperability.
- Transaction signing used strict low‑S verification only, which caused intermittent 401 "verification failed" responses (roughly 50% of attempts), depending on S.

Decision:
1) For transaction signing assertions, verify using the strict verifier first.
2) If and only if the failure is due to high‑S (`ErrHighS`), normalize S (S' = N − S) and verify again; accept if it verifies.
3) Keep all other requirements: P‑256 curve, strict DER, UV required, rpIdHash and origin policy checks, allowlist, and signCount policy.

Consequences:
- Eliminates intermittent verification failures for affected authenticators, improving UX.
- Aligns signing behavior with login for signature normalization without weakening other checks.

Alternatives:
- Keep strict low‑S only: leads to poor UX with intermittent failures.
- Accept high‑S everywhere without strict-verifier-first: loses observability; we prefer explicit fallback on ErrHighS.

References:
- Implementation: `server/internal/webauthn/sig.go::{VerifyAssertion, VerifyAssertionAllowHighS}`; calls added in `server/internal/tx/finish.go`.

Refs: goal transaction-content-signing; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails; decision webauthn-accept-high-s-login-only

