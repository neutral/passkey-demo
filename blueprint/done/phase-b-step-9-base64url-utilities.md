### Step 9 — Base64url utilities (Done: 2025-08-31)

   - Context
     - Provide safe and consistent base64url encoding/decoding utilities per ADR guardrails: strip '=' padding on encode; accept both with/without padding on decode. Used across JSON APIs for binary fields (ids, keys, signatures, CBOR bundles).

   - Structure
     - Add `server/internal/encoding/b64url.go` with two functions:
       - `func Encode(b []byte) string` — returns base64url without padding.
       - `func Decode(s string) ([]byte, error)` — tolerant: accepts with or without padding; returns error on invalid alphabet/length.
     - Keep the package small and dependency‑free (use `encoding/base64`).

   - Source to add (instructions only)
     - `server/internal/encoding/b64url.go`:
       - Implement `Encode` using `base64.RawURLEncoding.EncodeToString`.
       - Implement `Decode` trying `base64.RawURLEncoding.DecodeString` first; if it fails due to padding, fall back to `base64.URLEncoding.DecodeString` after normalizing input; on error, wrap with context.
       - Ensure empty input returns empty bytes without error.

   - Description files to add (instructions only)
     - `server/internal/encoding/b64url.go.desc.md`: Purpose (transport encoding), Key Logic (no padding on encode, tolerant decode), Interactions (used by handlers/parsers), Refs.
       - Refs: decision encoding-and-ceremony-guardrails; requirement R-PLAT-2; requirement R-ERR.

   - Blueprint updates
     - Refs to include upon implementation: decision encoding-and-ceremony-guardrails; requirement R-ERR (error mapping); requirement R-PLAT-2.

   - Unit tests (files, cases, invariants, commands)
     - Add `server/internal/encoding/b64url_test.go`:
       - Roundtrip invariants: random bytes; empty; large payload (e.g., 65KB) should roundtrip.
       - Tolerance: decode both padded and unpadded encodings of same bytes; outputs equal.
       - Invalid inputs: bad alphabet characters; malformed padding; expect errors.
       - Idempotence: `Decode(Encode(b)) == b` for diverse cases.
     - Command: `cd server && go test ./...` (fix‑forward loop: run, address failures, re‑run until green).

   - Verification (to run after implementation)
     - Build/tests: `cd server && go test ./...` (expect all tests pass).
     - Optional quick check in a tiny snippet that encodes/decodes a known value.

   - Notes
     - Keep this package isolated from higher‑level JSON concerns; callers handle struct marshalling.
     - Do not introduce external dependencies for base64.

