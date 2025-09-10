# Step-by-Step Implementation Plan (granular; compile- & verify-friendly)

> The steps assume a mono‑repo with `server/` (Go) and `web/` (React + Vite) directories. Each step yields a compilable state and a simple verification method.

## Done (One liner)

Completed steps are stored as individual files under `blueprint/done/`, with phase information prefixed in each filename (e.g., `phase-a-step-1-...md`). Simply add a single line summary here for all done steps with their number and link to the done file.

- [Step 1 — Initialize repo](blueprint/done/phase-a-step-1-initialize-repo.md)
- [Step 2 — Server Go module](blueprint/done/phase-a-step-2-server-go-module.md)
- [Step 3 — Add deps](blueprint/done/phase-a-step-3-add-deps.md)
- [Step 4 — Web app scaffold](blueprint/done/phase-a-step-4-web-app-scaffold.md)
- [Step 5 — Server config struct](blueprint/done/phase-b-step-5-server-config-struct.md)
- [Step 6 — DB init & migrations](blueprint/done/phase-b-step-6-db-init-and-migrations.md)
- [Step 7 — Types: COSE, WebAuthn, Bundle](blueprint/done/phase-b-step-7-types-cose-webauthn-bundle.md)
- [Step 8 — COSE → ECDSA helper](blueprint/done/phase-b-step-8-cose-to-ecdsa-helper.md)
- [Step 9 — Base64url utilities](blueprint/done/phase-b-step-9-base64url-utilities.md)
- [Step 10 — CBOR canonical codec](blueprint/done/phase-b-step-10-cbor-canonical-codec.md)
- [Step 11 — Parse authenticatorData](blueprint/done/phase-c-step-11-parse-authenticatordata.md)
- [Step 12 — ClientDataJSON validation](blueprint/done/phase-c-step-12-clientdatajson-validation.md)
- [Step 13 — Signature verify utility](blueprint/done/phase-c-step-13-signature-verify-utility.md)
- [Step 14 — RP ID & Origin checks](blueprint/done/phase-c-step-14-rp-origin-checks.md)
- [Step 15 — /authn/passkey/registration/options handler](blueprint/done/phase-d-step-15-registration-options-handler.md)
- [Step 16 — Attestation parsing (minimal)](blueprint/done/phase-d-step-16-attestation-parsing.md)
- [Step 17 — /authn/passkey/registration/finish handler](blueprint/done/phase-d-step-17-registration-finish-handler.md)
- [Step 18 — /authn/passkey/login/options handler](blueprint/done/phase-e-step-18-login-options-handler.md)
- [Step 19 — /authn/passkey/login/finish handler](blueprint/done/phase-e-step-19-login-finish-handler.md)
- [Step 20 — Bundle validation helper](blueprint/done/phase-f-step-20-bundle-validation-helper.md)
- [Step 21 — /tx/signing/options handler](blueprint/done/phase-f-step-21-signing-options-handler.md)
- [Step 22 — /tx/signing/finish handler](blueprint/done/phase-f-step-22-signing-finish-handler.md)
- [Step 23 — /tx/list handler](blueprint/done/phase-f-step-23-tx-list-handler.md)
- [Step 24 — Session middleware](blueprint/done/phase-g-step-24-session-middleware.md)
- [Step 25 — Rate limiting & limits](blueprint/done/phase-g-step-25-rate-limiting-limits.md)
- [Step 26 — CORS & cookies](blueprint/done/phase-g-step-26-cors-and-cookies.md)
- [Step 27 — Basic pages](blueprint/done/phase-h-step-27-basic-pages.md)
- [Step 28 — Base64url helpers (web)](blueprint/done/phase-h-step-28-base64url-helpers-web.md)
- [Step 29 — WebAuthn create() flow (Register)](blueprint/done/phase-h-step-29-webauthn-register.md)
- [Step 30 — WebAuthn get() flow (Login)](blueprint/done/phase-h-step-30-webauthn-login.md)
- [Step 31 — Dashboard: fetch list](blueprint/done/phase-h-step-31-dashboard-fetch-list.md)
- [Step 32 — Dashboard: build bundle](blueprint/done/phase-h-step-32-dashboard-build-bundle.md)
- [Step 33 — Dashboard: transaction signing](blueprint/done/phase-h-step-33-dashboard-transaction-signing.md)

## Next

### Phase H — Frontend (React) UI

---

