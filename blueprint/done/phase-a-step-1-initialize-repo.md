### Step 1 — Initialize repo (Done: 2025-08-31)

- Structure

  - Ensure `server/`, `web/`, `blueprint/` exist (present in this repo) and root `.gitignore` exists (present).
  - Commands: `ls -la`, `git status` (confirm clean state before adding files).

- Source to add

  - Root `.editorconfig` with: UTF‑8; LF; trim trailing whitespace; insert final newline; 2 spaces for TS/TSX/JS/JSON/MD/YAML; tabs for Go.
    - Example sections: `[*.{ts,tsx,js,json,md,yml,yaml}] indent_size = 2`, `[*.go] indent_style = tab`.

- Description files to add

  - `server/server.desc.md`: backend overview (WebAuthn endpoints, SQLite persistence), relations to frontend and DB; key invariants (UV required, canonical CBOR, low‑S, signCount monotonic).
    - Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-PLAT-3; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails.
  - `web/web.desc.md`: SPA overview (Register/Login/Dashboard; WebAuthn invocations; CBOR bundle build), relations to backend.
    - Refs: goal ui-simplicity-two-buttons; requirement R-PLAT-1; requirement R-UI-2BTN; flows registration/login/transaction-signing.

- Blueprint updates

  - Add a short note under this step with “Refs: goal simple-ui-and-storage; requirement R-PLAT-1; requirement R-PLAT-2; requirement R-OPS-DEV”.
  - Do not mark subsequent steps In‑Progress until their artifacts are Approved.

- Verify
  - `git status` shows only: `.editorconfig`, `server/server.desc.md`, `web/web.desc.md` as changes.
  - `test -f .editorconfig && echo ok` returns `ok`.
  - `ls server server/internal web/src >/dev/null` exits 0.
  - Editor/formatter recognizes `.editorconfig` (spot‑check by saving a TS file and a Go file).

