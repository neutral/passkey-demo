### Step 17 — /authn/passkey/registration/finish handler (Done: 2025-09-01)

Scope
- Complete registration by validating clientDataJSON and attestationObject against the stored registration session, enforcing RP/Origin and UV policies, and persisting the new account and credential.

Artifacts
- Code: `server/internal/webauthn/reg_finish.go`
  - Handler `RegistrationFinishHandler(cfg, regStore, db)` at `POST /authn/passkey/registration/finish`.
  - Validations: session (TTL/single-use), CDJ type=create + challenge match + origin policy (Step 14), attestation fmt=none extraction (Step 16), rpIdHash policy (Step 14), UV required, COSE EC2 → ECDSA P‑256.
  - Persistence: `accounts` (acct_cbor canonical CBOR; acct_thumb = SHA256("ACCTK1"||acct_cbor)), `credentials` (credential_id, acct_cbor_fk, sign_count, aaguid).
- Wiring: `server/cmd/api/main.go` mounts `/authn/passkey/registration/finish` and initializes SQLite.
- Docs: `server/internal/webauthn/reg_finish.go.desc.md`, `specs/explainer-registration-finish.md`.
- Tests: `server/internal/webauthn/reg_finish_test.go` — happy path + negative cases (expired session, wrong challenge/origin, rpId mismatch, missing UV, duplicate cred, unsupported fmt, bad attestation).

Verification
- `cd server && go test ./...` passes; DB rows created for happy path; error paths map to expected statuses.

Notes
- Attestation trust-chain verification intentionally out of scope (`attestation: none`).
- Responses are minimal and do not expose sensitive details; logging to follow in Future logging item.

Refs
- Step 14 policy checks; Step 16 attestation parsing; R-FLOW-REG; decision webauthn-corrections-and-standardizations.

