### Step 10 — CBOR canonical codec (Done: 2025-08-31)

    - Context
      - Provide deterministic, canonical CBOR encoding/decoding utilities used to bind signing challenges to exact bytes. Canonical encoding ensures the same content always serializes the same way (stable hashes/signatures).

    - Structure
      - Add `server/internal/encoding/cbor.go` with two helpers:
        - `func EncodeCanonical(v any) ([]byte, error)` — uses fxamacker/cbor in canonical mode.
        - `func DecodeCanonical(data []byte, v any) error` — decodes into a provided target.
      - Keep this package next to b64url utilities for coherence.

    - Source to add (instructions only)
      - `server/internal/encoding/cbor.go`:
        - Configure fxamacker encoder/decoder with canonical options (CTAP2/Deterministic where applicable): map key sorting, no indefinite length, shortest integer/length forms.
        - Expose thin wrappers over a configured `EncOptions/DecOptions` instance; avoid global mutable state.
        - Ensure zero values behave (nil/empty) and errors are wrapped with context.

    - Description files to add (instructions only)
      - `server/internal/encoding/cbor.go.desc.md`: Purpose (canonical CBOR), Key Logic (options/invariants), Interactions (used by bundle hashing/challenge derivation), Refs.
        - Refs: requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; requirement R-PLAT-2.

    - Unit tests (files, cases, invariants, commands)
      - Add `server/internal/encoding/cbor_test.go`:
        - Determinism: encode the same Go map with different key insertion orders; bytes must be identical.
        - Ordering: verify map keys are CBOR‑sorted (e.g., ints before texts by canonical rules); compare to a known hex string (golden) for a small fixture.
        - Roundtrip: struct and map roundtrip equals (deep‑equal) for representative types (ints, strings, byte slices, nested structs, optional field nil vs omitted behavior if applicable).
        - Error cases: malformed CBOR bytes return decode errors; excessively large indefinite forms (if attempted) are rejected by options.
      - Command: `cd server && go test ./...` (fix‑forward loop: run, address failures, re‑run until green).

    - Verification (to run after implementation)
      - Build/tests: `cd server && go test ./...` (expect all tests pass), including golden determinism test.
      - Optional: small CLI snippet to print hex of a canonical encoding for a known map, compare across runs.

    - Notes
      - This codec must not vary across runs or platforms; keep options explicit and rely on fxamacker’s canonical modes.
      - Consider adding a non‑developer “CBOR usage explainer” under the schema requirement if needed for stakeholders.
