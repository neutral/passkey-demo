# R-ID-KEY — Passkey‑First Identity (COSE public key)

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: fr

## Description
- Account identity equals the canonical CBOR bytes of the passkey’s COSE EC2 public key established at registration. No username is required; all subsequent operations (login, signing) resolve identity via credential→account mapping.

## Depends On
- R-PLAT-2 (single Go service backend)
- R-PLAT-3 (SQLite persistence)
- R-FLOW-REG, R-FLOW-LOGIN, R-FLOW-SIGN
- R-SEC-UV (UV required)

## Scope
- In-scope: storing canonical CBOR of COSE EC2 public key as primary key; deriving and exposing an `acct_thumb` (SHA-256("ACCTK1" || acct_cbor)) for display/logging; mapping credential_id→account; using account key for signature verification.
- Out-of-scope: multiple credentials per account (out of demo scope, but extensible later), federation, aliases/usernames.

## Acceptance Criteria
- On registration finish, server extracts COSE EC2 public key and stores canonical CBOR in `accounts.acct_cbor`; computes and stores `acct_thumb`.
- Credentials reference accounts via `acct_cbor_fk`; server does not duplicate the public key in `credentials`.
- Login and signing verifications use the account’s COSE key; flows complete end-to-end.
- Authenticated APIs (Node `/me/account_key`, Go parity) surface the canonical account key (`acct_cbor_b64`, sender key coordinates, thumb hex, created_at) only when a valid `sid` session cookie is present; missing/expired sessions yield 401 envelopes.

## Flows
- Registration: blueprint/_user-flows/registration.md
- Login: blueprint/_user-flows/login.md
- Transaction signing: blueprint/_user-flows/transaction-signing.md

## Interfaces
- Registration finish returns/establishes identity via stored account key.
- Downstream endpoints use the mapped account (credential→acct_cbor).

## Risks
- Incorrect CBOR canonicalization; COSE parsing errors; key duplication bugs; mismatched mapping from credential to account.

Refs: goal key-first-identity-cose; goal server-derived-challenge-and-txid; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-ID-KEY
