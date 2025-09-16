### Step Refactor-16 — Web: remove legacy Go backend compatibility

Scope

- Delete remaining Go-specific fallbacks from the web client so adapters/tests only target Node’s SimpleWebAuthn JSON and shared helpers.
- Simplify TypeScript types by removing snake_case/legacy session handling, ensuring all flows rely solely on Node conventions.
- Prune blueprint specs and docs that still reference Go interoperability or dual-backend parity.

Source to modify

- `web/src/lib/webauthn.ts` — drop flattening of `options.*` snake_case fields, remove default user-id generation used only for Go, tighten descriptor handling to accept Node shapes exclusively.
- `web/src/lib/api.ts` — confirm no Go-specific envelope handling remains (if any references to Go codes exist, remove them).
- `web/src/pages/Register.tsx` / `Login.tsx` / `Dashboard.tsx` — eliminate comments or branches referencing Go payloads; ensure runtime guards assume Node responses.
- `web/tests/*` — remove fixtures that mimic Go snake_case options; clean up expectations tied to Go error messaging.
- `web/playwright.config.ts` — ensure comments/documentation mention only Node server (no Go fallback).
- `docs/` and `README` segments mentioning Go backend (if still present) — update to reference Node-only stack.

Description files

- `web/src/lib/webauthn.ts.desc.md`
- `web/src/pages/Register.tsx.desc.md`
- `web/src/pages/Login.tsx.desc.md`
- `web/src/pages/Dashboard.tsx.desc.md`
- `web/tests/*` description files touching updated fixtures (adapter/post-body/signing/error toasts).
- `docs/*.desc.md` or section-specific descriptions updated to reflect Node-only flows.

Blueprint updates

- `blueprint/global/frontend-minimal-react/_specs/frontend-api-base-and-cors.md` — remove references to Go backend, emphasize Node-only support.
- `blueprint/global/easy-local-dev/_specs/spec.md` — confirm local dev narrative is Node-centric with no Go references.
- `blueprint/features/registration-flow/_specs/frontend-mapping-and-pitfalls.md` — strip historical Go notes; reinforce Node JSON mapping assumptions.
- `blueprint/features/login-flow/_specs/allowcredentials-*.md` — ensure examples only show Node camelCase fields and shared helpers.
- `blueprint/_user-flows/*.md` — verify no Go parity callouts remain; update Refs if scope changes.
- If any ADRs still cite Go fallback (e.g., router decisions), append note marking Go path deprecated.

Request/response shape

- No new endpoints; confirm that all documented shapes show Node camelCase JSON only (no `options.*` nesting beyond transaction payload’s `options`).

Algorithm

- Remove conditional branches that convert Go snake_case keys; rely on Node data contract exclusively.
- Simplify option adapters: require presence of Node fields, throw explicit errors when missing rather than “fill in” for Go gap.
- Ensure tests assert absence of legacy keys (e.g., verifying `options` isn’t needed for login, `allow_credentials` no longer supported).

Database interactions

- None (frontend-only refactor).

Policies & limits

- Confirm UV/residentKey requirements remain enforced via Node JSON; remove any fallback that weakened checks.
- Maintain bundle size/nonce policies in dashboard tests with Node error cases.

Sequencing

- Depends on Step 16 completion (Node parity). No backend changes expected; ensure Node server already deployed.
- Coordinate with docs (Step 18) so updates aren’t duplicated; doc changes should reference previous steps.

Tests

- Unit/adapters: `npx playwright test register-adapter.spec.ts login-adapter.spec.ts register-post-body.spec.ts login-post-body.spec.ts`.
- Dashboard/signing/error suites: `npx playwright test tx-signing.spec.ts tx-signing-negative.spec.ts error-toasts.spec.ts`.
- E2E smoke (optional but recommended after refactor): `npx playwright test login-e2e.chromium.spec.ts webauthn-e2e.chromium.spec.ts`.
- `npm -C web run build` to ensure TS stays strict after type removal.

Verification

- Ensure TypeScript build succeeds with no unused-branch warnings.
- Run targeted Playwright suites; rerun failing specs after fixes until green.
- Spot-check docs/blueprint references for lingering Go mentions.

User verification commands

```bash
npm -C web run build
cd web
npx playwright test register-adapter.spec.ts login-adapter.spec.ts register-post-body.spec.ts login-post-body.spec.ts
npx playwright test tx-signing.spec.ts tx-signing-negative.spec.ts error-toasts.spec.ts
npx playwright test login-e2e.chromium.spec.ts webauthn-e2e.chromium.spec.ts
cd ..
git grep -n "Go backend" docs web blueprint || true
```

Acceptance criteria

- Web codebase contains no conditional logic or tests referencing Go backend schemas or behaviors.
- Blueprint/docs describe Node-only interoperability and SimpleWebAuthn shared libraries.
- All targeted tests/build commands pass without regressions.

Notes

- If shared helpers (e.g., `encoding.ts`) still reference Go migrations, document whether to retain for historical context or clean up in later refactor step.
- Coordinate with backend team before removing any compatibility hooks they might still rely on.

Refs: goal passkey-registration-login-uv; goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-OPS-DEV; requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; decision webauthn-corrections-and-standardizations
