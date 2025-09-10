### Step 34 — Error Toasts (Done: 2025-09-10)

Verification notes
- Implemented `ErrorToast` component and `parseHttpError`/`normalizeError` helpers.
- Wired into Register, Login, and Dashboard without breaking existing negative-flow assertions.
- Added Playwright spec `web/tests/error-toasts.spec.ts` covering 400/401/413/429 scenarios; verified expected messages.

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

