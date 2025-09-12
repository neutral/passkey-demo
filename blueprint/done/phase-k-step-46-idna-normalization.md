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
- Unit tests will cover Unicode/A-label pairs and invalid hostcases in a follow-up; current short test suite remains green.

## Verification
- Build + short tests:
```bash
cd server && go build ./... && go test -short ./...
```

## Acceptance criteria
- Policy checks accept and compare normalized A-label hostnames consistently.

## Refs
- Refs: requirement R-PLAT-2

