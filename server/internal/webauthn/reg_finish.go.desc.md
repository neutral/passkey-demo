# Purpose
Complete registration by validating clientDataJSON and attestationObject against a stored registration session, enforcing RP/Origin and UV policies, and persisting the new account and credential.

# API
- HTTP: `POST /authn/passkey/registration/finish` → 201 `{ account_thumb_hex, credential_id_b64 }` on success.
- Handler: `RegistrationFinishHandler(cfg, regStore, db)`.

# Steps
1. Load `reg_session_id`; reject if expired/missing; single-use.
2. Parse `clientDataJSON`; require `type=create`; match challenge and check origin (Step 14 policy).
3. Extract `authenticatorData` and attested credential data from `attestationObject` (`fmt: none`): AAGUID, credential ID, COSE EC2.
4. Check `rpIdHash` (Step 14), require UV flag.
5. Validate COSE EC2 → ECDSA P-256 public key.
6. Persist rows: `accounts` (acct_cbor, acct_thumb), `credentials` (credential_id, acct_cbor_fk, sign_count, aaguid).
7. Return 201 with identifiers.

# Errors
- 401: session missing/expired; challenge mismatch.
- 403: origin/RP policy failures (mapped via `MapPolicyError`).
- 400: malformed JSON/base64/CBOR; unsupported attestation format; invalid public key.
- 409: duplicate credential id.
- 500: storage/internal errors.

# Notes
- Account thumb: `SHA256("ACCTK1" || acct_cbor)`; client receives hex.
- Dev localhost origin allowed when configured origin is `http://localhost:*`.
- Single-use semantics: session deleted after terminal outcome.

# Refs
Refs: Step 14 policy checks; Step 16 attestation parsing; specs/explainer-registration-finish.md; storage schema.

