# CBOR/COSE Interop and Decoding Fallbacks

## Status
Approved — 2025-09-10

## Context
- Symptom observed during Manual E2E (Step 37): `POST /tx/signing/options` returned 401 "sender_key mismatch" even with a valid session and CORS.
- Root cause: The frontend (JS, `cbor-x`) produced a nested COSE map (at bundle key `0`) whose shape did not always bind to the backend (Go, fxamacker/cbor) typed struct decoder. In these cases, the strict typed decode left `kty/alg/crv/x/y` as zero values even though a generic decode revealed a valid EC2 key (kty=2, alg=-7, crv=1, 32‑byte X/Y) matching the UI.
- Contributing factors:
  - Differences in map key typing and byte-string wrappers (e.g., `cbor.Tag`) across encoders.
  - Typed struct decode in Go requires exact key shapes; it does not populate fields if keys/types differ, and may not error.
  - Comparing raw CBOR bytes for identity is brittle across language libraries and encoder variants even when logical fields match.

## Decision
We standardize identity binding and decoding robustness at the transaction prep/signing boundary:

1) Identity comparison uses logical COSE fields only.
   - Compare `kty`, `alg`, `crv`, and byte arrays `x`, `y` for equality.
   - Do not compare raw CBOR bytes for identity binding.

2) Robust nested‑COSE decoding on the server.
   - Decode the bundle canonically to a typed struct; if the nested `sender_key` fields are all zero, apply a fallback:
     - Unmarshal the top‑level bundle as `map[int64]cbor.RawMessage`, else `map[uint64]`/`map[any]`, and extract key `0`.
     - Attempt typed unmarshal of that raw value into a struct keyed by COSE integers (`1, 3, -1, -2, -3`).
     - If still empty, unmarshal to `map[any]any`, coerce keys to ints, and unwrap `cbor.Tag` to extract `[]byte` for X/Y.
   - After populating fields, re‑encode the logical bundle canonically to form `B` used for anchors.

3) Canonical embedding on the client.
   - Fetch `/me/account_key`, decode `acct_cbor_b64`, normalize to a `Map<number, any>` with numeric COSE keys (`1, 3, -1, -2, -3`).
   - Build the bundle as `new Map([[0, senderCoseObj], [1, nonce], [2, message]])` and encode canonically.
   - Avoid reconstructing the COSE object from typed JSON; embed the server’s canonical object to eliminate structural drift.

## Consequences
- Increased robustness across language/library boundaries; fewer false mismatches.
- Canonical CBOR (`B`) remains the basis for anchors (`challenge`, `tx_id`); identity is based on logical fields.
- The server carries a small amount of fallback logic to tolerate benign encoder differences while only accepting the standard COSE keys.
- Debug/inspection endpoints are not required for normal operation and are removed from the minimal surface area.

## Alternatives Considered
- Raw CBOR byte equality for identity: rejected as brittle across encoders and map canonicalization details.
- Typed struct decode only (no fallback): rejected; breaks when encoders vary in key typing/tagging.
- Forcing a JSON canonical representation across the wire: avoided to keep CBOR/COSE end‑to‑end and minimize surface area.
- Relying on library knobs only (decoder options): may reduce, but not eliminate, cross‑encoder shape variance; fallback still recommended.
- Normalizing on the client without embedding server canonical COSE: still risks drift; embedding aligns stacks reliably.

## References
- Server implementation: `server/internal/tx/bundle.go`
- Client implementation: `web/src/pages/Dashboard.tsx`, `web/src/lib/bundle.ts`, `web/src/lib/cbor.ts`

## Refs
Refs: goal server-derived-challenge-and-txid; goal minimal-cbor-bundle; requirement R-FLOW-SIGN; requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations; spec bundle-shape-and-client-production-explainer; spec account-binding-explainer

