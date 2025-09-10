# Testing Setup Notes

- Date: 2025-09-10
- Scope: End-to-end testing approach for the repo, split into Unit Tests and Integration Tests.
- Goal: Make correctness, security, and UX guardrails repeatable and observable via fast, deterministic tests and pragmatic UI specs.

## Contents

- [Unit Tests](#unit-tests)
  - [Purpose & Strategy](#purpose--strategy)
  - [Layers & Coverage](#layers--coverage)
  - [Error Mapping Matrix (Handlers)](#error-mapping-matrix-handlers)
  - [Deterministic Test Data & Helpers](#deterministic-test-data--helpers)
  - [Fuzz Testing](#fuzz-testing)
  - [Commands & Troubleshooting (Unit)](#commands--troubleshooting-unit)
- [Integration Tests (Playwright)](#integration-tests-playwright)
  - [Purpose & Strategy](#purpose--strategy-1)
  - [Testbed](#testbed)
  - [Install & Run](#install--run)
  - [Suite Summary (selected specs)](#suite-summary-selected-specs)
  - [Stubs, Mocks, and Stability](#stubs-mocks-and-stability)
  - [UI Error Handling Conventions](#ui-error-handling-conventions-kept-for-test-stability)
  - [Troubleshooting (Integration)](#troubleshooting-integration)
  - [Current Status](#current-status)
- [Appendices](#appendices)
  - [Testing Philosophy](#testing-philosophy-appendix)
  - [Runbook: Investigating Failures](#runbook-investigating-failures-appendix)
  - [Local Dev Tips](#local-dev-tips-appendix)
  - [Current Status Snapshot](#current-status-snapshot-appendix)
  - [Next Improvements](#next-improvements-appendix)

## Unit Tests

### Purpose & Strategy
- Validate core logic where it’s cheapest and fastest to catch defects.
- Build a layered pyramid: Encoding/Crypto → WebAuthn Parse/Verify/Policy → HTTP Middleware → Transaction Validation/Handlers → Storage/Config.
- Prefer deterministic inputs and assertion of invariants over brittle snapshots.
- Inject time (`now func() time.Time`) where expiry matters; use in-memory SQLite for isolation.

### Layers & Coverage
- Encoding/Crypto
  - Purpose: canonical CBOR determinism, COSE→ECDSA validity.
  - Invariants:
    - EncodeCanonical(DecodeCanonical(B)) == B (byte equality).
    - COSE public keys are on curve; invalid/off-curve rejected.
  - Files: `internal/encoding/cbor_test.go`, `internal/crypto/crypto_cose_test.go` (+ fuzzers below).

- WebAuthn Parse/Verify/Policy
  - Purpose: robust CDJ/AD parsing, low‑S enforcement, strict DER decoding, RP/Origin/UV policy.
  - Invariants:
    - CDJ type ∈ {`webauthn.get`, `webauthn.create`}.
    - Base64url (padded/unpadded) challenges accepted; invalid base64 rejected.
    - AD length ≥ 37; signCount parsed big‑endian; flags helpers consistent.
    - VerifyAssertion rejects high‑S signatures, tampered AD/CDJ, malformed/trailing‑bytes DER, wrong curve.
    - UV bit required for finish flow.
  - Files: `internal/webauthn/{cdj_test.go, ad_test.go, sig_test.go, policy_test.go}` (+ fuzzers below).

- HTTP Middleware (Session)
  - Purpose: resolve session from cookie, optionally refresh expiry, attach session context.
  - Invariants:
    - With `refresh=true` and less than half‑TTL remaining, `expires_at` extended by TTL.
    - Missing/expired session passes through (downstream protected handlers return 401).
  - Files: `internal/http/session_test.go`, `internal/http/session_refresh_test.go`.

- Transaction Validation/Handlers
  - Purpose: canonical bundle validation (B), anchor derivation, options/finish semantics (allowlist, signCount, challenge), storage persistence, and status mapping.
  - Invariants:
    - Account binding: CBOR(sender_key) == `acct_cbor` (byte‑for‑byte).
    - Nonce strictly increasing per account; equal or lower → 409.
    - Anchors: `challenge = SHA256("CHALv1" || B)`, `tx_id = SHA256("TXIDv1" || B)`; `challenge != tx_id`; B change flips anchors.
    - Allowlist enforced: `rawId` ∈ options allow_credentials; else 401.
    - signCount strictly increases; equal/lower → 409.
    - Tx session is single‑use: success deletes `tx_session_id`.
  - Files: `internal/tx/{bundle_test.go, options_test.go, finish_test.go, list_test.go, anchors_test.go, finish_handler_negative_test.go, phase_f_e2e_test.go}`.

- Storage/Config
  - Purpose: DB PRAGMAs and schema; config defaults, normalization, allowlists.
  - Files: `internal/storage/storage_test.go`, `internal/config/config_test.go`.

### Error Mapping Matrix (Handlers)
- `/tx/signing/options`
  - 401: missing/expired auth session; sender key mismatch.
  - 400: invalid bundle base64/CBOR.
  - 409: nonce not monotonic; no credentials for account.
  - 200: returns `tx_session_id`, `challenge`, `options.allow_credentials`, `tx_id_hex`, `expires_at`.
- `/tx/signing/finish`
  - 401: `ErrAuthSession`, `ErrTxSession`, `ErrCredUnknown`, `ErrCredMismatch`, `ErrAllowlist`, `ErrChallenge`.
  - 400: `ErrBadJSON`, `ErrBadBase64`, `ErrTypeMismatch`, malformed AD/CDJ.
  - 409: `ErrSignCount` (≤ stored).
  - 403: policy errors (Origin/RP allowlists); verification policy (e.g., UV missing → forbidden).
  - 200: `stored: true`, `tx_id_hex`, deletes tx session.

### Deterministic Test Data & Helpers
- In‑memory SQLite: `file::memory:?_busy_timeout=5000&_foreign_keys=on` and `storage.Migrate(db)` in each test.
- Deterministic COSE EC2: pad `Gx`, `Gy` (P‑256 base point) to 32 bytes; stable across runs.
- Canonical CBOR: `internal/encoding` `EncodeCanonical` / `DecodeCanonical` only.
- Builders:
  - AD: 37‑byte header (rpIdHash, flags, signCount) plus attested data as needed; `UP=0x01`, `UV=0x04`.
  - CDJ: `{ type, challenge(b64url), origin }` JSON.
  - Low‑S DER: adjust S to be ≤ N/2 to pass verify.
- Time injection: pass `now func() time.Time` to helpers that enforce TTL/expiry.

### Fuzz Testing
- Goals: prove panic‑freeness and robustness across decoders/converters.
- Targets & seeds:
  - `FuzzParseClientDataJSON` (seed: minimal valid JSON with padded/unpadded challenge).
  - `FuzzParseAuthData` (seed: 37‑byte header and a short invalid case).
  - `FuzzDecodeCanonical` (seed: small CBOR map).
  - `FuzzToECDSA` (seed: zeroed 32‑byte X/Y buffers).
- Commands (time‑boxed):
```
# From repo root
# Single fuzz function per package (10s each)
go -C server test -run=^$ -fuzz=FuzzParseClientDataJSON -fuzztime=10s ./internal/webauthn
go -C server test -run=^$ -fuzz=FuzzParseAuthData      -fuzztime=10s ./internal/webauthn
go -C server test -run=^$ -fuzz=FuzzDecodeCanonical     -fuzztime=10s ./internal/encoding
go -C server test -run=^$ -fuzz=FuzzToECDSA             -fuzztime=10s ./internal/crypto

# All fuzzers in a package (keep short for CI)
go -C server test -run=^$ -fuzz=Fuzz -fuzztime=10s ./internal/webauthn
```
- Notes:
  - Requires Go ≥ 1.18 (we use 1.25).
  - Reduce `-fuzztime`, add `-parallel=1` if needed.
  - FUZZERS IGNORE ERRORS: only assert “no panics”; logic correctness covered by unit tests.

### Commands & Troubleshooting (Unit)
```
# All unit tests
go -C server test ./...

# With race detector
go -C server test -race ./...

# Focused package/test
go -C server test ./internal/tx -run Anchors -v
```
- Troubleshooting:
  - unknown flag `-fuzz`: ensure Go ≥ 1.18 and run from `server/` root.
  - Package failures block `./...`: narrow to specific package to isolate.
  - Use `-cover -coverprofile=coverage.out` and `go tool cover -html=coverage.out` for coverage.

---

## Integration Tests (Playwright)

### Purpose & Strategy
- Exercise realistic UI flows in a headful browser engine with mocked server edges where needed.
- Keep specs fast/stable by:
  - Stubbing `navigator.credentials` for create/get.
  - Intercepting network routes to simulate server responses and error codes.
  - Waiting on specific network responses (`waitForResponse`) around key POSTs.
- Reflect backend policies in UI error rendering (toasts + unauthorized prompts).

### Testbed
- Vite dev server (`:5173`) for frontend; Go API server (`:8080`) for backend, both launched by Playwright `webServer` config.
- `baseURL: http://localhost:5173`, project: Chromium desktop, fully parallel.

### Install & Run
```
# From repo root
npm -C web ci
npx -C web playwright install --with-deps

# Run all tests (list reporter)
npm -C web run test:ui -- --reporter=list

# Filter by name
npm -C web run test:ui -- -g "signing"
npm -C web run test:ui -- -g "error toast"
```

### Suite Summary (selected specs)
- `ui.spec.ts`: Home navigation and primary buttons UI sanity.
- `encoding.spec.ts`: Base64url/UTF‑8 helper behavior in browser via dynamic import.
- `login-webauthn.spec.ts`: `toRequestOptions` mapping semantics and `buildLoginFinish` encoding.
- `webauthn.spec.ts`: `toCreationOptions` mapping and `buildRegFinish` encoding.
- `dashboard-list.spec.ts`: Unauthorized prompt (401) rendering.
- `tx-signing.spec.ts`: Mocked happy path options → get → finish; waits for options/finish; asserts bundle sending and finish payload fields; clears preview.
- `tx-signing-negative.spec.ts`: Options (401/409/400) and finish (409/400) error surfacing in UI.
- `error-toasts.spec.ts`: Toast messages for 400/401/413/429; dashboard account_key 401 shows toast + unauthorized prompt.
- `webauthn-e2e.chromium.spec.ts`: Registration with Virtual Authenticator (Chromium CDP).
- `login-e2e.chromium.spec.ts`: E2E login may produce HTTP 401 or NotAllowed; both acceptable in unseeded envs.

### Stubs, Mocks, and Stability
- WebAuthn stubs: `page.addInitScript` to override `navigator.credentials.create/get` returning objects with ArrayBuffer fields.
- Network interception: `page.route` → `route.fulfill` with status/body; capture `route.request().postData()` to assert outgoing payloads (e.g., `bundle_cbor_b64`).
- Stability techniques:
  - Use `Promise.all([ page.waitForResponse(...), click ])` to avoid racing assertions on POSTs.
  - Prefer `getByRole` selectors; fall back to `getByText` for explicit messages.

### UI Error Handling Conventions (kept for test stability)
- ErrorToast renders both a detail line and a back‑compat line `Error: <detail>`.
- Dashboard shows the unauthorized prompt whenever `unauthorized === true`, even if a toast is rendered.
- `/tx/signing/options` 401: only mark unauthorized and show prompt (no toast) to align with negative specs.
- `/tx/signing/options` 413/429: include `HTTP <status>` and any `message`/`code` in the toast content.

### Troubleshooting (Integration)
- Run a single spec with `-g`, enable `trace: 'on-first-retry'`, and view with `npx playwright show-trace`.
- Logs: Playwright prints Go server logs (helpful for HTTP errors).
- Ports: ensure 5173/8080 are free; config uses `reuseExistingServer` for iterative runs.

### Current Status
- Frontend Playwright: all specs passing locally (25/25).

---

## Appendices

### Testing Philosophy (Appendix)
- Prefer invariant‑style assertions and deterministic inputs.
- Keep the majority of logic validated in unit tests; use UI specs to check integration surfaces and UX messaging.
- Treat fuzz tests as panic‑free safety nets; keep them time‑boxed in CI and richer locally.

### Next Improvements (Appendix)
- Add a concurrency replay test for finish (race; exactly one success) and run under `-race`.
- Step 36: golden vectors CLI for fixed bundle → exact hex(B)/challenge/tx_id verification.
- Step 38: standardized error envelopes (`{ code, error }`) and frontend parser/tests alignment.
- Optionally enable Playwright traces persistently for CI artifacts.

---

### Runbook: Investigating Failures (Appendix)

### Integration (Playwright)
- Re-run a single spec: `npm -C web run test:ui -- -g "<title fragment>" --reporter=list`.
- Enable tracing: ensure `trace: 'on-first-retry'` in config, then `npx -C web playwright show-trace <trace.zip>`.
- Inspect network mocks: temporarily `console.log(route.request().postData())` in the route handler.
- Watch logs: Playwright prints Go server stdout/stderr; check status codes and error lines.
- Flake suspects: missing `waitForResponse` for POSTs, unauthorized prompt not visible due to state, port reuse collisions.

### Unit (Go)
- Narrow scope: `go -C server test ./internal/tx -run Finish -v`.
- Add diagnostics: `t.Logf("tx_id=%x challenge=%x", tx, ch)` and re-run with `-v`.
- Coverage view: `go -C server test ./... -cover -coverprofile=coverage.out && go tool cover -html=coverage.out`.
- Race detector for concurrency issues: `go -C server test -race ./internal/tx`.

---

### Local Dev Tips (Appendix)
- Ports: keep `5173/8080` open; config uses `reuseExistingServer` to avoid churn.
- Versions: Node 22+, Go 1.25 recommended; Go ≥ 1.18 required for `-fuzz`.
- Determinism: use the provided deterministic COSE helpers and canonical encoders; avoid time.Now() in tests—inject `now`.
- Test data: keep messages/nonce small and fixed in specs; assert minimal, stable strings.

---

### Current Status Snapshot (Appendix)
- Frontend (Playwright): 25 passed locally; Chromium project.
- Backend (Go): unit tests green; fuzzers panic‑free when time‑boxed (10s); race detector clean on targeted packages.
