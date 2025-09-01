# RP ID and Origin — Risks, Ambiguities, and Decisions

Objective
- Surface subtle issues in RP ID and Origin checks before implementation to avoid footguns and regressions.

Key Risks
- Misinterpreting RP ID vs Origin: Treating RP ID like a URL (scheme/port) will break rpIdHash matching.
- Overbroad matching: Using eTLD+1 or wildcard patterns can accidentally trust subdomains operated by third parties.
- IDNA/Unicode pitfalls: Unicode domains must be punycoded; mixing forms between config and runtime inputs leads to mismatches.
- Trailing dot handling: DNS canonical hostnames may include a trailing dot; comparison must accept this variant.
- Default ports: Failing to normalize default ports creates false mismatches (`https://example.com` vs `https://example.com:443`).
- Dev exceptions: Allowing `http` too broadly (beyond `localhost`) undermines WebAuthn guarantees.
- Allowlist sprawl: Large allowlists become hard to audit; prefer minimal, explicit entries.

Decisions (current plan)
- Exact matching only; no eTLD+1 or wildcards. Subdomains must be enumerated explicitly in allowlists when truly needed.
- `http` only for `localhost` in dev; IP loopback not allowed to align with spec behavior.
- Require configuration to use ASCII (punycode) for internationalized domains; revisit automatic IDNA normalization later.
- Accept a single trailing dot on hostnames during comparison.
- Normalize hostnames to lowercase; preserve scheme and port semantics.

Operational Considerations
- Multi-env workflows: Add staging/dev origins to allowlist rather than toggling validation off.
- Reverse proxies: Origin comes from the browser’s clientDataJSON, not proxy headers, so standard proxy origin gotchas do not apply.
- Telemetry: Map origin/RP errors to structured error kinds for safe logging and alerting.

Open Questions
- Should we permit loopback IPs in dev to ease testing? (Leaning no; would diverge from WebAuthn’s localhost-only exception.)
- Should we add automatic IDNA normalization using `x/net/idna`? (Likely yes, as a follow-up behind a deterministic function with tests.)
- Should we permit an RP ID allowlist, or pin strictly to one RP ID? (Allowlist is supported but default is a single RP ID.)

Next Steps
- Implement accordance with `specs/webauthn-rp-origin-verification-spec.md`.
- Validate with `specs/rp-origin-test-vectors.md` and add any missing cases.
- Reassess IDNA and loopback IP policy based on dev feedback.

