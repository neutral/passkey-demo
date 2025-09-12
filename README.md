# Passkey Demo (WebAuthn) — Go API + React UI

This repository is a minimal demo of a **WebAuthn Passkey** flow: register, login, and sign a small content bundle. It’s organized as a monorepo with a **Go HTTP API** (`server/`) and a **React (Vite) UI** (`web/`). A Vite dev proxy is set up so the UI can call the API during development without CORS hassles.

Project goals, requirements, specs, decisions, and the implementation plan live under `blueprint/` to keep requirements traceable to code and tests.

---

## Prerequisites

- **Go** ≥ 1.21

  ```bash
  go version
  ```

- **Node** ≥ 18 and **npm**

  ```bash
  node -v && npm -v
  ```

- **C toolchain** (required by `go-sqlite3`)

  - macOS: `xcode-select --install`
  - Ubuntu/Debian: `sudo apt-get update && sudo apt-get install -y build-essential`
  - Fedora: `sudo dnf install -y @development-tools`
  - Windows: use WSL2 (Ubuntu) and install `build-essential`

- **Git** (optional, for version control)

---

## Repository Structure

```
.
├─ server/
│  ├─ cmd/api/main.go          # API entrypoint (slog, router)
│  ├─ internal/
│  │  ├─ app/                  # router builder and wiring
│  │  ├─ httpx/                # middleware (request id, cors, session, errors)
│  │  ├─ webauthn/             # ceremonies, policy, verification, utilities
│  │  ├─ tx/                   # transaction bundle, options/finish, listing
│  │  ├─ repos/                # prepared-statement repositories
│  │  ├─ storage/              # SQLite open/migrate helpers
│  │  ├─ types/                # shared types (COSE, options)
│  │  ├─ encoding/             # base64url, canonical CBOR helpers
│  │  └─ util/                 # small utilities (rand, ttl store)
│  ├─ go.mod
│  ├─ go.sum
│  └─ (config via env vars)
├─ web/
│  ├─ src/
│  │  ├─ pages/                # Register, Login, Dashboard
│  │  └─ lib/                  # helpers
│  ├─ vite.config.ts           # dev proxy: /api -> http://localhost:8080
│  ├─ package.json
│  └─ .env.development         # VITE_API_BASE=/api
├─ contracts/                  # optional shared schemas/types
├─ blueprint/                  # goals, requirements/specs, decisions, tasks
├─ Makefile                    # dev/build helpers
└─ .gitignore
```

---

## Configuration

### Root `.env` (optional)

```
PORT=8080
RP_ID=localhost
ORIGIN=http://localhost:5173
DB_PATH=server/demo.db
```

Notes:
- The server does not auto-load `.env`; Quickstart sources it before `make run` for convenience.
- `RP_ID` must match the effective domain used by the browser for WebAuthn (e.g., `localhost` for local dev; a real domain in production).

### Web (`web/.env.development`)

```
VITE_API_BASE=/api
```

### Dev Proxy (already set)

`web/vite.config.ts` proxies `^/api` to `http://localhost:8080` so you don’t need CORS during development.

---

## API Overview (current)

Core endpoints (see `docs/api-examples.md` and Postman collection for examples):

- Health: `GET /health`
- Registration: `POST /authn/passkey/registration/options`, `POST /authn/passkey/registration/finish`
- Login: `POST /authn/passkey/login/options`, `POST /authn/passkey/login/finish`
- Transactions: `POST /tx/signing/options`, `POST /tx/signing/finish`, `GET /tx/list`

---

## Install Dependencies

From the repo root:

```bash
# Server deps (CBOR + SQLite driver are already referenced in go.mod)
cd server
go mod tidy
cd ..

# Web deps
cd web
npm install
cd ..
```

---

## Run (Development)

You can either run both together (recommended) or separate terminals:

Together (single terminal):

```bash
cp -n .env.example .env || true
set -a; source .env 2>/dev/null || true; set +a
make run
```

Separate terminals:

```bash
# Terminal A — API
make server  # http://localhost:8080 (health: /health)

# Terminal B — UI
make web     # http://localhost:5173 (proxying /api → :8080)
```
Then visit: http://localhost:5173

---

## Build

```bash
# Build both API binary and UI production bundle
make build

# Clean artifacts (removes web/dist and server/*.db)
make clean
```

Build outputs:

- API binary in `server/` (from `go build ./cmd/api`)
- UI static files in `web/dist`

---

## Changing Ports / Origin

- To change the **UI dev port**, edit `web/vite.config.ts` (`server.port`) and update `ORIGIN` for the server accordingly.
- To change the **API port**, export `PORT` before running:

  ```bash
  cd server
  export PORT=9090 ORIGIN=http://localhost:5173 RP_ID=localhost
  go run ./cmd/api
  ```

---

## Common Issues

- **`sqlite3` compile errors**: ensure a C toolchain is installed (see prerequisites).
- **Port in use**: adjust ports as noted above.
- **CORS errors**: use the dev proxy (`/api`), or add CORS handling on the server if calling it directly from a different origin.
- **WebAuthn `RP_ID` mismatch**: the browser’s origin must be a registrable domain that matches `RP_ID` (e.g., `localhost` in dev). Mismatches cause `NotAllowedError`/`SecurityError` during ceremonies.

---

## Quickstart

```bash
# 1) Copy environment sample (optional; defaults are fine for local dev)
cp -n .env.example .env || true

# 2) Run both API and UI together (Ctrl-C to stop both)
set -a; source .env 2>/dev/null || true; set +a
make run
```

API: http://localhost:${PORT:-8080} (health: /health)
UI: http://localhost:5173

### Make Targets

```bash
make run       # start API + UI together (with trap)
make server    # run API
make web       # run UI (Vite dev)
make build     # build API + UI
make clean     # remove db + dist

# Go tests (short suite by default)
make test                  # short tests across all packages
make test PKG=./internal/webauthn RUN=LoginFinish   # filtered
make test-all              # full suite
```

---
## Blueprint

- `blueprint/goals.md`: project goals and design tenets.
- `blueprint/_decisions/`: ADRs (architecture decisions).
 - `blueprint/implementation.md`: implementation plan; Done links to completed steps.

---

## API Examples & Tools

- Copy/paste curls: `docs/api-examples.md`.
- Postman collection: `docs/postman/passkey-demo.postman_collection.json` (env: `docs/postman/local.postman_environment.json`).
- VS Code REST Client: `docs/rest-client/api.http`.

---

## WebAuthn Notes (Dev)

- Browser prompts (e.g., Touch ID) appear during register/login.
- RP and Origin must match: `RP_ID=localhost` and `ORIGIN=http://localhost:5173` are the defaults for local dev.
- The server logs structured JSON (slog). For readable dev logs:

```bash
export LOG_FORMAT=text LOG_LEVEL=debug
```
- `blueprint/global/` and `blueprint/features/`: NFRs/FRs + specs.
- `blueprint/_user-flows/`: canonical user flows.
- `blueprint/implementation.md`: Source of truth for implementation tasks.

Pull tasks from the implementation plan and keep “Refs” current in all artifacts to maintain traceability.

---

That’s it—happy passkey hacking!
