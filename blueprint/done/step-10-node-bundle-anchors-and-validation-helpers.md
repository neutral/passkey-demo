# Step 10 — Node: Bundle anchors and validation helpers (Done)

Completed: 2025-09-16
Verification notes:
- `node --test test/tx-bundle.test.js`
- `npm test`
- `bash tools/desc-check.sh HEAD~1 HEAD`

### Step 10 — Node: Bundle anchors and validation helpers

Scope
- Implement a Node helper mirroring Go’s `ValidateAndAnchorBundle`: decode bundle base64url → CBOR, recover nested COSE sender key fields even when the decoder omits them, canonicalize to bytes `B`, enforce policies (account binding, nonce monotonicity, length/size caps), and derive `challenge`/`tx_id` anchors.
- Expose structured error kinds so upcoming HTTP handlers can translate to envelope codes without re-inspecting the failure.
- Keep helper side-effect free beyond reading nonce history from SQLite; no Express wiring yet.

Source to add/modify
- `node-server/src/tx/bundle.js` (new) — exports `validateAndAnchorBundle`, `BundleValidationError`, and constants for error kinds/prefixes; owns canonical encoder/decoder singletons and COSE fallback logic.
- `node-server/test/tx-bundle.test.js` (new) — Node test coverage for happy path, golden vectors, and policy failures using in-memory SQLite + migrations.

Description files
- `node-server/src/tx/tx.desc.md` (new) — folder overview tying transaction helpers to signing routes; Refs to signing requirements/specs.
- `node-server/src/tx/bundle.js.desc.md` (new) — documents helper responsibilities, error kinds, and DB interaction (nonce lookup); Refs align with signing requirement + CBOR decisions.
- `node-server/test/tx-bundle.test.js.desc.md` (new) — explains test scenarios (golden parity, limits, error mapping) and cites the same Refs.

Blueprint updates
- No requirement/spec/ADR changes: existing `R-FLOW-SIGN` spec + CBOR/COSE ADR already mandate canonical bundles, anchors, and fallback decoding; implementation aligns without doc edits.

Request/response shape
- Not applicable for this helper-only step (no HTTP surface yet).

Algorithm
- Decode `bundle_cbor_b64` with `Buffer.from(value, 'base64url')`; if decode fails raise `bundle_base64` error.
- Use `cbor-x` decoder (`useMaps: true`) to parse the bundle; if nested `sender_key` fields decode empty, probe alternative shapes per ADR (`Map`, object with numeric/string keys, tagged byte strings) and coerce to integers/Uint8Arrays.
- Ensure `nonce` and `message` populate; fallback to loose decode (e.g., decode to generic object) if typed mapping left zeros; reject missing/zero nonce or non-string message.
- Enforce `nonce` integer bounds (`1 <= nonce <= Number.MAX_SAFE_INTEGER`) and message length ≤ 1024 bytes (by `Buffer.byteLength`); record optional `valid_until` when unsigned int.
- Canonicalize the logical bundle via `Encoder({ canonical: true })` and retain as Buffer `B`; copy outputs so downstream handlers cannot mutate shared slices.
- Decode `acct_cbor` canonically to COSE fields and compare (kty, alg, crv, x, y) against bundle sender key using byte equality; mismatch => `sender_key_mismatch` error.
- Query DB for `SELECT MAX(nonce) AS max_nonce FROM transactions WHERE acct_cbor = ?`; if result exists and `nonce <= max_nonce` emit `nonce_not_monotonic` error.
- Derive anchors with `createHash('sha256')` using prefixes `CHALv1`/`TXIDv1` to produce 32-byte Buffers; return logical bundle, canonical `B`, and anchors.
- Wrap policy failures in `BundleValidationError` instances carrying `kind` so callers can map to HTTP codes.

Database interactions
- Read-only prepared statement on `transactions`: `SELECT MAX(nonce) AS max_nonce FROM transactions WHERE acct_cbor = ?`; allow dependency injection for tests (pass statement/stub) while default prepares lazily and caches in module scope.

Policies & limits
- Nonce strictly increasing per account; reject equal or lower values.
- Nonce upper bound: `Number.MAX_SAFE_INTEGER (2^53-1)` and must be positive integer.
- Message length cap 1024 bytes after UTF-8 encoding.
- Accept optional `valid_until` when unsigned int; preserve in canonical encoding without enforcement yet.
- Surface deterministic `BundleValidationError.kind` (`bundle_base64`, `bundle_cbor`, `sender_key_mismatch`, `nonce_not_monotonic`, `message_too_long`, `nonce_out_of_range`).

Sequencing
- Requires prior steps: registration/login already populate accounts/transactions schema (Steps 3–9).
- Helper feeds upcoming `/tx/signing/options`/`finish` (Steps 11–12); export surface must remain stable for those implementations.
- No dependency on future error envelope wiring (Step 14) but keep error kind table ready for mapping.

Tests
- Add `node-server/test/tx-bundle.test.js` covering:
  - Happy path using golden vector (`specs/goldens/tx-bundle-v1.json`): assert canonical `B` base64/hex, anchors (challenge base64url + tx_id hex), parsed bundle fields.
  - Account binding mismatch: modify acct X/Y bytes to trigger `sender_key_mismatch`.
  - Nonce monotonic policy: insert prior transaction with nonce 10, ensure nonce 9/10 fail and 11 succeeds.
  - Message length cap: send 1025-byte UTF-8 string → `message_too_long`.
  - Nonce range: use `Number.MAX_SAFE_INTEGER + 1` → `nonce_out_of_range`.
  - Base64 decode failure and CBOR decode failure cases.
  - Decoder fallback tolerance: craft bundle where `sender_key` encodes with string keys or tagged byte strings to ensure fallback populates fields.
  - Each error surfaces as `BundleValidationError` with expected `.kind`.
- Tests rely on in-memory `better-sqlite3`, `applyMigrations`, and close DB handles after each case.
- Commands: `cd node-server && node --test test/tx-bundle.test.js` for targeted run; `npm test` for regression sweep.

Verification
- After implementation, run targeted test `node --test test/tx-bundle.test.js`; fix issues and rerun until all subtests pass.
- Run full suite `npm test` from `node-server/` to guard against regressions.
- Optionally execute `bash tools/desc-check.sh HEAD~1 HEAD` from repo root to confirm description policy compliance.
- Success criteria: targeted + full test runs exit 0; golden vector assertions match expected `B`, `challenge`, `tx_id`; git status shows only planned files.

User verification commands
```bash
cd node-server
node --test test/tx-bundle.test.js
npm test
cd ..
bash tools/desc-check.sh HEAD~1 HEAD
```

Acceptance criteria
- Helper reproduces Go bundle validation semantics (canonical encoding, nonce policy, anchor derivation) and emits deterministic error kinds for all failure classes.
- Golden vector comparison passes without diff; canonical bytes/anchors align with `specs/goldens/tx-bundle-v1.json`.
- Tests exercise happy path + policy failures; no regressions in existing suites.

Notes
- Leave integration of helper into Express routes for Steps 11–12; keep exported surface minimal but sufficient (function + error type/constants).
- Cache prepared statements/encoders inside module scope to avoid repeated instantiation; document caching in description file.

Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision cbor-cose-interop-and-decoding-fallbacks; decision encoding-and-ceremony-guardrails