34. **Error Toasts**

    - Purpose: Surface backend/frontend errors consistently as small, dismissible toasts across Register, Login, and Dashboard, without blocking flows. Best-effort mapping now; standardized envelopes come in Step 38.

    - Scope
      - In-scope: Page-local toasts; parse Response to show status + JSON error fields if present; graceful fallback to status text or thrown error message; minimal styling in `App.css`.
      - Out-of-scope: Global error bus/store; i18n; standardized error envelopes and codes (deferred to Step 38); retries/backoff; analytics.

    - Non-Goals
      - Do not change backend payload shapes now.
      - Do not introduce cross-page shared state or a third-party toast lib.

    - Source To Add/Modify
      - Add: `web/src/components/ErrorToast.tsx`
        - Props: `{ title?: string; detail?: string; code?: string; status?: number; onClose?: () => void }`
        - Behavior: Renders a small, dismissible banner (role="alert") at top-right; auto-dismiss optional future enhancement (not required here).
      - Add: `web/src/components/ErrorToast.tsx.desc.md`
      - Add: `web/src/lib/http.ts`
        - `async parseHttpError(r: Response): Promise<{ title: string; detail: string; code?: string; status: number }>`
          - Attempts JSON parse; picks `error || message || detail || title || statusText`.
          - Copies `code` if present; includes `status`.
        - `normalizeError(e: unknown): { title: string; detail: string }`
          - For network/throw cases: prefers `e.message` else stringified `e`.
      - Add: `web/src/lib/http.ts.desc.md`
      - Modify: `web/src/pages/Register.tsx`
        - Replace inline error paragraph with `<ErrorToast>` and use `parseHttpError` for non-200 responses; `normalizeError` for thrown errors.
      - Modify: `web/src/pages/Login.tsx`
        - Same as above; maintain existing messages for explicit branch errors so current tests remain stable.
      - Modify: `web/src/pages/Dashboard.tsx`
        - Replace error paragraph with `<ErrorToast>` using existing `error` state string as fallback; where fetch returns `!ok`, switch to `parseHttpError` where we don’t already set specific messages (keep current explicit texts for 401/409/400 as-is to preserve tests).
      - Modify: `web/src/App.css`
        - Add `.toast` styles (position, background, border, shadow, focus outline, reduced motion friendly).
      - Update Descriptions:
        - `web/src/pages/Register.tsx.desc.md`, `web/src/pages/Login.tsx.desc.md`, `web/src/pages/Dashboard.tsx.desc.md` to mention toast-based errors.
        - `web/src/components/ErrorToast.tsx.desc.md`, `web/src/lib/http.ts.desc.md`.

    - Interfaces
      - `parseHttpError(Response)`:
        - Returns `{ status, code?, title, detail }`
        - JSON priority: `error | message | detail | title`; else `r.statusText || 'Request failed'`.
      - Error Toast props as above.

    - Data/Models
      - No persisted data. Pure UI/view-model additions.

    - Algorithms
      - Fetch failure path:
        - If `!r.ok` → `const err = await parseHttpError(r)` → set toast state `{ status, code, title, detail }`.
        - If thrown (network/abort) → `set toast = normalizeError(e)`.
      - Display:
        - Render `title` emphasized, include `code` and `status` if present, show `detail` on next line.
        - Provide a close button; pressing Esc focuses it via standard tab sequence.

    - Policies & Behaviors
      - Do not read cookies; rely on HTTP statuses, consistent with prior steps.
      - Preserve existing explicit error strings that tests assert against (e.g., “Error: invalid bundle”, “finish: HTTP 409”) to avoid breaking tests; only apply parser where messages aren’t hardcoded.
      - Prefer concise messages; no stack traces to users.

    - Accessibility
      - `role="alert"`, `aria-live="assertive"`, clear contrast, keyboard focusable close button, no motion/transitions necessary.

    - Risks
      - Overriding existing textual expectations could break tests. Mitigation: preserve explicit message branches; introduce toast as a visual container only.
      - Inconsistent backend error shapes. Mitigation: defensive parser with robust fallbacks.

    - Testing Strategy
      - Unit-ish (Playwright page context):
        - Import `parseHttpError` and test against samples:
          - JSON: `{ "error": "invalid bundle", "code": "ERR_BUNDLE" }` → picks error/code.
          - JSON: `{ "message": "rate limited" }` → picks message.
          - Non-JSON: 500 text → uses statusText fallback.
      - UI (Playwright):
        - New: `web/tests/error-toasts.spec.ts`
          - Register: mock `registration/finish` 400 JSON `{ error: "invalid attestation" }` → toast appears with “invalid attestation”.
          - Login: mock `login/finish` 401 → toast with “HTTP 401” present.
          - Dashboard Build: mock `/me/account_key` 401 → toast and unauthorized prompt can coexist; verify toast visible.
          - Dashboard Sign:
            - 413 for `/tx/signing/options` → toast shows “HTTP 413”.
            - 429 for `/tx/signing/options` JSON `{ message: "rate limited", code: "RATE_LIMIT" }` → toast shows “rate limited” and code.
        - Backwards-compat checks:
          - Existing negative signing tests still pass (strings kept).
      - Snapshot assertions avoided; rely on text visibility and role="alert".

    - Verification
      - Manual:
        - Simulate failures via browser devtools or by running backend with knobs (if available). Confirm toasts appear, dismiss works, and no crash.
      - Commands:
        - `npm -C web run test:ui -- -g "error toast"` to run new spec subset.
        - Full: `npm -C web run test:ui`

    - Acceptance Criteria
      - Register, Login, Dashboard render a visible, dismissible error toast on non-2xx responses and thrown errors.
      - Toast text shows meaningful info:
        - Includes HTTP status (e.g., 401/409/413/429) when applicable.
        - Uses JSON `error | message` if present; otherwise falls back gracefully.
      - Existing negative-flow tests continue to pass (no breaking text changes).
      - No global state introduced; no additional dependencies.

    - Open Questions
      - Auto-dismiss after N seconds? Proposed: leave off for now; add in a later UX step.
      - Multiple simultaneous toasts? For now, replace the current toast; queueing is out-of-scope.

    - User Verification Commands
      - Start servers and run tests:
        - `RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 go -C server run ./cmd/api`
        - `npm -C web run dev`
        - In another terminal: `npm -C web run test:ui -- -g "error toast"`

    - Rollback Plan
      - Retain paragraph-based error rendering code paths behind the toast state; if necessary, remove `ErrorToast` imports and rely on existing `error` string paragraphs.

    - Done Artifact Name
      - `blueprint/done/phase-h-step-34-error-toasts.md`

    - Refs
      - Refs: requirement R-ERR (global/error-responses-and-limits); requirement R-PLAT-1; goal simple-ui-and-storage; decision webauthn-corrections-and-standardizations; spec frontend-api-base-and-cors

