### Step 7 — Types: COSE, WebAuthn, Bundle (Done: 2025-08-31)

- Context

  - Define core data shapes used across registration/login/signing: COSE EC2 public key, the signing `Bundle` (per CDDL), and minimal WebAuthn request/response payload shapes. Centralizing these types reduces duplication and mismatches across handlers.

- Structure

  - Add `server/internal/types/types.go` with plain structs and doc comments; no logic beyond JSON/CBOR tags where needed.
  - Keep enums/consts simple (e.g., ceremony types `"webauthn.create"|"webauthn.get"`).

- Source to add (instructions only)

  - `server/internal/types/types.go`:
    - `type CoseEC2 struct { Kty int \tAlg int \tCrv int \tX []byte \tY []byte }` — holds COSE EC2 public key fields parsed from attestation.
    - `type Bundle struct { SenderKey CoseEC2; Nonce uint64; Message string; ValidUntil *uint64 }` — mirrors the CDDL; used in signing.
    - Minimal WebAuthn payload shapes (JSON):
      - `type RegOptions struct { RP_ID string; Origin string; UVRequired bool; Attestation string }`
      - `type RegFinish struct { ID string; RawID string; Response struct{ AttestationObject string; ClientDataJSON string } }`
      - `type LoginOptions struct { RP_ID string; Origin string; UVRequired bool }`
      - `type LoginFinish struct { ID string; RawID string; Response struct{ AuthenticatorData string; ClientDataJSON string; Signature string; UserHandle string } }`
    - Note: Binary fields are base64url strings at the API surface; server decodes to bytes before verification.

- Description files to add (instructions only)

  - `server/internal/types/types.go.desc.md`: Purpose (shared models), Key Types (CoseEC2, Bundle, WebAuthn payloads), Interactions (used by handlers and helpers), Refs.
    - Refs: requirement R-ID-KEY; requirement R-SCHEMA-LITE; requirement R-PLAT-2.

- Blueprint updates

  - Refs to include upon implementation: goal key-first-identity-cose; goal minimal-cbor-bundle; requirement R-ID-KEY; requirement R-SCHEMA-LITE; requirement R-PLAT-2.

- Verification (to run after implementation)

  - Lint/build: `cd server && go vet ./... && go build ./...` (expect exit 0).
  - Import sanity: run `rg -n "package types" server` to confirm package compiles and is discoverable.

- User verification commands (copy/paste)

  ```bash
  cd server
  go vet ./... && go build ./...
  rg -n "package types|type CoseEC2|type Bundle" internal/types/types.go
  cd -
  ```

   - Notes
     - Keep types minimal and transport‑oriented; parsing/validation logic belongs in subsequent steps (e.g., WebAuthn parsers, CBOR codec, crypto helpers).

