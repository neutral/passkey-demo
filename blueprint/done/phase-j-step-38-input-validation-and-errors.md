# Step 38 — Input validation & errors (Done: 2025-09-12)

## Verification Notes
- Health endpoint returns 200 with `X-Request-Id` header.
- `POST /tx/signing/options` without session → 401 envelope `{code: ERR_UNAUTHORIZED}`.
- `POST /authn/passkey/login/finish` with empty body → 401 envelope `{code: ERR_UNAUTHORIZED}`.
- Oversize body to `POST /tx/signing/finish` (~1.2 MiB) → 413 envelope `{code: ERR_PAYLOAD_TOO_LARGE}`.
- Short test suite runs fast with `-short`; db/crypto/E2E tests gated accordingly.

## Scope
- Tighten request- and content-level validation and make error responses consistent across handlers.
- Add per-bundle limits (message length ≤ 1024, nonce ≤ 2^53-1) and enforce Origin/RP allowlists where applicable.
- Standardize all error responses to the JSON envelope with stable codes; include `correlation_id` when present.

## Source to add/modify
- Modify `server/internal/tx/bundle.go`:
  - Enforce `bundle.Message` length ≤ 1024 bytes (runes trimmed to bytes; reject >1024).
  - Enforce `bundle.Nonce` ≤ 9_007_199_254_740_991 (`2^53-1`); reject out-of-range.
  - Add sentinel errors `ErrMessageTooLong`, `ErrNonceOutOfRange` that map to 400.
- Modify `server/internal/tx/options.go`:
  - Ensure new sentinels map to `400 ERR_BAD_REQUEST` via `errx.WriteReq`.
  - Keep existing mappings for `ErrNonceNotMonotonic` (409), `ErrNoCredentials` (409), and internal errors (500).
- Modify `server/internal/tx/finish.go`:
  - Replace `http.Error` with `errx.WriteReq` everywhere; include codes: `ERR_METHOD_NOT_ALLOWED`, `ERR_UNAUTHORIZED`, `ERR_BAD_REQUEST`, `ERR_CONFLICT`, `ERR_FORBIDDEN`, `ERR_INTERNAL`.
  - Use `webauthn.MapPolicyError` and `webauthn.MapVerifyError` to choose 400/401/403 with envelope codes aligned to status.
- Modify `server/internal/webauthn/login_finish.go` and `server/internal/webauthn/reg_finish.go`:
  - Replace `http.Error` with `errx.WriteReq`; map policy failures to 403, bad inputs to 400, session issues to 401, uniqueness conflicts to 409.
  - Keep single-use session deletion behavior.
- Modify `server/internal/http/bodylimit.go`:
  - Return JSON envelope on 413 using `errx.WriteReq`; define or reuse code (see below).
- Modify `server/internal/http/rate.go`:
  - Return envelope on 429 via `errx.WriteReq` with `ERR_RATE_LIMIT`.
- Modify `server/internal/httpx/errors/envelope.go`:
  - Add `CodeTooLarge = "ERR_PAYLOAD_TOO_LARGE"` constant for 413 responses.
- Frontend `web/src/pages/Dashboard.tsx`:
  - Use `postJson` (Api client) for `/tx/signing/finish` to surface `code` and `correlation_id` consistently in error toasts.

## Description files (to create AND updates for modified sources)
- Update `server/internal/tx/bundle.go.desc.md`: add invariants for message-length and nonce range; mapping to 400.
- Update `server/internal/tx/options.go.desc.md`: note envelope codes and mapping table; list new error sentinels.
- Update `server/internal/tx/finish.go.desc.md`: describe envelope adoption and policy/verify error mapping; list HTTP codes.
- Update `server/internal/webauthn/login_finish.go.desc.md` and `server/internal/webauthn/reg_finish.go.desc.md`: document envelope, codes, and single-use session policy.
- Update `server/internal/http/bodylimit.go.desc.md`: note envelope 413 response and code `ERR_PAYLOAD_TOO_LARGE`.
- Update `server/internal/http/rate.go.desc.md`: note envelope 429 `ERR_RATE_LIMIT`.
- Update `server/internal/httpx/errors/envelope.go.desc.md`: include new constant and example responses including `correlation_id`.
- Update `web/src/pages/Dashboard.tsx.desc.md`: document error handling path now showing `{code}`; mention ApiError usage.

