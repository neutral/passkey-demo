# RP ID and Origin — Test Vectors

Purpose: Concrete pass/fail cases to validate Step 14 behavior and uncover edge-case issues.

Assumptions
- Config: `RP_ID=example.com`, `Origin=https://example.com`, no allowlists unless specified.

RP ID Hash Checks
- Pass: `rpID=example.com` → `SHA256("example.com")` equals `ad.rpIdHash`.
- Fail: `rpID=example.org` (mismatch) → `ErrRpIdHashMismatch`.
- Fail: `rpID=https://example.com` (contains scheme) → `ErrRpIdInvalid`.
- Fail: `rpID=example.com/` (contains path) → `ErrRpIdInvalid`.
- Pass: `rpID=example.com.` (trailing dot) treated as `example.com`.
- Pass with allowlist: `RP_ID_ALLOWLIST=app.example.com` and `ad.rpIdHash=SHA256("app.example.com")`.

Origin Checks (no allowlist)
- Pass: `https://example.com` (no port) vs expected → OK (default 443).
- Pass: `https://example.com:443` vs expected → OK.
- Fail (port): `https://example.com:444` → `ErrOriginPort` (or `ErrOriginNotAllowed`).
- Fail (scheme): `http://example.com` → `ErrOriginScheme`.
- Fail (host): `https://app.example.com` → `ErrOriginHost`.
- Fail (malformed): `:` or empty string → `ErrOriginMalformed`.
- Pass (trailing dot): `https://example.com.` → OK.
- Pass (case): `https://EXAMPLE.COM` → OK (host lowercased).

Origin Checks with allowlist
- Config: `ORIGIN_ALLOWLIST=https://app.example.com,https://example.net:8443`.
- Pass: `https://app.example.com` → OK.
- Pass: `https://example.net:8443` → OK.
- Fail: `https://example.net` (missing 8443) → `ErrOriginPort`.

Development exception
- Pass: `http://localhost:5173` only when dev exception is enabled (policy flag or inferred dev mode).
- Fail: `http://127.0.0.1:5173` → not allowed by spec; treat as `ErrOriginScheme` or `ErrOriginNotAllowed`.

Punycode / IDNA (documented behavior)
- If `RP_ID` is stored as punycode `xn--bcher-kva.ch`, then:
  - Pass: ad.rpIdHash uses `SHA256("xn--bcher-kva.ch")`.
  - Pass: origin `https://xn--bcher-kva.ch` (host lowercased).
  - Note: If Unicode form is used in config, behavior is undefined; document punycode requirement or add IDNA normalization later.

IPv6 literal origin (explicit)
- Config allowlist contains `https://[2001:db8::1]:8443`.
- Pass: exact origin string matches (scheme/host/port).
- Fail: different port or missing brackets → malformed or not allowed.

