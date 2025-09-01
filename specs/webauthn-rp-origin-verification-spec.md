# WebAuthn RP ID and Origin Verification — Technical Spec

Date: 2025-09-01
Status: Draft (analyze-first)
Owners: passkey-demo maintainers

Scope
- Define server-side verification rules and algorithms for:
  - RP ID hash check: `SHA-256(rpId)` vs `authenticatorData.rpIdHash`.
  - Origin policy check: `clientDataJSON.origin` must be an allowed origin.
- Establish normalization, error taxonomy, and test vectors to uncover risks before implementation.

Terminology
- RP ID: Hostname-like string (no scheme, no port); typically the effective domain of the origin (e.g., `example.com`).
- rpIdHash: SHA-256 hash of the RP ID (binary, 32 bytes) carried inside authenticatorData.
- Origin: Scheme + host + optional port (e.g., `https://example.com:443`).

Design Goals
- Bind assertions to the correct RP and browser context.
- Keep rules simple, explicit, and auditable; avoid fuzzy matching (eTLD+1) unless explicitly configured.
- Make dev-mode behavior explicit and safe-by-default.

Configuration Inputs
- `RP_ID` (string): primary RP ID (lowercase host name).
- `Origin` (string): primary origin (scheme://host[:port]).
- `RP_ID_ALLOWLIST` (csv): optional exact RP IDs allowed in addition to `RP_ID` (no wildcards).
- `ORIGIN_ALLOWLIST` (csv): optional exact origins allowed in addition to `Origin` (no wildcards).

Algorithm — RP ID Hash Check
1) Validate `rpID` string from config:
   - Non-empty; contains no scheme (`://`), no path (`/`), and no spaces.
   - Lowercase; optionally strip a single trailing dot.
2) Compute `h = SHA-256([]byte(rpID))`.
3) Compare `h` to `ad.RpIDHash` (constant-time compare not strictly necessary for public values, but acceptable if used consistently).
4) If mismatch, return `ErrRpIdHashMismatch`.
5) Optionally support allowlist: if primary `rpID` mismatches, try each `rp` in `RP_ID_ALLOWLIST` with the same process. Only pass if any match.

Algorithm — Origin Check
1) Parse `cdj.origin` with URL parser; if fails, `ErrOriginMalformed`.
2) Evaluate scheme:
   - Production: require `https`.
   - Dev exception: allow `http` only for `localhost` (exact hostname). IP loopback is not allowed by WebAuthn; keep aligned.
   - If violation, `ErrOriginScheme`.
3) Normalize host to lowercase and strip a single trailing dot for comparison.
4) Determine the expected set of allowed origins: `{ Origin } ∪ ORIGIN_ALLOWLIST` (exact strings).
   - For comparison, match scheme, normalized host, and port (considering default ports 80/443 when unspecified).
   - If none match, return `ErrOriginNotAllowed` with more specific `ErrOriginHost` or `ErrOriginPort` when applicable.
5) Success if any allowed entry equals the candidate origin after normalization.

Normalization
- Hostnames are case-insensitive; compare lowercased.
- Trailing dot: accept `example.com.` as equivalent to `example.com`.
- Punycode/IDNA: defer to configuration providing ASCII (punycode) form; document that if Unicode hostnames are used, the RP should store punycode in config. Future enhancement: normalize via `x/net/idna`.
- Ports: If URL omits port, assume default (80 for http, 443 for https) during equality checks.

Error Taxonomy (Sentinels)
- RP ID: `ErrRpIdInvalid`, `ErrRpIdHashMismatch`.
- Origin: `ErrOriginMalformed`, `ErrOriginScheme`, `ErrOriginHost`, `ErrOriginPort`, `ErrOriginNotAllowed`.

Security Rationale
- Exact matching prevents cross-site or subdomain confusion attacks.
- Dev-only `http://localhost` exception aligns with WebAuthn allowances; avoids accidental broadening to IPs.
- Avoids eTLD+1 heuristics which are easy to misconfigure.

Edge Cases & Decisions
- IP addresses as RP ID: not supported (WebAuthn expects hostnames); treat as invalid `RP_ID`.
- IPv6-literal origins: allowed only if explicitly listed in allowlist and scheme is https.
- Proxy/forwarded headers: origin is client-supplied (browser); do not rely on `X-Forwarded-*` for these checks.
- Multiple environments: add explicit allowlist entries per environment (e.g., staging).

Test Vectors
- See `specs/rp-origin-test-vectors.md` for pass/fail cases across schemes, ports, trailing dots, allowlist usage, and malformed inputs.

Open Questions
- Should loopback IPs (127.0.0.1, ::1) be allowed for dev? Current plan: no; stick to spec’s localhost-only allowance.
- Should we perform IDNA conversion automatically? Current plan: not initially; document requirement to set config in punycode.
- Should subdomains be wildcard-allowed? Current plan: no; use explicit allowlist.

References
- W3C WebAuthn L2/L3: RP ID and origin processing, rpIdHash definition.
- WHATWG URL Standard: origin definition.

