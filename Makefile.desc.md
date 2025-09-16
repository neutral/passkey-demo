# Purpose
Top-level developer convenience targets to build, run, and test the project from the repo root.

# Key Logic
- `server` / `node-server`: run the Node API (`npm run dev --prefix node-server`).
- `web`: run the Vite dev server (`npm run dev --prefix web`).
- `build`: run Node server tests then build the web bundle.
- `clean`: remove local SQLite DBs in `node-server/` and the web build output.
- `run`: start Node server and web concurrently; traps INT/TERM and stops both; surfaces logs in one terminal.
- `test`: run Node server unit tests.
- `test-all`: run Node server tests and Playwright UI tests.

# Interactions
- `run` relies on environment variables for the Node server (`RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`); defaults are provided for local dev.
- Test targets wrap `npm run test`/`npm run test:ui` for consistent execution.

# Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-2
