# Passkey Demo — Review and Next Steps

## Purpose
Summarize how the project aligns with its goals and current WebAuthn best practices, and capture prioritized improvement ideas and alternate designs to guide the next phase of work.

## Executive Summary

- Minimal yet robust passkey (WebAuthn) demo in Go + React with strong traceability (blueprint), policy checks (RP ID, origin, UV), canonical CBOR, and careful crypto verification (ES256, low‑S enforcement with high‑S fallback).
- Aligns with best practices (discoverable credentials, UV required, tolerant base64url, strict error envelopes, rate limits). Notable strengths: clean separation of concerns, tests, and description files.
- Largest spec‑alignment gap: deterministic challenge for transaction signing; WebAuthn requires high‑entropy, unpredictable challenges. Mix server randomness while retaining content binding.
- If moving beyond a demo, add CSRF defense, metrics, unified error taxonomy, challenge entropy updates, optional MDS attestation policy, and consider a verification abstraction or established library.

## Fit Against Stated Goals

- Key‑first identity and minimal UI/storage: Achieved. Account identity = COSE public key; SQLite with simple schema and prepared repos; “two buttons” UI with post‑login signing.
- Security defaults: UV required in all ceremonies; RP ID and origin checks include IDNA normalization and dev‑localhost exceptions; signCount semantics accept 0 and enforce monotonicity otherwise; ES256 + low‑S enforced with compatibility fallback.
- Deterministic encoding and server authority: Canonical CBOR via fxamacker; consistent bundle anchors; server‑side session stores for ceremonies; structured logging with correlation IDs.
- Local‑first dev: Single binary, Vite proxy, CORS handled, clean Makefile; comprehensive description files.

## Alignment With Best Practices

- Passkeys and counters: Multi‑device passkeys often report signCount=0 or non‑monotonic; handling matches guidance (accept 0, enforce > stored otherwise).
- UV and discoverable credentials: Requiring UV and residentKey “required” is passkey‑first; some RPs choose UV “preferred” to soften UX, but this choice matches goals.
- Origin and RP ID: IDNA/Punycode normalization and scheme/host/port checks with dev exceptions align with specs and field guidance.
- Signature verification: Low‑S enforcement with high‑S normalization fallback mitigates malleability while preserving compatibility.
- Challenge generation: Login challenges are random; tx challenges are deterministic from content. Spec emphasizes unpredictability/entropy—recommend mixing server randomness while retaining content binding.
- Attestation: “none”/“packed” without trust is fine for a demo; production often treats attestation as optional with MDS‑backed policy.

## Strengths

- Clear, testable policy surface (RP ID hash vs allowlist; origin scheme/host/port granularity) and good error mapping helpers (policy/verify).
- Crypto correctness: COSE EC2 to ECDSA conversion validates curve/coords; strict DER; low‑S enforcement.
- Defensive encoders/decoders: Canonical CBOR; tolerant base64url; robust COSE fallbacks in attestation and bundle handling.
- Operational guardrails: Token‑bucket rate limiting, request body limits, credentialed CORS with correct Vary, session cookie attributes based on https.
- Traceability and documentation: Blueprint, ADRs, user flows, goldens, e2e and fuzz tests, and .desc.md coverage.

## Gaps and Opportunities

### WebAuthn challenge entropy for tx signing
- Issue: Deterministic `challenge = SHA256("CHALv1" || B)` may violate spec expectations for randomness/unpredictability.
- Risk: If other controls regress, deterministic challenges can aid replay and conflict with normative guidance.
- Recommendation: `challenge = SHA256("CHALv1" || B || rand128 || sessionID)`; store `rand128` in tx session. Keep `tx_id = SHA256("TXIDv1" || B)` for idempotence. Continue single‑use tx sessions and nonce monotonic checks.

### Replay/idempotence handling
- Today duplicates likely 500 on insert (PK conflict). Map duplicate transaction insert to 409 Conflict and document behavior.

### CSRF posture for session‑based POSTs
- With cookie sessions and Lax, some navigations may still CSRF. Options: add CSRF token (double‑submit) or set SameSite=Strict and validate Origin/Referer for session‑bound POSTs.

### Observability and metrics
- Add metrics (Prometheus or similar) for options/finish outcomes, verify failure kinds, policy violations, and rate‑limit hits. Keep structured logs.

### Error taxonomy unification
- Centralize error→HTTP mapping using typed errors that carry status, envelope code, and telemetry key. Reduce duplicated mapping branches across handlers.

### Session/TTL store scale‑out
- Keep TTL stores for dev; add an interface and a Redis‑backed implementation for multi‑instance deployments.

### Attestation policy (optional)
- Add a toggle to enforce AAGUID allow/deny lists via MDS3; default off. Document trade‑offs and testing.

### Frontend UX enhancements
- Consider Conditional UI (autofill) for `get()` where supported, and hybrid transport (QR/BT) for cross‑device sign‑in. Maintain UV “required” unless intentionally widened.

### Crypto agility (future)
- Today ES256 only (good). Keep design modular to enable EdDSA later if platform authenticators support it.

## Alternate Designs and Abstractions

- Verification service interface: Extract a `Verifier` that encapsulates CDJ/AD parsing, policy checks, signature verification, and signCount rules; thin handlers call into it. Improves reuse and testability.
- Policy object pattern: Centralize RP ID/origin/UV/signCount in an injectable policy used by login and tx flows to avoid drift.
- Library adoption: Current custom code is clear and tested; as an option, wrap or migrate to a mature library for ceremony validation while keeping your API, repos, and policies.

## Concrete Next Steps

- Fix tx challenge entropy while retaining content binding.
- Map duplicate transaction inserts to HTTP 409.
- Add minimal CSRF defense for session‑bound POSTs (token or Strict + Origin checks).
- Instrument metrics on auth/tx flows; keep existing structured logs.
- Unify error mapping in a small internal package.
- Introduce interfaces for session/TTL stores and add a Redis implementation (optional for scale‑out).
- Add optional attestation policy gated by config.
- Document “Security Posture” (UV required, origin/RP ID policy, signCount rules, challenge entropy, CSRF posture, replay/idempotence).

## References

- W3C WebAuthn L3 (challenge entropy and ceremony requirements): https://www.w3.org/TR/webauthn-3/
- Discussion on challenge randomness: https://github.com/w3c/webauthn/issues/1856
- Yubico WebAuthn best practices: https://developers.yubico.com/WebAuthn/WebAuthn_Developer_Guide/Best_Practices.html
- Passkeys: single vs multi‑device counters: https://developers.yubico.com/Passkeys/Passkey_concepts/Single_device_vs_multi_device_credentials.html
- SimpleWebAuthn passkeys guidance: https://simplewebauthn.dev/docs/advanced/passkeys

## Refs
Refs: goal passkey-registration-login-uv; goal transaction-content-signing; goal minimal-cbor-bundle; goal simple-ui-and-storage; decision encoding-and-ceremony-guardrails; decision webauthn-corrections-and-standardizations

