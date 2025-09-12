# Step 44 — Env sample (Done: 2025-09-12)

## Scope
- Provide a `.env.example` with the minimal environment variables required to run the backend locally.

## Source to add/modify
- Add `.env.example` at repo root with variables:
  - `RP_ID` (default `localhost`)
  - `ORIGIN` (default `http://localhost:5173`)
  - `PORT` (default `8080`)
  - `DB_PATH` (default `server/demo.db`)
  - Optional: `LOG_LEVEL`, `LOG_FORMAT`

## Description files
- N/A (env sample only).

## Request/response shape
- N/A.

## Algorithm
- Author a commented file with safe defaults for local dev; avoid secrets.

## Database interactions
- N/A.

## Policies & limits
- Do not commit actual secrets; the file is a sample. Users copy to `.env` locally.

## Sequencing
- Complements Step 41 `make run` usage; can be sourced via `set -a; source .env; set +a`.

## Tests
- Manual source and run.

## Verification
- `test -f .env.example` returns 0.
- Export and run server locally using the sample.

## User verification commands
```bash
# Verify the file exists
 test -f .env.example && echo OK

# Export envs and run
 set -a; source .env.example; set +a
 make run
```

## Acceptance criteria
- `.env.example` present with clear defaults and comments for local dev.

## Notes
- In CI, prefer explicit env configuration or a separate `.env.ci` (out of scope here).

## Refs
- Refs: requirement R-OPS-DEV; requirement R-PLAT-2