### Phase I — Testing & Fixtures

---

35. **Unit tests (Go)**

    - COSE→ECDSA, CDJ parse, AD parse, low‑S check, canonical encoding, hash anchors.
    - _Verify_: `go test ./...` passes.

---

36. **Golden vectors**

    - Create a tiny CLI in `server/cmd/vectors` that:

      - Loads a known COSE key, builds a sample bundle, prints hex(B), challenge, tx_id.

    - _Verify_: consistent outputs between runs.

---

37. **Manual E2E**

    - Run server + web; perform registration, login, and sign two messages with nonces 1, 2 in Safari and Chrome on macOS.
    - _Verify_: `/tx/list` shows two entries; DB reflects persisted rows; flows succeed in both browsers with platform authenticators.

### Phase J — Hardening (Demo-grade)

---

38. **Input validation & errors**

    - Enforce message length ≤ 1024; nonce ≤ `2^53 - 1`; origin/rpId in allowlist; body size ≤ 64 KB.
    - Implement standardized error envelope `{ code, error, correlation_id? }` and map to HTTP 400/401/403/409/413/429/5xx per R-ERR.
    - _Verify_: oversize blocked with 400/413; invalid/replay/UV/origin issues map to correct codes; frontend displays `code`.

---

39. **Session security**

    - Random session IDs (≥128 bits), expiry 1h, renewal on activity.
    - _Verify_: expired session returns 401; re-login works.

---

40. **Logging**

    - Structured logs with event names: `reg_options`, `reg_finish`, `login_options`, `login_finish`, `tx_options`, `tx_finish`.
    - _Verify_: logs show account thumb and tx_id (hex).

---

41. **Build scripts**

    - Root Makefile: `make server`, `make web`, `make run`, `make clean`.
    - _Verify_: one command runs both.

### Phase K — Developer Experience

---

42. **API examples**

    - Add `docs/` with curl examples for each endpoint (sans WebAuthn ceremony).
    - _Verify_: docs render in repo.

---

43. **Postman / REST Client file**

    - Provide a collection with placeholders; helpful for observing JSON shapes.
    - _Verify_: collection can be imported.

