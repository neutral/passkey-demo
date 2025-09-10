### Step 35 — Go Unit Tests (Done: 2025-09-10)

Verification notes
- Added unit tests for anchors, finish handler error mappings, session refresh, bundle roundtrip; plus fuzzers for CDJ/AD/CBOR/COSE. Unit tests compile locally; fuzz runs are time-boxed (10s) and panic-free.

35. **Go Unit Tests**

    - Purpose
      - Raise confidence in core crypto/encoding, WebAuthn parsing/verification/policy, session/middleware, and transaction flows.
      - Lock in critical invariants: canonical CBOR, anchor derivation, low‑S enforcement, strict signCount increase, account binding, and nonce monotonicity.

    - General Strategy
      - Layered pyramid: Encoding/crypto → WebAuthn parse/verify → HTTP middleware → TX validation/handlers → Storage/config.
      - Determinism: use deterministic or seeded inputs (fixed COSE points, canonical CBOR) and `t.TempDir` SQLite DBs.
      - Time injection: where expiry matters, pass `now func() time.Time` so tests can simulate TTL/expiry.
      - Table‑driven tests: group error mappings, policy checks, and edge cases per handler.
      - Property & fuzz tests: fuzz decoders (CBOR, CDJ), ToECDSA rejection for off‑curve, canonical encoder determinism.
      - Golden vectors: fix a bundle B and assert exact hex for `CHALv1`/`TXIDv1` anchors and canonical encoding output (Step 36 will deepen this).
      - Concurrency: add replay/parallel finish checks and run with `-race`.

    - Existing Tests (inventory)
      - Encoding/crypto: canonical CBOR determinism, roundtrips, malformed decode; COSE→ECDSA valid/invalid; base64url helpers.
      - WebAuthn: AD flags/length/counter; CDJ types/base64 and invalids; VerifyAssertion happy/tamper/high‑S/DER/curve; policy helpers; login/registration options/finish.
      - HTTP: session middleware happy/expired/missing; CORS; rate limiting; body limits.
      - TX: ValidateAndAnchorBundle happy + errors; options handler happy + negatives; finish happy; list; phase‑F E2E.
      - Storage/config: PRAGMAs and tables; config defaults/normalize/allowlists.

    - Coverage Gaps & Additional Tests
      - TX/anchors (deterministic/golden):
        - Build a fixed bundle B; assert `challenge = SHA256("CHALv1"||B)`, `tx_id = SHA256("TXIDv1"||B)` match expected hex; `challenge != tx_id`; B’ ≠ B flips anchors.
      - Finish handler mapping (table):
        - No cookie/expired auth (401); expired tx session (401); challenge mismatch (401); allowlist mismatch (401); type mismatch (400); bad base64 (400); bad JSON (400); RP/Origin not allowed (403 via policy); signCount not increasing (409). Ensure tx session is deleted only on success.
      - Single‑use & replay:
        - After success, second POST with same `tx_session_id` → 401; parallel double‑finish → exactly one succeeds; other fails (401/409). Run with `-race`.
      - WebAuthn/UV policy:
        - AD without UV bit with otherwise valid assertion → mapped to 403; finish handler maps accordingly.
      - Fuzzing:
        - `FuzzParseClientDataJSON` and `FuzzParseAuthData`: no panics; expected errors for malformed; consistency for valid.
        - `FuzzDecodeCanonical` (encoding) and ToECDSA fuzz: no panics; valid cases remain stable.
      - HTTP/session refresh:
        - With `refresh=true`, near‑expiry request extends `expires_at` (assert delta).
      - Storage constraints:
        - Foreign‑key violations rejected; PRAGMAs effective.
      - Types/bundle:
        - Roundtrip with/without `valid_until`; canonical key order; omission of zero values.

    - Advanced Invariants
      - Canonical CBOR uniqueness: re‑encode of parsed bundle B equals original canonical B.
      - Anchor sensitivity: small change in message/nonce changes both anchors.
      - Strict signCount: `ad.SignCount` must be strictly greater than stored; equal/lower = 409.
      - Account binding: canonical CBOR of `sender_key` equals stored `acct_cbor` byte‑for‑byte.
      - Allowlist: `rawId` must be present in options allowlist; otherwise 401.
      - UV required: `HasUV(ad.Flags)` must be true even if signature verifies.

    - Test Harness Utilities
      - Deterministic COSE generators (base point or from ecdsa key → padded 32‑byte coords).
      - `openMemDB()` + `storage.Migrate()` per test; `t.Cleanup` handles closing.
      - Builders: `mkBundle()`, `encodeCanonical()`, `mkAD(flags, signCount)`, `mkCDJ(challenge,type,origin)`.
      - Time helpers: closures to advance `now` for expiry tests.

    - Files To Add
      - `server/internal/tx/anchors_test.go`: golden anchors for B; inequality for B’ ≠ B; `challenge != tx_id`.
      - `server/internal/tx/finish_handler_negative_test.go`: table‑driven error→status mapping; single‑use deletion.
      - `server/internal/tx/replay_race_test.go`: two goroutines finishing same session; exactly one success.
      - `server/internal/webauthn/cdj_fuzz_test.go`, `ad_fuzz_test.go`: fuzzers.
      - `server/internal/encoding/cbor_fuzz_test.go`: fuzz decode/encode properties.
      - `server/internal/crypto/cose_fuzz_test.go`: fuzz ToECDSA inputs.
      - `server/internal/http/session_refresh_test.go`: expiry extension when `refresh=true`.
      - `server/internal/types/bundle_roundtrip_test.go`: optional `valid_until` behavior and canonical order.

    - Execution
      - Unit tests: `go test ./server/...` and with race: `go test -race ./server/...`.
      - Fuzz (local/time‑boxed CI): `go test -run=^$ -fuzz=Fuzz ./server/internal/webauthn` (and encoding/crypto).
      - Focused: `go test ./server/internal/tx -run Anchors` etc.

    - Acceptance Criteria
      - New tests compile and pass locally and under `-race`.
      - Fuzz targets stable (no panics for arbitrary input in allotted time).
      - Critical invariants are enforced via tests listed above.
      - Existing tests continue to pass.

    - Open Questions
      - Scope/timebox for fuzzing in CI vs nightly.
      - Body limit/rate‑limit envelopes align with Step 38.

    - Refs
      - Refs: requirement R-FLOW-SIGN; requirement R-SEC-UV; requirement R-ERR; goal server-derived-challenge-and-txid; decision encoding-and-ceremony-guardrails; spec bundle-anchors-explainer; spec assertion-signature-verification-explainer