## Request/response shape
- Error envelope (all non-2xx): `{ "code": string, "error": string, "correlation_id"?: string }`.
- 413 example: `{ "code": "ERR_PAYLOAD_TOO_LARGE", "error": "payload too large", "correlation_id": "..." }`.
- 429 example: `{ "code": "ERR_RATE_LIMIT", "error": "too many requests", "correlation_id": "..." }`.

## Algorithm
- Bundle validation (server/internal/tx/bundle.go):
  - Decode base64url → CBOR; fail `ErrBundleBase64` on decode errors.
  - Decode canonical CBOR → `types.Bundle`; if fallback path is used, coerce numeric fields as today.
  - Reject `Nonce == 0` (already enforced) and add max constraint `Nonce <= 2^53-1`; else `ErrNonceOutOfRange`.
  - Trim-check `Message` as UTF-8 string; compute byte length; if >1024 bytes → `ErrMessageTooLong`.
  - Re-encode canonical CBOR to `B`; perform account binding and nonce monotonicity as today.
  - Derive anchors and return.
- Handler mappings:
  - Use `errx.WriteReq` uniformly; prefer stable codes by status:
    - 400: `ERR_BAD_REQUEST` (decode/shape issues, `ErrBundleCBOR`, `ErrBundleBase64`, `ErrMessageTooLong`, `ErrNonceOutOfRange`).
    - 401: `ERR_UNAUTHORIZED` (missing/expired session, challenge mismatch, unknown credential, allowlist miss treated as unauthorized).
    - 403: `ERR_FORBIDDEN` (origin/RP policy failures from `MapPolicyError`, UV required failures).
    - 409: `ERR_CONFLICT` (nonce not increasing; signCount not increasing; unique constraint collisions).
    - 413: `ERR_PAYLOAD_TOO_LARGE` (body limit middleware).
    - 429: `ERR_RATE_LIMIT` (rate limit middleware).
    - 5xx: `ERR_INTERNAL`.
  - Include `correlation_id` using request-id middleware.

## Database interactions
- None new. Existing reads/writes remain unchanged; only validation/mapping occurs earlier in the request.

## Policies & limits
- Content limits: `message` ≤ 1024 bytes; `nonce` ≤ `2^53 - 1` (JS safe integer); `nonce > 0` already enforced.
- Transport limits: keep router group body limit at 1 MiB (per ADR router-builder-wiring); bodylimit middleware returns 413 envelope.
- Security policies: Origin and RP ID checks remain via `webauthn.CheckOrigin` and `CheckRpIdHashAllowed`; UV required.

## Sequencing
- Do not change router layering; only replace error writes in handlers/middleware.
- Adopt envelope in `/tx/*` first (finish), then `/authn/*` finish handlers; options handlers already covered (`/tx/signing/options`).
- Frontend change to display `code` for `/tx/signing/finish` can land with backend envelope adoption.

## Tests (happy path required, negative cases, invariants)
- Bundle limits: add `server/internal/tx/bundle_limits_test.go`:
  - Happy: message exactly 1024 bytes and nonce = `2^53-1` → success.
  - Negative: message 1025 bytes → `ErrMessageTooLong`.
  - Negative: nonce 0 → existing `ErrBundleCBOR` path.
  - Negative: nonce `2^53` → `ErrNonceOutOfRange`.
