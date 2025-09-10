### Step 36 — Golden Vectors (Done: 2025-09-10)

Verification notes
- Implemented internal generator and CLI; generated and committed `specs/goldens/tx-bundle-v1.json`.
- Repeated CLI runs produce identical JSON; deterministic inputs verified.
- Added `anchors_golden_test.go`; passes locally and fails on any golden field change.
- Build/test commands used:
  - `go -C server build ./cmd/vectors && ./server/vectors -fmt json > specs/goldens/tx-bundle-v1.json`
  - `go -C server test ./internal/tx -run Golden -v`

36. **Golden Vectors**

    - Purpose

      - Produce deterministic, reviewable reference vectors for the transaction bundle (B), challenge, and tx_id that other tests and tools can rely on. Lock in canonical CBOR encoding and anchor formulas across platforms.

    - Scope

      - Add a small CLI to emit golden vectors as JSON and plaintext (hex/base64url) using a fixed, deterministic input (COSE key derived from the P‑256 base point; nonce=1; message="hello").
      - Store the emitted JSON file under `specs/goldens/` and add a unit test that compares programmatic recomputation against the committed golden file.
      - Keep outputs explicit: canonical CBOR (B) in hex and base64url; challenge in hex and base64url; tx_id in hex; plus round‑trippable inputs.

    - Source to add/modify

      - Add: `server/cmd/vectors/main.go`
        - Flags:
          - `-fmt json|text` (default: `json`). When `json`, write JSON to stdout. When `text`, print a human‑readable block (hex + base64url lines).
        - Deterministic inputs:
          - COSE EC2 key: P‑256 base point (Gx,Gy) padded to 32 bytes.
          - Bundle: `{ sender_key=k, nonce=1, message="hello" }` encoded via canonical CBOR.
        - Outputs (JSON schema below) computed via existing helpers:
          - `bundle_cbor_hex`, `bundle_cbor_b64`
          - `challenge_hex`, `challenge_b64` where `challenge = SHA-256("CHALv1"||B)`
          - `tx_id_hex` where `tx_id = SHA-256("TXIDv1"||B)`
          - `sender_key_cbor_b64`, `sender_key_cbor_hex`, `message`, `nonce` (inputs echoed)
      - Add: `server/internal/vectors/vectors.go`
        - Small package that exposes `Generate() (Vectors, error)` sharing the logic used by CLI and tests.
      - Add: `specs/goldens/tx-bundle-v1.json`
        - The committed reference JSON produced by `vectors -fmt json`.
      - Add (test): `server/internal/tx/anchors_golden_test.go`
        - Loads `specs/goldens/tx-bundle-v1.json`, recomputes via `internal/vectors.Generate()`, and asserts equality for all fields (hex/base64url).
        - Optional script: `tools/update-goldens.sh` that rebuilds and writes the JSON file.

    - Description files (create/update alongside code changes)

      - Create: `server/cmd/vectors/vectors.desc.md` — Folder overview of the vectors CLI, responsibilities, inputs/outputs, and relations to `internal/vectors`. Refs included.
      - Create: `server/cmd/vectors/main.go.desc.md` — Purpose, flags (`-out`, `-fmt`), deterministic inputs, emitted fields, error behavior. Refs included.
      - Create: `server/internal/vectors/vectors.desc.md` — Folder responsibilities, consumers (CLI and tests), determinism guarantees, and schema it produces. Refs included.
      - Create: `server/internal/vectors/vectors.go.desc.md` — Generation logic, COSE/key construction, canonical encoding, anchor formulas, and returned struct. Refs included.
      - Create/Update (docs): `specs/goldens/goldens.desc.md` — Purpose of golden vectors, update policy, how to regenerate, and how tests consume the file. Refs included.
      - Create: `server/internal/tx/anchors_golden_test.go.desc.md` — Purpose, fixtures used, validation strategy (compare against committed golden), and failure modes. Refs included.
      - If script added: `tools/update-goldens.sh.desc.md` — Script purpose, idempotency, and when to run. Refs included.

    - JSON shape (golden)

      ```json
      {
        "version": "tx-bundle-v1",
        "inputs": {
          "message": "hello",
          "nonce": 1,
          "sender_key_cbor_b64": "...",
          "sender_key_cbor_hex": "..."
        },
        "bundle": {
          "bundle_cbor_b64": "...",
          "bundle_cbor_hex": "..."
        },
        "anchors": {
          "challenge_b64": "...",
          "challenge_hex": "...",
          "tx_id_hex": "..."
        }
      }
      ```

    - Algorithms & Determinism

      - Canonical CBOR: use `internal/encoding.EncodeCanonical` exclusively for both the bundle and the sender key.
      - COSE key: derive from P‑256 base point (Gx,Gy) and left‑pad to 32 bytes; same as used in unit tests; guarantees stable sender key CBOR.
      - Anchors: `challenge = SHA-256("CHALv1" || B)` and `tx_id = SHA-256("TXIDv1" || B)`; keep prefix constants as in `internal/tx/bundle.go`.
      - Base64url: use unpadded base64url for binary → string fields (`*_b64`).

    - Verification

      - Build & run CLI:

        ```bash
        go -C server build ./cmd/vectors
        ./server/vectors -fmt json > /tmp/tx-bundle-v1.json
        ./server/vectors -fmt text | sed -n '1,120p'
        ```

      - Update goldens (if needed):

        ```bash
        ./server/vectors -fmt json > specs/goldens/tx-bundle-v1.json
        git add specs/goldens/tx-bundle-v1.json
        ```

      - Unit test:

        ```bash
        go -C server test ./internal/tx -run Golden -v
        go -C server test ./... # full pass
        ```

      - Acceptance:
        - Running `vectors` twice yields identical outputs.
        - `anchors_golden_test.go` passes and fails if any field in `specs/goldens/tx-bundle-v1.json` is modified.
        - Canonical CBOR from the golden equals the programmatic re‑encode of the bundle object (round‑trip confirmed in tests).

    - Documentation & Notes

      - Add a short README snippet under `specs/goldens/` explaining the purpose and update policy.
      - Goldens are tightly coupled to the current canonical encoding and anchor prefixes; changes require an ADR or explicit step update.
      - Keep goldens small and focused; avoid embedding entire transactions—only bundle B and anchors.

    - Refs

      - Refs: goal server-derived-challenge-and-txid; requirement R-SCHEMA-LITE; requirement R-FLOW-SIGN; decision encoding-and-ceremony-guardrails

