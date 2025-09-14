### Step 16 — Attestation parsing (minimal) (Done: 2025-09-01)

Scope
- Parse `attestationObject` (fmt: "none") to extract authenticatorData header, AAGUID, credential ID, and COSE EC2 key for registration finish.

Artifacts
- Code: `server/internal/webauthn/att.go`
  - `ParseAttestationObject`, `ParseAttestedCredentialData`, `ParseCOSEKeyEC2`, `ExtractRegistrationData`.
  - Sentinels: `ErrAttestationCBOR`, `ErrAttestationFormat`, `ErrAttestedDataMissing`, `ErrAuthDataShort`, `ErrCredIDLength`, `ErrCOSEKeyDecode`.
- Docs: `server/internal/webauthn/att.go.desc.md`.
- Tests: `server/internal/webauthn/att_test.go` — happy path; missing AT flag; unsupported fmt; truncated authData; bad COSE CBOR.

Verification
- `cd server && go test ./...` passes; extraction returns expected fields for fmt "none" and rejects malformed inputs with meaningful errors.

Notes
- Attestation trust-chain verification is out of scope (policy: `attestation: none`).
- Extensions remainder after COSE key is not parsed at this step.

Refs
- WebAuthn spec (attestationObject, attestedCredentialData); encoding/cbor utilities; Step 14 AD parsing; `types.CoseEC2`.

