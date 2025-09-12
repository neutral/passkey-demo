# Purpose
Top-level developer convenience targets to build, run, and test the project from the repo root.

# Key Logic
- `server`: run the Go API (`cd server && go run ./cmd/api`).
- `web`: run the Vite dev server (`cd web && npm run dev`).
- `build`: build Go API and web assets.
- `clean`: remove local build artifacts and demo DB.
- `run`: start server and web concurrently; traps INT/TERM and stops both; surfaces logs in one terminal.
- `test`: run Go tests in short mode; supports `PKG` (package path) and `RUN` (regex) filters.
- `test-all`: like `test` but without `-short`.

# Interactions
- `run` relies on environment variables for the server (`RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`); defaults are provided for local dev.
- Test targets wrap `go test` with optional filters for efficient iteration.

# Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-2