---

44. **Env sample**

    - `.env.example` with `RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`.
    - _Verify_: loads correctly.

---

45. **README**

    - Quickstart, limitations (no attestation trust), and demo notes (Touch ID prompts).
    - _Verify_: teammate can bootstrap in <10 minutes.

---

## Future

- Handler-level logging and error mapping integration

  - What: Wire a minimal JSON logger using Go `log/slog` in `server/cmd/api` and apply the `MapVerifyError` and `LogAssertion` utilities from `server/internal/webauthn` in the assertion-finish handler. Keep client responses generic while emitting structured, privacy-preserving logs with stable `error_kind` values from sentinel errors.
  - Why: Improves observability, incident triage, and auditability without leaking sensitive data. Cleanly separates transport concerns (HTTP codes) from cryptographic failure semantics via sentinel errors, enabling accurate metrics and alerts (e.g., spikes in `ErrMalformedDER`).
  - How: Initialize `slog` with a JSON handler for dev; in the handler, call `VerifyAssertion(...)`, map the error with `MapVerifyError`, log once via `LogAssertion` using hashed identifiers (`HashID`), and return an appropriate status code with a standard error envelope. This remains compatible with the existing plan’s later steps for endpoints and error envelopes.

- IDNA (punycode) normalization support
  - What: Normalize internationalized domain names to ASCII (punycode) for RP ID and origin comparisons using `golang.org/x/net/idna`.
  - Why: Prevent mismatches and policy bypass due to Unicode vs punycode inconsistencies; ensure consistent hashing for rpIdHash and accurate origin validation across i18n domains.
  - How: Add a deterministic normalization helper that converts Unicode hostnames to punycode before validation/comparison; guard with unit tests and vectors (e.g., `bücher.ch` ⇄ `xn--bcher-kva.ch`), and document deployment guidance to keep config values consistent.

## Acceptance Checks

- **Build**: `go build ./server/...` and `npm run build` in `/web` both succeed.
- **Register/Login**: Using macOS with Touch ID, both ceremonies prompt for fingerprint; login sets cookie.
- **Sign**: Create 2 messages with nonces 1 and 2; both appear in `/tx/list` and DB.
- **Ephemeral TTLs**: Registration/login/tx option sessions expire after 5 minutes; expired attempts return 409.
- **Replay/Nonce**: Re-submit nonce 2 → **409** conflict.
- **UV check**: If browser returns an assertion without UV (simulate by forcing options incorrectly) → **403** forbidden.
- **Origin/RP guard**: Change `origin` in request body → **403**.
- **Low‑S enforced**: Hand-craft a signature with high‑S (unit test) → **400**.

## Coverage & Refs (Traceability)

- R-FLOW-REG: Steps 11–17, 26, 38. Refs: requirement R-FLOW-REG; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails.
- R-FLOW-LOGIN: Steps 11–14, 18–19, 24, 26, 38–39. Refs: requirement R-FLOW-LOGIN; decision webauthn-corrections-and-standardizations.
- R-FLOW-SIGN: Steps 20–23, 32–33, 26, 38. Refs: requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails.
- R-ID-KEY: Steps 7–8, 16–17, 19, 22–23. Refs: requirement R-ID-KEY; decision webauthn-corrections-and-standardizations.
- R-SCHEMA-LITE: Steps 10, 20, 32. Refs: requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails.
- R-UI-2BTN: Steps 27, 31–34. Refs: requirement R-UI-2BTN.
- R-PLAT-1: Steps 4, 27–34. Refs: requirement R-PLAT-1.
- R-PLAT-2: Steps 5, 11–26, 38–40. Refs: requirement R-PLAT-2.
- R-PLAT-3: Steps 6, 17, 19, 22–23, 34–35. Refs: requirement R-PLAT-3.
- R-OPS-DEV: Steps 5, 26, 41, 44–45. Refs: requirement R-OPS-DEV.
- R-ERR: Steps 25, 34, 38, 40. Refs: requirement R-ERR; decision encoding-and-ceremony-guardrails.
- R-NO-BROKER: Entire plan avoids brokers; synchronous calls. Refs: requirement R-NO-BROKER.
- R-PORTABLE: Steps 26, 27–34, 37. Refs: requirement R-PORTABLE.
- R-SEC-UV: Steps 14–15, 17–19, 21–22. Refs: requirement R-SEC-UV; decision webauthn-corrections-and-standardizations.
