### Step 14 — RP ID & Origin checks (Done: 2025-09-01)

Scope
- Enforce relying party binding by verifying:
  - RP ID hash match: `SHA-256(rpID)` equals `authenticatorData.rpIdHash`.
  - Origin policy: `clientDataJSON.origin` matches configured origin or explicit allowlist.
- Implement normalization and strict policy per specs (no wildcards/eTLD+1, default-port handling, trailing-dot tolerance, dev-only localhost HTTP).

Artifacts
- Code: `server/internal/webauthn/policy.go`
  - `CheckRpIdHash`, `CheckRpIdHashAllowed` (primary + allowlist), `CheckOrigin` (scheme/host/port validation).
  - Sentinel errors: `ErrRpIdInvalid`, `ErrRpIdHashMismatch`, `ErrOriginMalformed`, `ErrOriginScheme`, `ErrOriginHost`, `ErrOriginPort`, `ErrOriginNotAllowed`.
- Docs: `server/internal/webauthn/policy.go.desc.md`
- Tests: `server/internal/webauthn/policy_test.go` (vectors for rpID validity, trailing dot, allowlist, ports/scheme/host, dev localhost, IPv6 literal).
- Integration: `MapPolicyError` added to `server/internal/webauthn/verifyutil.go` with tests; maps policy errors to 400/403 and stable `error_kind` strings.

Verification
- `cd server && go test ./...` passes; policy tests enforce expected pass/fail per `specs/rp-origin-test-vectors.md`.

Notes
- IDNA/punycode normalization is tracked in Future (not implemented here). Config should use ASCII/punycode for i18n domains.
- Dev exception is narrowly scoped to `http://localhost:<port>`.

Refs
- specs/webauthn-rp-origin-verification-spec.md; specs/rp-origin-test-vectors.md; specs/rp-origin-risks-and-ambiguities.md
- requirement R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations

