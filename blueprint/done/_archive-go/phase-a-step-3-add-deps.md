### Step 3 — Add deps (Done: 2025-08-31)

- Context

  - Add core dependencies for canonical CBOR encoding/decoding and SQLite persistence to support later server features.

- Structure

  - Work within `server/` Go module; ensure `server/go.mod` exists.
    - If missing, complete Step 2 prerequisites for module init.

- Source to add (instructions only)

  - Add CBOR library: `github.com/fxamacker/cbor/v2@v2.9.0` (canonical options support).
  - Add SQLite driver: `github.com/mattn/go-sqlite3@v1.14.32` (CGO‑based; acceptable for local dev).
  - Command sequence:
    - `cd server && go get github.com/fxamacker/cbor/v2@v2.9.0`
    - `cd server && go get github.com/mattn/go-sqlite3@v1.14.32`
    - `cd server && go mod tidy`

- Description files to add (instructions only)

  - None for this step; description files will accompany code that uses these deps (Steps 6–10 and DB/CBOR helpers).

- Blueprint updates

  - Refs to include upon implementation: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-PLAT-3; requirement R-SCHEMA-LITE.

- Verification (to run after implementation)

  - Inspect module files changed: `git diff -- server/go.mod server/go.sum` (shows added deps and checksums).
  - Build all packages: `cd server && go build ./...` (expect exit 0).
  - Optional: print versions resolved: `cd server && go list -m -json github.com/fxamacker/cbor/v2 github.com/mattn/go-sqlite3`.

- Notes

  - `github.com/mattn/go-sqlite3` requires CGO; macOS/Linux dev environments satisfy this by default. For CI or cross‑compile, consider build tags or `modernc.org/sqlite` in future ADRs (out of scope for demo).
  - Canonical CBOR usage will be implemented in Step 10; no code changes in this step beyond dependency resolution.

- User verification commands (copy/paste)

  ```bash
  # Ensure server module exists
  test -f server/go.mod && echo OK:server go.mod

  # Add dependencies (idempotent if already present)
  cd server
  go get github.com/fxamacker/cbor/v2@v2.9.0
  go get github.com/mattn/go-sqlite3@v1.14.32
  go mod tidy

  # Verify versions and build
  go list -m -json github.com/fxamacker/cbor/v2 github.com/mattn/go-sqlite3 | sed -n '1,80p'
  go build ./...
  cd -
  ```

