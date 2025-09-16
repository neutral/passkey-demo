# Passkey Demo (WebAuthn) — Node API + React UI

This repository is a minimal demo of a **WebAuthn Passkey** flow: register, login, and sign a small content bundle. It’s organized as a monorepo with a **Node HTTP API** (`node-server/`) built on Express + SQLite and a **React (Vite) UI** (`web/`). Project goals, requirements, specs, decisions, and the implementation plan live under `blueprint/` to keep requirements traceable to code and tests.

---

## Prerequisites

- **Node** ≥ 18 and **npm**

  ```bash
  node -v && npm -v
  ```

- **Git** (optional, but recommended for version control)

- **C toolchain** (required by `better-sqlite3`)
  - macOS: `xcode-select --install`
  - Ubuntu/Debian: `sudo apt-get update && sudo apt-get install -y build-essential`
  - Fedora: `sudo dnf install -y @development-tools`
  - Windows: use WSL2 (Ubuntu) and install `build-essential`

---

## Repository Structure

```
.
├─ node-server/               # Node/Express API + SQLite schema/tests
│  ├─ src/
│  │  ├─ server.js            # entrypoint (Express wiring)
│  │  ├─ config.js            # env config loader
│  │  ├─ db.js                # SQLite open + migrations
│  │  ├─ limits.js            # body + rate limit middleware
│  │  ├─ logger.js            # structured logging helpers
│  │  ├─ session.js           # cookie-backed sessions
│  │  ├─ webauthn/            # registration/login helpers
│  │  └─ tx/                  # transaction signing routes/helpers
│  ├─ test/                   # node --test suites
│  └─ package.json
├─ web/                       # React (Vite) SPA
│  ├─ src/
│  │  ├─ pages/               # Register, Login, Dashboard
│  │  └─ lib/                 # API + WebAuthn helpers
│  ├─ tests/                  # Playwright specs (unit + E2E)
│  ├─ vite.config.ts          # dev server config
│  └─ package.json
├─ blueprint/                 # goals, requirements/specs, decisions, tasks
├─ docs/                      # API examples, Postman collection, REST client
├─ tools/                     # repo tooling (desc-check, ck search guide, etc.)
├─ Makefile                   # dev/build/test helpers
└─ .env.example               # sample environment configuration
```

---

## Configuration

### Root `.env` (optional)

```
RP_ID=localhost
ORIGIN=http://localhost:5173
PORT=8080
DB_PATH=demo.db
```

Notes:
- `RP_ID` must match the effective domain used by the browser for WebAuthn (e.g., `localhost` for local dev).
- `ORIGIN` should point at the UI host so the server can enforce origin checks.
- `DB_PATH` is the SQLite file used by the Node backend (relative paths are resolved from the server’s working directory).

### Web (`web/.env.development`)

```
VITE_API_BASE=/api
```

`vite.config.ts` proxies `/api` to the Node backend during development so the browser does not hit CORS issues.

---

## Install Dependencies

```bash
npm ci --prefix node-server
npm ci --prefix web
```

---

## Run (Development)

Single terminal:

```bash
cp -n .env.example .env || true
set -a; source .env 2>/dev/null || true; set +a
make run
```

Separate terminals:

```bash
# Terminal A — API
make server   # npm run dev --prefix node-server

# Terminal B — UI
make web      # npm run dev --prefix web
```

Then visit http://localhost:5173 and walk through Register → Login → Dashboard Sign.

---

## Build / Clean

```bash
make build    # runs node-server tests + builds the web bundle
make clean    # removes node-server/*.db and web/dist
```

---

## Tests

```bash
make test       # node-server unit tests (node --test)
make test-all   # node-server tests + Playwright UI suites
```

You can also run individual Playwright suites via `npm run test:ui --prefix web` or single files with `npx playwright test <file>` inside `web/`.

---

## Quickstart

```bash
# 1) Install deps
npm ci --prefix node-server
npm ci --prefix web

# 2) Start both services with defaults
cp -n .env.example .env || true
set -a; source .env 2>/dev/null || true; set +a
make run
```

API: http://localhost:${PORT:-8080} (health: `/health`)
UI: http://localhost:5173

---

## Blueprint

- `blueprint/goals.md`: project goals and design tenets.
- `blueprint/_decisions/`: ADRs (architecture decisions).
- `blueprint/_user-flows/`: canonical flows for register/login/sign.
- `blueprint/implementation.md`: implementation plan (source of truth for tasks, Done sections link to completed steps).

---

## API Examples & Tools

- Copy/paste curls: `docs/api-examples.md`.
- Postman collection: `docs/postman/passkey-demo.postman_collection.json` (env: `docs/postman/local.postman_environment.json`).
- VS Code REST Client: `docs/rest-client/api.http`.

---

## WebAuthn Notes (Dev)

- Browser prompts (e.g., Touch ID) appear during register/login; ensure `RP_ID`/`ORIGIN` match.
- The Node backend enforces UV, resident keys, challenge binding, and transaction bundle policies.
- Structured JSON logs include correlation ids and hashed identifiers; set `LOG_FORMAT=text LOG_LEVEL=debug` for human-readable output.

Happy passkey hacking! :key:
