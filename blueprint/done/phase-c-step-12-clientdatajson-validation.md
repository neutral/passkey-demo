### Step 12 — ClientDataJSON validation (Done: 2025-08-31)

    - Context
      - Parse WebAuthn ClientDataJSON (CDJ) from browser assertions/attestations to extract `type` ("webauthn.get"/"webauthn.create"), `challenge` (base64url string), and `origin` (string). Decode the base64url `challenge` with tolerant rules and expose helpers used by later verification.

    - Structure
      - Add `server/internal/webauthn/cdj.go` with:
        - Type: `ClientData { Type string; Challenge string; Origin string }` (raw fields) and `ClientDataParsed { Type string; Challenge []byte; Origin string }`.
        - `func ParseClientDataJSON(b []byte) (ClientDataParsed, error)` — parses JSON, validates required fields, decodes challenge via `internal/encoding` base64url.
        - Optional helper: `func IsGet(c ClientDataParsed) bool` and `func IsCreate(c ClientDataParsed) bool`.

    - Source to add (instructions only)
      - `server/internal/webauthn/cdj.go`:
        - Unmarshal JSON into a temporary struct; ensure `type` and `origin` are non‑empty; keep `challenge` as string and decode to bytes using `encoding.Decode` (tolerant); return `ClientDataParsed`.
        - Do not perform origin/RP checks here (reserved for Step 14); this step ensures correct parsing and decoding only.
        - Accept only documented `type` values ("webauthn.get", "webauthn.create"); return clear error otherwise.

    - Description files to add (instructions only)
      - `server/internal/webauthn/cdj.go.desc.md`: Purpose (parse CDJ), Key Logic (JSON parse + base64url decode), Interactions (used by assertion/attestation verification), Refs.
        - Refs: requirement R-PLAT-2; decision webauthn-corrections-and-standardizations; decision encoding-and-ceremony-guardrails.

    - Unit tests (files, cases, invariants, commands)
      - Add `server/internal/webauthn/cdj_test.go`:
        - Happy path (get): valid JSON with unpadded base64url challenge and origin; decode matches expected bytes.
        - Happy path (create): valid JSON with padded base64url challenge; tolerant decode succeeds.
        - Invalid: missing fields (type/origin/challenge) → error; unknown type → error; malformed base64 → error; malformed JSON → error.
      - Command: `cd server && go test ./...` (fix‑forward loop: run, address failures, re‑run until green).

    - Verification (to run after implementation)
      - Build/tests: `cd server && go test ./...` (expect all tests pass for CDJ parser and existing suites).

    - Notes
      - Keep parsing concerns separate from policy checks (origin/RP validation in Step 14). Use `internal/encoding` for base64url to stay consistent with ADR guardrails.
