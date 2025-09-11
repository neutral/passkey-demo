# Overview
Centralized repository for credential lookups backed by prepared statements. Reduces SQL duplication across handlers/services.

# Relations
- Used by transaction options and finish flows to list allowed credential IDs and load `{acct_cbor_fk, sign_count}` for a credential.
- Prepared once at startup and shared.

# Interfaces & Models
- `NewCredentials(ctx, db)` → `*CredentialsRepo` with prepared statements.
- `ListIDsByAccount(ctx, acctCBOR) ([][]byte, error)`
- `GetAccountAndCount(ctx, credID) (acctCBOR []byte, signCount int64, err error)`

# Refs
Refs: goal simple-ui-and-storage; requirement sqlite-persistence; decision data-access-repos-prepared

