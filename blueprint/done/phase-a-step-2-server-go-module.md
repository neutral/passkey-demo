### Step 2 — Server Go module (Done: 2025-08-31)

- Context

  - Initialize a minimal Go HTTP server with a health endpoint to unblock future backend steps.

- Structure

  - Ensure `server/` exists and contains a Go module.
    - If `server/go.mod` is missing, run: `cd server && go mod init txkit-demo && go mod tidy`.

- Source to add (instructions only)

  - `server/cmd/api/main.go`: package `main`; start `net/http` server on `:8080`; define `/health` handler returning `200 OK` with body `ok` (text/plain).
    - Use `http.NewServeMux()` and `http.ListenAndServe(":8080", mux)`; log a startup line `listening :8080`.
    - Keep constants in-file for now; Step 5 will introduce `config.go` and refactor port/origin.

- Description files to add (instructions only)

  - `server/cmd/api/main.go.desc.md`: Purpose (entrypoint; health), Key Logic (mux, handlers), Interactions (no DB yet), Refs.
    - Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV.

- Blueprint updates

  - Add “Refs: goal simple-ui-and-storage; requirement R-PLAT-2; requirement R-OPS-DEV” under this step after implementation.

- Verification (to run after implementation)

  - Build: `cd server && go build ./...` (expect exit 0).
  - Run dev: `cd server && go run ./cmd/api` (in a separate terminal).
  - Health: `curl -i http://localhost:8080/health` → `HTTP/1.1 200 OK` and body `ok`.

- User verification commands (copy/paste)

  ```bash
  # Build all server packages
  cd server && go build ./... && cd -

  # Run the API server in the background
  cd server
  go run ./cmd/api > /tmp/step2_api.log 2>&1 & echo $! > /tmp/step2_api.pid
  sleep 1

  # Verify health endpoint
  curl -i http://localhost:8080/health

  # Stop the server
  kill $(cat /tmp/step2_api.pid) && rm -f /tmp/step2_api.pid
  tail -n +1 /tmp/step2_api.log | sed -n '1,50p'
  cd -
  ```

- Notes
  - No CORS, cookies, DB, or config yet; these land in later steps (26, 24, 6, 5).
  - Keep the stub minimal to ensure fast builds and clear verification.