- Finish handlers envelope mapping:
  - `server/internal/tx/finish_handler_envelope_test.go`:
    - Happy: valid flow → 200 JSON body with `tx_id_hex`.
    - 401: missing auth session or challenge mismatch → `{code:"ERR_UNAUTHORIZED"}`.
    - 403: origin or rpId policy failure → `{code:"ERR_FORBIDDEN"}`.
    - 409: signCount not increasing → `{code:"ERR_CONFLICT"}`.
    - 400: malformed JSON/base64 → `{code:"ERR_BAD_REQUEST"}`.
  - `server/internal/webauthn/login_finish_handler_envelope_test.go` and `reg_finish_handler_envelope_test.go` with analogous cases (include 409 on duplicate credential for registration).
- Middleware envelopes:
  - `server/internal/http/bodylimit_test.go`: assert 413 returns JSON envelope with `code: ERR_PAYLOAD_TOO_LARGE`.
  - `server/internal/http/rate_test.go`: assert 429 returns envelope with `code: ERR_RATE_LIMIT`.
- Commands:
  - `cd server && go test ./internal/tx -run TestValidate.* -v`
  - `cd server && go test ./internal/tx -run TestTx.*Envelope -v`
  - `cd server && go test ./internal/webauthn -run Test(Login|Reg).*Envelope -v`
  - `cd server && go test ./internal/http -run Test(BodyLimit|Rate).* -v && go test ./...`

## Verification
- Unit: all new and existing tests pass; bundle limit tests validate boundaries precisely.
- Manual:
  - Start server; POST `/tx/signing/options` with valid bundle → 200.
  - POST `/tx/signing/options` with `message` length 1025 → 400 `{code:ERR_BAD_REQUEST}`.
  - POST `/tx/signing/finish` with tampered `clientDataJSON.challenge` → 401 `{code:ERR_UNAUTHORIZED}`.
  - POST oversized body to `/tx/signing/finish` (e.g., > 1 MiB) → 413 `{code:ERR_PAYLOAD_TOO_LARGE}`.
  - Trigger rate limit bursts on `/authn/*` → 429 `{code:ERR_RATE_LIMIT}`.
- User verification commands
```bash
# Backend tests
cd server
go test ./internal/tx -v
go test ./internal/webauthn -v
go test ./internal/http -v
go test ./... -v

# Run server (in a separate terminal)
go build ./cmd/api && ./cmd/api

# 1) Valid options
curl -sS -X POST :8080/tx/signing/options \
  -H 'Content-Type: application/json' \
  --cookie "sid=<valid>" \
  -d '{"bundle_cbor_b64":"<valid-b64>"}' | jq .

# 2) Message too long → 400 with envelope
node -e 'console.log("{\"bundle_cbor_b64\":\""+"A".repeat(2000)+"\"}")' | \
  curl -sS -X POST :8080/tx/signing/options -H 'Content-Type: application/json' --cookie "sid=<valid>" -d @- | jq .

# 3) Oversized body → 413 with envelope
head -c $((2**20 + 1)) </dev/zero | base64 | \
  sed -e 's/.*/{"bundle_cbor_b64":"&"}/' | \
  curl -sS -X POST :8080/tx/signing/finish -H 'Content-Type: application/json' --cookie "sid=<valid>" -d @- | jq .
```

## Acceptance criteria
- Message length >1024 is rejected with 400 envelope; ≤1024 accepted.
- Nonce outside [1, 2^53-1] rejected with 400 envelope.
- Policy failures (origin/rpId, UV) return 403 envelope; verify/credential/session issues map to 400/401 as specified; nonce/signCount conflicts map to 409.
- Middleware returns envelopes for 413 and 429.
- Frontend surfaces `code` in error toasts for finish errors.

## Notes
- Router group body limit remains 1 MiB as per ADR; per-bundle message limit provides strong guardrails without breaking WebAuthn payloads.
- Keep error messages concise; prefer stable codes for client UX; include `correlation_id` for triage.

## Refs
Refs: requirement R-ERR; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-PLAT-2; decision http-error-envelope; decision router-builder-wiring; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails; decision request-id-and-slog-json

