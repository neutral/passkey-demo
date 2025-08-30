# R-PLAT-3 — SQLite Persistence (on-disk)

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Use SQLite as the persistence layer to store accounts, credentials, sessions, and transactions. Binary fields stored as BLOBs.

## Depends On
- R-PLAT-2 (backend service)

## Scope
- In-scope: schema migration/creation, CRUD for flows; using base64url for binary in the API surface.
- Out-of-scope: replication, HA, external DB engines.

## Acceptance Criteria
- DB schema created on startup if not present; tables support all demo flows.
 - Sessions may be managed in-memory or persisted via the `sessions` table; both are acceptable for the demo.

## Schema (SQL)
```sql
CREATE TABLE IF NOT EXISTS accounts (
  acct_cbor    BLOB PRIMARY KEY,
  acct_thumb   BLOB NOT NULL,
  created_at   INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS credentials (
  credential_id BLOB PRIMARY KEY,
  acct_cbor_fk  BLOB NOT NULL,
  sign_count    INTEGER NOT NULL,
  aaguid        BLOB,
  created_at    INTEGER NOT NULL,
  FOREIGN KEY(acct_cbor_fk) REFERENCES accounts(acct_cbor) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions (
  session_id  TEXT PRIMARY KEY,
  acct_cbor   BLOB NOT NULL,
  expires_at  INTEGER NOT NULL,
  created_at  INTEGER NOT NULL,
  FOREIGN KEY(acct_cbor) REFERENCES accounts(acct_cbor) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
  tx_id       BLOB PRIMARY KEY,
  acct_cbor   BLOB NOT NULL,
  nonce       INTEGER NOT NULL,
  message     TEXT NOT NULL,
  bundle_cbor BLOB NOT NULL,
  auth_data   BLOB NOT NULL,
  client_data BLOB NOT NULL,
  signature   BLOB NOT NULL,
  created_at  INTEGER NOT NULL,
  FOREIGN KEY(acct_cbor) REFERENCES accounts(acct_cbor) ON DELETE CASCADE
);
```

## Flows
- Used by all flows for persistence.

## Interfaces
- Go `database/sql` + SQLite driver.

## Risks
- Binary encoding/decoding mismatches; locking/contention in dev; schema drift.

Refs: goal simple-ui-and-storage; decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; requirement R-PLAT-3
