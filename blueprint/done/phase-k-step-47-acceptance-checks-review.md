# Step 47 — Acceptance checks review (Done: 2025-09-12)

## Scope
- Review the current implementation against the Acceptance Checks in the plan and capture any gaps or clarifications needed.

## Observations
- Build: OK — `make build` succeeds; Go short tests pass.
- Register/Login prompts: Likely OK — Endpoints and UI are wired; server enforces UV and session creation. (E2E prompt display requires manual run.)
- Sign flow: OK — `/tx/signing/options` and `/tx/signing/finish` implemented; dashboard signs and list endpoint returns items.
- TTLs: OK — Registration/login/tx option sessions use 5-minute TTL in their in-memory stores.
- Nonce replay → 409: OK — `ErrNonceNotMonotonic` mapped to 409.
- UV check → 403: OK — UV required in reg/login/tx finish; missing UV returns 403.
- Origin/RP guards: OK — `CheckOrigin`/`CheckRpIdHashAllowed` enforced with IDNA normalization; mapping returns 403 for mismatches, 400 for malformed.
- Low‑S enforcement (mismatch):
  - Acceptance text says “high‑S → 400”, but code intentionally accepts normalized high‑S for compatibility (login and tx signing), per ADR direction.
  - Options:
    1) Update acceptance criteria to reflect compatibility policy (accept high‑S with normalization).
    2) Or change signing path to strictly reject high‑S (400) while keeping login compatible; update specs and code accordingly.
  - Recommendation: Keep current compatibility policy (lower operator friction across authenticators). If stricter policy is desired for tx signing, we can scope a follow‑up step to make signing strict and keep login compatible.

## Acceptance checks (as tracked)
- Build: `go build ./server/...` and `npm run build` in `/web` both succeed.
- Register/Login: Using macOS with Touch ID, both ceremonies prompt for fingerprint; login sets cookie.
- Sign: Create 2 messages with nonces 1 and 2; both appear in `/tx/list` and DB.
- Ephemeral TTLs: Registration/login/tx option sessions expire after 5 minutes; expired attempts return 409.
- Replay/Nonce: Re-submit nonce 2 → 409 conflict.
- UV check: If browser returns an assertion without UV (simulate by forcing options incorrectly) → 403 forbidden.
- Origin/RP guard: Change `origin` in request body → 403.
- Low‑S enforcement: CURRENT IMPLEMENTATION accepts high‑S by normalization for compatibility; adjust acceptance note or policy as decided.

## Verification
- Build + short tests:
```bash
cd server && go build ./... && go test -short ./...
```
- Manual run (dev):
```bash
set -a; source .env.example; set +a
make run
```

## Notes
- The only deviation is the low‑S acceptance policy vs. the older acceptance note. Align the acceptance criteria to the implemented policy or file a follow‑up change to enforce strict low‑S for signing.

## Refs
- Refs: requirement R-PLAT-2; decision webauthn-accept-high-s-login-only; decision webauthn-accept-high-s-signing-too; decision structured-logging-with-slog-guidelines

