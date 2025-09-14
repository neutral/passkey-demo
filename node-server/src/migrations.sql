-- SQLite schema for the Passkey Demo (Node parity with Go backend)

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

