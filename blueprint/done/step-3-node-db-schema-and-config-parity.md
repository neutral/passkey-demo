### Step 3 — Node: DB schema and config parity

Scope

- Implement SQLite schema in Node identical to the Go backend: same tables, columns, types, and FK constraints.
- Keep PRAGMAs aligned (WAL, synchronous=NORMAL, foreign_keys=ON) and run idempotent schema creation at startup.

Source to add/modify

- Modify `node-server/src/migrations.sql` — add the canonical schema (CREATE TABLE IF NOT EXISTS):
  - `accounts(acct_cbor BLOB PRIMARY KEY, acct_thumb BLOB NOT NULL, created_at INTEGER NOT NULL)`
  - `credentials(credential_id BLOB PRIMARY KEY, acct_cbor_fk BLOB NOT NULL, sign_count INTEGER NOT NULL, aaguid BLOB, created_at INTEGER NOT NULL, FOREIGN KEY(acct_cbor_fk) REFERENCES accounts(acct_cbor) ON DELETE CASCADE)`
  - `sessions(session_id TEXT PRIMARY KEY, acct_cbor BLOB NOT NULL, expires_at INTEGER NOT NULL, created_at INTEGER NOT NULL, FOREIGN KEY(acct_cbor) REFERENCES accounts(acct_cbor) ON DELETE CASCADE)`
  - `transactions(tx_id BLOB PRIMARY KEY, acct_cbor BLOB NOT NULL, nonce INTEGER NOT NULL, message TEXT NOT NULL, bundle_cbor BLOB NOT NULL, auth_data BLOB NOT NULL, client_data BLOB NOT NULL, signature BLOB NOT NULL, created_at INTEGER NOT NULL, FOREIGN KEY(acct_cbor) REFERENCES accounts(acct_cbor) ON DELETE CASCADE)`
- Modify `node-server/src/db.js` — add `applyMigrations(db, sqlPath)` to read and execute `migrations.sql` once (idempotent). Keep PRAGMAs unchanged.
- Modify `node-server/src/server.js` — call `applyMigrations(db, pathToSql)` on startup after opening the DB; log `db_init` with `journal_mode` and `foreign_keys`.

Description files

- Update `node-server/src/db.js.desc.md` — describe `applyMigrations` and idempotence; reiterate PRAGMAs.
- Add `node-server/src/migrations.sql.desc.md` — schema overview, invariants, and parity note with Go.
- Update `node-server/src/server.js.desc.md` — note calling `applyMigrations` during boot.

Blueprint updates (requirements/specs/ADRs/user-flows to add/update)

- Requirements (NFR): update `blueprint/global/sqlite-persistence/requirement.md` Acceptance Criteria to state Node boots create schema and enforce FKs.
- Specs (NFR): update `blueprint/global/sqlite-persistence/_specs/spec.md` Interfaces to include Node (`better-sqlite3`), and `migrations.sql` execution on startup; add idempotence note.
- ADRs: none new; ensure updated specs reference `encoding-and-ceremony-guardrails` and `R-PLAT-3`.
- Refs maintenance: keep `Refs:` current across updated artifacts.

Request/response shape

- Not applicable (DB only).

Algorithm

- On startup: open DB (PRAGMAs), execute `migrations.sql` in a transaction, verify `PRAGMA foreign_keys=ON`, and log `db_init` with key fields.

Database interactions

- Create tables only; no CRUD in this step. FKs use `ON DELETE CASCADE` as listed; types are `INTEGER`, `TEXT`, `BLOB` exactly.

Policies & limits

- Single writer connection; PRAGMAs enforced (`WAL`, `synchronous=NORMAL`, `foreign_keys=ON`).
- Migrations are safe to run on every boot.

Sequencing

- Requires Step 2 scaffold; precedes auth/session/WebAuthn steps which rely on these tables.

Tests

- Add `node-server/test/db-schema.test.js`:
  - Use a temporary on-disk DB path; call `applyMigrations` directly.
  - Assert: `PRAGMA foreign_keys` equals 1; `PRAGMA journal_mode` reports `wal`.
  - Assert `sqlite_master` includes `accounts`, `credentials`, `sessions`, `transactions`.
  - Assert FK list for `credentials` includes `acct_cbor_fk → accounts(acct_cbor)`.
- Negative case: corrupt/absent SQL path → `applyMigrations` throws; map a readable error.
- Commands: `node --test node-server/test/db-schema.test.js`.

Verification

- Manual: start server, then inspect DB using a Node one-liner with `better-sqlite3` to print `foreign_keys` and table names; stop server.
- User verification commands

```bash
npm -C node-server i
RP_ID=localhost ORIGIN=http://localhost:5173 PORT=8080 DB_PATH=server/demo-node.db \
  node node-server/src/server.js &
API_PID=$!
node --input-type=module -e "import Database from 'better-sqlite3'; const db=new Database('server/demo-node.db'); console.log('foreign_keys=', db.pragma('foreign_keys', { simple: true })); console.log('tables=', db.prepare('select name from sqlite_master where type=\'table\' order by name').all().map(r=>r.name)); db.close();"
kill $API_PID || true
```

Acceptance criteria

- `migrations.sql` contains the full schema matching Go exactly.
- `server.js` applies migrations at startup; `db.js` exports `applyMigrations`.
- PRAGMAs enforced; tests planned for PRAGMA/table checks and a negative case.

Notes

- Keep migrations simple (single SQL file executed transactionally) for the demo; no versioning needed.
- Strictly mirror Go column names/types to avoid drift.

Refs: requirement R-PLAT-3; decision encoding-and-ceremony-guardrails

---

