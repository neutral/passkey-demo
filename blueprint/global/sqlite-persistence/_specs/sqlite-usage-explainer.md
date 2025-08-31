# R-PLAT-3 — SQLite Usage Explainer (Non‑Developer)

## Purpose & Audience
- Explain how SQLite is used in this project in clear, practical terms for technical collaborators who are not writing code.
- Describe what data we store, why SQLite was chosen, operational basics, and limits.

## What SQLite Is
- A lightweight, file‑based database engine embedded in the application (no separate database server to install or manage).
- Data lives in a single file on disk; the app reads/writes that file directly.

## Why We Use It Here
- Demo simplicity: one binary + one data file is easy to run locally.
- Adequate for our needs: low traffic, single process, small dataset.
- Avoids extra infrastructure (no cloud DB, no admin overhead) while still giving real transactions and constraints.

## What We Store
- Accounts: the passkey’s public key (binary) as the account identity, plus a short “thumbprint” for display.
- Credentials: the passkey credential identifier and usage counter linked to an account.
- Sessions: the login session token for the browser (optional; demo can also keep these in memory).
- Transactions: the signed message records (nonce, message, binary materials) for the Dashboard list.

## How It Works in This App
- One file: created automatically at first run (default `server/demo.db`). It’s ignored by git and safe to delete when you want a fresh start.
- Integrity: the schema enforces relationships (e.g., deleting an account removes its credentials/transactions) and basic constraints.
- Binary data: keys, signatures, and other binary fields are stored as BLOBs in the DB; the API encodes these as base64url strings in JSON responses/requests.
- Concurrency: tuned for a single running app process on a laptop. Write conflicts are rare at demo scale; we set a short retry timeout.

## Performance & Safety Settings
- Write‑Ahead Logging (WAL): improves reliability and read/write behavior for a local app.
- Foreign keys: enabled to keep linked rows consistent (e.g., credentials must belong to an existing account).
- Timeouts: a short “busy timeout” prevents errors if two operations collide briefly.

## Security & Privacy
- No separate DB credentials: the app opens the local file directly.
- Sensible logging: the app logs non‑sensitive identifiers (thumbprints, IDs), not raw binary contents.
- Local protection: rely on your OS user permissions for the DB file. The demo does not add at‑rest encryption.

## Operating It Day‑to‑Day
- Start/stop: just run the server; the DB file appears automatically if missing.
- Reset: stop the server and delete `server/demo.db` to reset all data.
- Backup: stop the server and copy the DB file elsewhere; restore by copying it back (same app version recommended).

## Limitations & When to Upgrade
- Not for heavy load or multi‑server setups. There’s one writer at a time, which is fine for a demo.
- No built‑in high availability. For production or team scale, move to a client/server DB (e.g., Postgres) and keep the same tables/constraints.

## Errors & Observability
- The app maps storage errors to clear HTTP responses (e.g., conflict, bad request), and keeps logs minimal but useful.
- If you see database “busy” messages, it usually means two operations briefly overlapped; the app retries quickly.

## Glossary
- BLOB: raw binary data stored as‑is in the database.
- WAL: a journaling mode that writes changes to a log first, improving reliability on local machines.

## Refs
- Refs: requirement R-PLAT-3; goal simple-ui-and-storage; decision encoding-and-ceremony-guardrails; requirement R-PLAT-2
