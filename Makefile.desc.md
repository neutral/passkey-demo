# Purpose
Top-level developer convenience targets to build, run, and test the project from the repo root.

# Key Logic
- `server`: run the Node API (`npm run dev --prefix node-server`).
- `web`: run the Vite dev server (`npm run dev --prefix web`).
- `build`: build the web bundle (`npm run build --prefix web`).
- `clean`: remove local SQLite DBs in `node-server/` and the web build output.
- `run`: start Node server and web concurrently; traps INT/TERM and stops both; surfaces logs in one terminal.
- `test`: run Node server unit tests.
- `test-ui`: run Playwright UI suites.
- `test-all`: run both `test` and `test-ui` targets for convenience.

# Interactions
- `run` relies on environment variables for the Node server (`RP_ID`, `ORIGIN`, `PORT`, `DB_PATH`); defaults are provided for local dev.
- Test targets wrap `npm run test` / `npm run test:ui` for consistent execution.

# Refs
Refs: requirement R-OPS-DEV; requirement R-PLAT-2
