# Step 46 — IDNA (punycode) normalization (Done: 2025-09-12)

## Scope
- Normalize RP ID and Origin hostnames to A-label (punycode) with `golang.org/x/net/idna` to ensure consistent comparison and hashing.

## Source to add/modify
- `server/internal/webauthn/policy.go`:
  - IDNA ToASCII in `validateAndNormalizeRpID` and `parseAndNormalizeOrigin` (host portion).
  - Keep scheme (`https` only; `http` only for `localhost` in dev) and port logic unchanged.
- `server/internal/config/config.go`:
  - Normalize `RP_ID` and `RP_ID_ALLOWLIST` entries to A-label at load; origin allowlist remains URLs (host normalization occurs in policy).

## Description files
- `server/internal/webauthn/policy.go.desc.md`: updated Normalization & Policy to state IDNA ToASCII on RP ID and Origin host.
- `server/server.desc.md`: policy notes align with normalization; observability unchanged.

## Request/response shape
- N/A.

## Algorithm
- RP ID: lower → strip trailing dot → IDNA ToASCII → reject ip (except `localhost`) → SHA-256 → compare vs ad.RpIDHash.
- Origin: parse URL → host lower → strip trailing dot → IDNA ToASCII → compare host/scheme/port vs expected + allowlist.

## Database interactions
- None.

## Policies & limits
- Reject invalid hostnames failing IDNA conversion as 400-class policy errors (`ErrOriginMalformed`/`ErrRpIdInvalid`).
- No wildcards; no IPs for RP (except `localhost`).

## Sequencing
- Purely server-side change.

## Tests
- Added unit tests for Unicode/A-label pairs:
  - RP ID: `bücher.ch` ⇄ `xn--bcher-kva.ch`, `café.fr` ⇄ `xn--caf-dma.fr`.
  - Origin allowlist/candidate cross-matching (unicode vs ascii).
- IPv6 literal exact allow remains supported (no IDNA on IPs).

## Verification
- Build + short tests:
```bash
cd server && go build ./... && go test -short ./...
```

## Acceptance criteria
- Policy checks accept and compare normalized A-label hostnames consistently.

## Refs
- Refs: requirement R-PLAT-2

## Appendix — Additional notes
- Config: RP IDs and RP allowlist are normalized to A-label at load; invalid entries fail fast.
- Policy: Origin hostnames normalized to A-label at request time; IPv6 literals bypass IDNA.
- Specs/Docs: Updated server/config and policy descriptions, and platform spec to document normalization points and behavior.

