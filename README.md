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
│  ├─ cmd/api/main.go          # minimal HTTP server (health check)
│  ├─ internal/
│  │  ├─ http/                 # HTTP handlers (planned)
│  │  ├─ webauthn/             # WebAuthn ceremonies, policy, verifiers (planned)
│  │  ├─ cbor/                 # minimal CBOR helpers for content bundle (planned)
│  │  ├─ store/                # SQLite-backed account/credential storage (planned)
│  │  └─ config/               # config/env loading (planned)
│  ├─ go.mod
│  ├─ go.sum
│  └─ .env                     # PORT, RP_ID, ORIGIN, DB_PATH (reference)
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

### Server (`server/.env`)

```
PORT=8080
RP_ID=localhost
ORIGIN=http://localhost:5173
DB_PATH=./demo.db
```

Notes:
- The minimal server does **not** auto-load `.env`. Export vars in your shell if you change them.
- `RP_ID` must match the effective domain used by the browser for WebAuthn (e.g., `localhost` for local dev; a real domain in production).

### Web (`web/.env.development`)

```
VITE_API_BASE=/api
```

### Dev Proxy (already set)

`web/vite.config.ts` proxies `^/api` to `http://localhost:8080` so you don’t need CORS during development.

---

## API Overview (planned)

The API surface for passkey flows (to be implemented; tracked in `blueprint/implementation.md`):

- POST `/authn/passkey/registration/options` → `PublicKeyCredentialCreationOptions`
- POST `/authn/passkey/registration/finish` → verifies attestation and creates account+credential
- POST `/authn/passkey/login/options` → `PublicKeyCredentialRequestOptions`
- POST `/authn/passkey/login/finish` → verifies assertion and establishes session
- POST `/tx/sign` → sign a minimal CBOR content bundle with passkey; server verifies signature
- GET `/transaction/list` → list signed transactions for the current user

Current status: only `/health` is wired; features are being built.

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

Open **two terminals** in the repo root:

**Terminal A — API**

```bash
make server
# runs: cd server && go run ./cmd/api
# API: http://localhost:8080  (health: /health)
```

**Terminal B — UI**

```bash
make web
# runs: cd web && npm run dev
# UI: http://localhost:5173  (proxying /api → :8080)
```

Visit: **[http://localhost:5173](http://localhost:5173)**

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

## Make Targets

```bash
make server   # run API
make web      # run UI (Vite dev)
make build    # build API + UI
make clean    # remove db + dist
```

---
## Blueprint

- `blueprint/goals.md`: project goals and design tenets.
- `blueprint/_decisions/`: ADRs (architecture decisions).
- `blueprint/global/` and `blueprint/features/`: NFRs/FRs + specs.
- `blueprint/_user-flows/`: canonical user flows.
- `blueprint/implementation.md`: Source of truth for implementation tasks.

Pull tasks from the implementation plan and keep “Refs” current in all artifacts to maintain traceability.

---

That’s it—happy passkey hacking!
