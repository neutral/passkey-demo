### Step 11 — Parse authenticatorData (Done: 2025-08-31)

    - Context
      - Parse WebAuthn authenticatorData (AD) bytes from assertions to extract: `rpIdHash` (32 bytes), `flags` (bitfield), and `signCount` (big‑endian uint32). This is required for login and signing verification and for enforcing UV and counter policies.

    - Structure
      - Add `server/internal/webauthn/ad.go` with:
        - Types: `AuthData { RpIDHash [32]byte; Flags byte; SignCount uint32 }` and flag masks (UP/UV/AT/ED).
        - `func ParseAuthData(b []byte) (AuthData, []byte, error)` — parses fixed header (37 bytes) and returns `AuthData` and the remaining bytes (for future use: attested credential data or extensions).
        - Helpers: `func HasUV(flags byte) bool`, `func HasUP(flags byte) bool`.

    - Source to add (instructions only)
      - `server/internal/webauthn/ad.go`:
        - Validate length ≥ 37 bytes; copy first 32 as `RpIDHash`; read flags; read `SignCount` as big‑endian uint32; return remainder slice.
        - Do not parse attested credential data or extensions here (reserved for later steps); return remainder for callers that need it.
        - Document flag meanings: Bit 0 UP, Bit 2 UV, Bit 6 AT, Bit 7 ED.

    - Description files to add (instructions only)
      - `server/internal/webauthn/ad.go.desc.md`: Purpose (parse AD header), Key Logic (length checks, big‑endian counter, flags), Interactions (used by login/signing and registration later), Refs.
        - Refs: requirement R-PLAT-2; requirement R-SEC-UV; decision webauthn-corrections-and-standardizations.

    - Unit tests (files, cases, invariants, commands)
      - Add `server/internal/webauthn/ad_test.go`:
        - Happy path: 32‑byte rpIdHash + flags with UV set + signCount=0x00000005; assert parsed fields, big‑endian value 5, remainder length 0.
        - Flags: craft flags with UP only, UV only, both; assert helpers `HasUP/HasUV` work.
        - Invalid: length < 37 returns error; ensure no panics and clear error messages.
        - Counter invariants: parsing preserves big‑endian; specific byte order test (e.g., 0x01 0x02 0x03 0x04 → 0x01020304).
      - Command: `cd server && go test ./...` (fix‑forward loop: run, address failures, re‑run until green).

    - Verification (to run after implementation)
      - Build/tests: `cd server && go test ./...` (expect all tests pass for AD parser and existing suites).

    - Notes
      - Keep this parser minimal and focused on assertion use‑cases. Attested credential data (AAGUID, credentialId, COSE key) is parsed in Step 16.
