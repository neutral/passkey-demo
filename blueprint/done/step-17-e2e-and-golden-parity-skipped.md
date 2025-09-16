# Step 17 — E2E and golden parity (Skipped)

## Completion
- Date: 2025-09-16
- Outcome: Skipped after verification; existing Node unit tests and Playwright coverage already enforce golden parity.

## Purpose
- Document verification results showing golden parity already enforced; skipping additional tooling.

### Step 17 — E2E and golden parity

#### Scope
- Reassess whether new tooling is required: existing Node unit tests (`node-server/test/tx-bundle.test.js`, `tx-options.test.js`, `tx-finish.test.js`) already load `specs/goldens/tx-bundle-v1.json` and assert anchors; only build a standalone CLI if stakeholders still expect a manual parity check.
- Confirm Playwright configuration (`web/playwright.config.ts`) starts both Vite and the Node server using virtual authenticators; extend or adjust cases to cover register → login → sign → list flows when necessary.

#### Source to add/modify
- _Conditional_: add `node-server/scripts/golden-check.js` that imports the bundle helpers, loads `specs/goldens/tx-bundle-v1.json`, compares canonical CBOR bytes plus `challenge`/`tx_id`, and supports a guarded `--update` flag for intentional rewrites.
- Augment Playwright specs (`web/tests/webauthn-e2e.chromium.spec.ts`, `tx-signing.spec.ts`) only if gaps remain after reviewing current coverage; prefer extending scenarios over creating new files.
- Leave `web/playwright.config.ts` unchanged unless orchestration for the Node server differs (ports/timeouts/env vars).

#### Description files
- Create `node-server/scripts/golden-check.js.desc.md` if a CLI is introduced, documenting purpose, inputs, invariants, and failure modes.
- Update `specs/goldens/goldens.desc.md` to mention the Node tooling replacing the archived Go vectors command when applicable.
- Refresh `.desc.md` companions for any modified Playwright specs to reflect new flow coverage.

#### Blueprint updates
- Requirements: verify `blueprint/features/lightweight-transaction-schema/requirement.md` already states golden parity; append acceptance note only if manual CLI becomes a mandated check.
- Specs: extend `blueprint/features/lightweight-transaction-schema/_specs/spec.md` testing strategy with details on the Node unit tests (and optional CLI) relied upon for parity.
- ADRs/User-flows: none anticipated; if CLI introduces operational change, cross-reference existing decisions (e.g., encoding guardrails) instead of drafting a new ADR.

#### Request/response shape
- No new endpoints; ensure `/tx/signing/options|finish` documented shapes continue to match golden expectations (base64url challenge, hex `tx_id`, canonical CBOR bundle `B`).

#### Algorithm
- Golden validation: load JSON fixture → invoke `validateAndAnchorBundle` → assert canonical CBOR bytes match `bundleHex` and derived anchors equal `challenge`/`tx_id`; non-zero exit on mismatch.
- Playwright flows: seed Chromium virtual authenticator (resident key, UV=true) → register (store credential) → login → initiate signing (options) → finish signing → confirm list shows signed transaction.

#### Database interactions
- Playwright runs should point `DB_PATH` to a disposable SQLite file (`node-server/playwright.db`); ensure teardown removes or truncates it between runs to avoid state bleed.

#### Policies & limits
- Confirm UV-required policy remains satisfied in virtual authenticator setup.
- Ensure signing limits (nonce monotonicity, session TTL) retain test coverage via existing negative Playwright specs and Node limit tests.

#### Sequencing
- Execute Node unit tests first for fast feedback on golden regressions.
- Start Playwright servers only after confirming ports 5173/8080 are free; rely on `reuseExistingServer` during local runs, disable in CI.

#### Tests
- Happy path: `npm -C node-server test -- tx-bundle.test.js tx-options.test.js tx-finish.test.js` followed by `npx playwright test --project=chromium webauthn-e2e.chromium.spec.ts tx-signing.spec.ts`.
- Negative/edge cases: `npx playwright test tx-signing-negative.spec.ts login-e2e.chromium.spec.ts`; `npm -C node-server test -- limits.test.js error-envelope.test.js` to confirm policy enforcement.
- Invariants: canonical CBOR round-trip equality; anchors stable vs golden; UI error messages map to HTTP 4xx when failure expected.

#### Verification
- Run targeted Node tests, then full suite (`npm -C node-server test`); rerun after fixes until all pass.
- Execute `npx playwright test --project=chromium` to ensure E2E flows succeed end-to-end; rerun after addressing failures.
- If CLI created, run `node node-server/scripts/golden-check.js` (and `node ... --update` only when intentionally refreshing goldens).

#### User verification commands
```bash
# Golden/unit coverage
npm -C node-server test -- tx-bundle.test.js tx-options.test.js tx-finish.test.js

# Full node-server suite
npm -C node-server test

# Playwright E2E (Chromium only)
npx playwright test --project=chromium webauthn-e2e.chromium.spec.ts tx-signing.spec.ts
```

#### Acceptance criteria
- Golden fixture matches recomputed anchors without needing new artifacts; if a CLI is introduced, it exits 0 when goldens align and documents the update flow.
- Playwright tests demonstrate register → login → sign → list with the virtual authenticator succeeding (or expected failures captured via existing negative specs).
- No undocumented API or policy drift uncovered during verification; any gaps are recorded for future steps.

#### Notes
- Verification showed existing Node unit tests and Playwright specs already enforce golden parity; no new tooling required.
- Replace references to the archived Go golden CLI only after the Node equivalent (if created) is validated.
- Mark step as skipped (documented verification only) when moved to `## Done`.


Refs: requirement R-FLOW-SIGN; spec golden-vectors; spec manual-e2e
