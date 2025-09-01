### Step 13 — Signature verify utility (Done: 2025-09-01)

Scope
- Implement ES256 assertion verification for WebAuthn: verify `SHA256(authenticatorData || SHA256(clientDataJSON))` using P‑256.
- Enforce strict DER (no trailing bytes) and low‑S to prevent malleability.
- Add exported sentinel errors for granular handling.

Artifacts
- Code: `server/internal/webauthn/sig.go` (`VerifyAssertion`, strict DER, low‑S, P‑256 only)
- Errors: `ErrUnsupportedCurve`, `ErrMalformedDER`, `ErrHighS`, `ErrBadSignature`
- Tests: `server/internal/webauthn/sig_test.go` (happy path, high‑S rejection, AD/CDJ tamper, malformed/trailing DER, wrong curve/nil)
- Handler helpers: `server/internal/webauthn/verifyutil.go` (`MapVerifyError`, `HashID`, `LogAssertion`) with tests
- Docs: `server/internal/webauthn/sig.go.desc.md`, `server/internal/webauthn/verifyutil.go.desc.md`
- Audit: `specs/crypto-audit-report.md`

Verification
- `cd server && go test ./...` passes; signature verification suite green.

Notes
- Integration of logging and error mapping into HTTP handlers tracked in Future section (no code changes here).
- RP ID/origin checks and signCount monotonicity validation are planned in subsequent steps.

Refs
- requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-PLAT-2
- decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails
