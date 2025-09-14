# Library Adoption Evaluation — WebAuthn/Passkeys (Go + JS)

## Purpose

Provide an in‑depth review of adopting third‑party WebAuthn/passkey libraries on the Go (server) and JS (browser/server) sides. Compare with current custom implementation, list what code would be replaced, how to configure the libraries, pros/cons in production, and impacts on the custom CBOR transaction signing flow (nonce‑anchored bundles).

## TL;DR

- Yes: we can keep signing canonical CBOR transaction bundles with nonces while using off‑the‑shelf WebAuthn libraries. Libraries validate assertions over whatever challenge the RP provides, so content‑bound challenges remain compatible (recommend mixing in server randomness).
- Most value: adopt a mature verification library for ceremony parsing/verification (Go: `go-webauthn`; or a Node microservice using `@simplewebauthn/server`) while keeping our transaction layer and repositories.
- Browser DX: `@github/webauthn-json` and/or `@simplewebauthn/browser` can simplify binary handling and enable Conditional UI; they don’t constrain our challenge semantics.

## Current Architecture (What We Have)

- Custom server verification stack under `server/internal/webauthn/`:
  - Policy: `policy.go` (RP ID hash, origin, IDNA)
  - Parsing: `cdj.go`, `ad.go`, `att.go` (clientDataJSON, authenticatorData, attestation)
  - Signature: `sig.go` (ES256, strict DER, low‑S enforcement with high‑S normalization fallback)
  - Flows: `reg_options.go`, `reg_finish.go`, `login_options.go`, `login_finish.go`
- Transaction signing: `server/internal/tx/*` derives canonical CBOR bytes B, anchors (`challenge = SHA256("CHALv1"||B)`, `tx_id = SHA256("TXIDv1"||B)`), and validates allowlist/nonce/signature.
- Frontend: `web/src/lib/webauthn.ts`/`encoding.ts` convert options and responses to/from base64url and `PublicKeyCredential` formats.

## Key Requirement: Content‑Bound Signatures

- Libraries do not limit challenge semantics—RP defines challenge bytes. We can keep “content‑bound challenges” by deriving the challenge from canonical CBOR B, and (per spec) add high‑entropy randomness (recommended) before issuing options.
- Verification libraries only validate the assertion against the challenge, RP ID hash, origin policy, and signCount; they don’t inspect our CBOR bundle.

## Candidate Libraries (Server‑Side)

### Go: `github.com/go-webauthn/webauthn`

- What it provides
  - Registration/authentication ceremonies: build options and verify responses (attestation/assertion) with spec compliance.
  - SignCount handling, attestation format validation; common algorithms (ES256, etc.).
- What it replaces
  - Most of `server/internal/webauthn/*`: parsing (CDJ/AD/attestation), signature verification, ceremony finish logic, challenge/session scaffolding.
  - Keeps: our DB schema/repos, session cookie handling, rate‑limiting, error envelope; transaction module `server/internal/tx/*` unchanged.
- Configuration
  - RP ID/Name/Origins; User entity (we can map to our account thumb or generated user ID), UV/residentKey preferences, attestation preference. You keep generating/saving challenges; library session data must be stored server‑side (we already do).
- Pros
  - Well‑maintained Go ecosystem option; reduces parser/crypto maintenance; easier to add attestation policy and metadata later.
  - Conformance coverage and community usage.
- Cons
  - API expectations (user/credential models) may require adapters around our key‑first identity (account = COSE key); still doable.
  - Less direct control over some error/telemetry surfaces (can wrap).

### Go (emerging): `passkey-go` (pure Go, assertion focus)

- What it provides
  - Lower‑level CBOR/COSE/AD parsing and `VerifyAssertion(...)` high‑level call.
- Use case
  - If we want a lighter helper while keeping our flows/policy intact. Could replace portions of `sig.go`, `cdj.go`, `ad.go`.
- Risk
  - Newer project; smaller user base; vet maturity before production.

### Node: `@simplewebauthn/server` (via microservice)

- What it provides
  - `generateRegistrationOptions`, `verifyRegistrationResponse`, `generateAuthenticationOptions`, `verifyAuthenticationResponse`; MDS integration; FIDO conformance support.
- Integration pattern
  - Run a small Node microservice responsible for WebAuthn challenge issuance + response verification. Our Go API calls it via HTTP for finish phases. Keep DB and transactions in Go.
- Pros
  - Highly adopted; thorough docs; MDS and conformance feature set; fast iteration cadence.
- Cons
  - Polyglot complexity; cross‑service latency; shared session/challenge storage coordination; operational footprint.

### Node: `fido2-lib`

- What it provides
  - Attestation/Assertion options + result verification; supports many attestation formats; extensible; used by webauthn.io.
- Integration pattern
  - Similar to SimpleWebAuthn microservice; we’d wrap its API.
- Pros/Cons
  - Mature and flexible; fewer batteries‑included conveniences than SimpleWebAuthn; still adds polyglot and ops overhead if used as a sidecar service.

## Candidate Libraries (Browser‑Side)

### `@github/webauthn-json`

- Role
  - Thin wrapper to convert between WebAuthn binary fields and JSON; simplifies base64url/ArrayBuffer handling.
- What it replaces
  - `web/src/lib/webauthn.ts` and parts of `encoding.ts` (option/response transforms). Our app logic (fetches, UI) remains.
- Pros
  - Less binary glue, fewer footguns; tiny surface.
- Cons
  - Very focused—doesn’t add business logic like Conditional UI or UX helpers.

### `@simplewebauthn/browser`

- Role
  - Helpers for invoking `create()`/`get()`, Conditional UI examples, typing.
- What it replaces
  - Similar to `webauthn-json`; could also centralize UX patterns for passkeys.
- Pros/Cons
  - Good companion to the server package; small dependency; may still need custom transforms depending on our API shapes.

## What Changes in Our Codebase

### Option A — Go‑native adoption (recommended)

- Replace (server):
  - `server/internal/webauthn/*` handlers and parsers → wrap `go-webauthn` for options/finish.
  - Retain our session middleware, error envelope, rate limiting, repos, and all `tx/*` modules.
- Keep (server):
  - Transaction signing remains exactly as is; update challenge derivation to include randomness and store `rand` in `TxSession`.
  - Encoding/CBOR packages for bundles continue to be used.
- Update (web):
  - Optionally replace `web/src/lib/webauthn.ts` with `webauthn-json` and adjust pages to new helpers.

### Option B — Node microservice for WebAuthn

- Add a Node service exposing endpoints:
  - POST `/reg/options`, POST `/reg/verify`, POST `/authn/options`, POST `/authn/verify` using `@simplewebauthn/server` (or `fido2-lib`).
- Replace (server):
  - Our Go verification code in finish handlers → HTTP calls to microservice; map results into our error envelope and telemetry.
- Keep (server):
  - All repos, sessions, tx layer, policy allowlists; we still derive challenges (or let the service do it with a custom hook) and store them in our DB/TTL.
- Trade‑offs: ops complexity, but strong feature velocity and conformance story.

### Option C — Browser‑only libs

- Replace (web):
  - `web/src/lib/webauthn.ts` → `webauthn-json`/`@simplewebauthn/browser` wrappers.
- Keep (server):
  - Entire server as is; no ceremony logic changes.

## Configuration Examples (Conceptual)

### go-webauthn

- Construct RP:
  - `RPID = cfg.RP_ID`, `RPOrigins = cfg.OriginAllowlist ∪ {cfg.Origin}`; `RPDisplayName = "Passkey Demo"`.
- Registration options:
  - `residentKey = required`, `userVerification = required`, `attestation = none` (or policy‑driven); generate challenge (we can pass our own bytes).
- Finish registration:
  - Supply saved challenge and request JSON; library returns validated credential (COSE pubkey, AAGUID, counter). Persist via our repos.
- Authentication options:
  - `userVerification = required`, `allowCredentials` from DB for tx flow; challenge derived from bundle with randomness.
- Finish authentication:
  - Provide expected challenge and origins; check `verified` and `newCounter` and do our signCount policy.

### @simplewebauthn/server

- `generateRegistrationOptions({ rpID, rpName, attestation: none, authenticatorSelection: { residentKey: required, userVerification: required } })`.
- `verifyRegistrationResponse({ expectedRPID, expectedOrigin, expectedChallenge, ... })`.
- `generateAuthenticationOptions({ rpID, allowCredentials, userVerification: required })`.
- `verifyAuthenticationResponse({ expectedRPID, expectedOrigin, expectedChallenge, ... })`.
- MDS: configure `MetadataService` and enable in verification for attestation policy.

### webauthn-json

- Import and use `create()`/`get()` wrappers that handle ArrayBuffer/JSON transforms.
- Keep our API wire format (base64url strings) or switch to library defaults and update server accordingly.

## Pros vs Cons (Production)

### Pros of adopting libraries

- Security & correctness
  - Reduce risk in CBOR/COSE parsing, DER signature handling, edge‑case policy logic; track spec evolution.
  - Easier attestation format coverage and MDS integration; some libs are FIDO conformance tested.
- Velocity & maintenance
  - Faster adoption of new platform quirks, features (Conditional UI, passkey device types), and bug fixes maintained by community.
- Interop & testing
  - Reuse community test vectors and examples; lower ongoing maintenance burden.

### Cons / Risks

- Abstraction fit
  - Model mismatches (e.g., library expects a conventional user model) require adapters around our key‑first identity and repository layout.
- Control & observability
  - Error taxonomies and telemetry fields may be less granular; need wrapping for our envelope and logs.
- Dependency surface
  - Operational and supply‑chain considerations; need to track security advisories and API changes.
- Polyglot (if Node service)
  - More services to run/monitor; extra latency; shared session/challenge coherence.

## Impact on CBOR Transaction Signing

- Remains fully supported. The library verifies that `sig = ES256(ad || SHA256(cdj))` matches our challenge bytes; the RP defines the challenge.
- Recommendation update: change `challenge = SHA256("CHALv1"||B)` to `challenge = SHA256("CHALv1"||B||rand128||sessionID)` and store `rand128` in `TxSession`.
- `tx_id = SHA256("TXIDv1"||B)` stays deterministic for idempotence. All `tx/*` logic and repos remain.

## Migration Plan (Option A — Go‑native)

1. Spike: Wrap go‑webauthn in an internal `verifier` package exposing `BeginRegistration`, `FinishRegistration`, `BeginAuthentication`, `FinishAuthentication`.
2. Adapter: Map our account thumb (SHA256("ACCTK1"||acctCBOR)) to library’s user ID; implement credential store bridges using our repos.
3. Handlers: Replace internals of `*_options.go` and `*_finish.go` to call the adapter; preserve our error envelopes and logs.
4. Tx flow: Update challenge derivation to include randomness; keep allowlist logic.
5. Tests: Port existing tests to target adapter; add happy‑path and negative cases using library’s vectors.
6. Docs: Update `.desc.md` and blueprint refs; note library config in README.

## Migration Plan (Option B — Node microservice)

1. Build a small Node service exposing the four endpoints using `@simplewebauthn/server`.
2. Shared state: choose where challenges live (Node service or our Go TTL stores). Keep a single source of truth.
3. Go API: Replace finish handlers to call the service; map `verified/newCounter` to our signCount rules.
4. Tx: Ensure content‑bound challenge computation and storage remain in our Go app or are delegated with a hook.
5. Observability: Propagate correlation IDs between services; centralize metrics.

## Recommendation

- For this codebase and goals, Option A (Go‑native `go-webauthn`) offers the best risk‑reduction/complexity trade‑off while preserving our current shape and performance model. Option C (browser wrappers) is a safe incremental win.
- Option B (Node microservice) is viable for teams standardized on SimpleWebAuthn and comfortable with polyglot ops.

## References

- go‑webauthn: https://github.com/go-webauthn/webauthn
- @simplewebauthn/server: https://simplewebauthn.dev/docs/packages/server
- FIDO conformance (SimpleWebAuthn): https://simplewebauthn.dev/docs/advanced/fido-conformance
- fido2-lib: https://github.com/webauthn-open-source/fido2-lib and https://webauthn-open-source.github.io/fido2-lib/
- webauthn-json: https://github.com/github/webauthn-json
- WebAuthn L3 spec (challenge entropy): https://www.w3.org/TR/webauthn-3/

## Refs

Refs: goal passkey-registration-login-uv; goal transaction-content-signing; goal minimal-cbor-bundle; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations
