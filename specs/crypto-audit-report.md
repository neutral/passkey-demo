# Crypto Audit Report

Scope: Internal crypto and WebAuthn verification utilities in this repository, specifically:

- `server/internal/crypto/crypto_cose.go`: COSE EC2 → ECDSA key conversion
- `server/internal/webauthn/sig.go`: WebAuthn ES256 assertion verification
- `server/internal/webauthn/ad.go`: authenticatorData parsing (flags, counter)
- `server/internal/webauthn/cdj.go`: clientDataJSON parsing (type, challenge, origin)
- `server/internal/encoding/b64url.go`: tolerant base64url encode/decode
- `server/internal/encoding/cbor.go`: canonical CBOR encode/decode (fxamacker)

Date: 2025-09-01

Author: Internal audit (engineering)

---

## Summary

The implementation enforces ES256 (ECDSA over P-256, SHA-256) for WebAuthn assertion verification, requires strict ASN.1 DER with no trailing bytes, and rejects high-S signatures to prevent malleability. COSE EC2 keys are validated for type, algorithm, curve, coordinate length, and on-curve membership before conversion to Go’s `ecdsa.PublicKey`. Canonical CBOR encoding uses fxamacker/cbor with canonical options. Encoding utilities provide unpadded base64url and tolerant decoding.

The choices align with WebAuthn normative requirements and common security guidance. Additional hardening recommendations are noted below.

---

## Components and Decisions

### WebAuthn Signature Verification (`webauthn/sig.go`)

- Algorithm: ES256 (ECDSA P-256, SHA-256).
- Message: `SHA256(authenticatorData || SHA256(clientDataJSON))` per WebAuthn.
- DER Parsing: Strict ASN.1 DER with no trailing bytes; rejects malformed DER.
- Low-S Enforcement: Rejects signatures where `s > N/2` to avoid malleability.
- Curve: Only P-256 accepted; nil or other curves rejected.

Rationale:
- ES256 is the required baseline for WebAuthn and well supported across authenticators and browsers.
- Strict DER parsing + low-S align with best practices (e.g., Bitcoin Core rules, Go TLS stack approaches) to prevent signature malleability and parse ambiguities.

Residual risks:
- Side-channel characteristics of ECDSA verification depend on Go stdlib; generally acceptable for verification, but be mindful if reused in contexts involving secret material.
- Incorrect upstream inputs (e.g., wrong RP ID/origin checks) would undermine successful verification; these checks are planned in subsequent steps.

Recommendations:
- Add exported sentinel errors to distinguish failure causes (malformed DER, high-S, bad signature, wrong curve) for better observability.
- Consider adding timing-invariant compare where needed for non-crypto equality checks adjacent to verification (not needed for ECDSA Verify itself).

### COSE EC2 → ECDSA Conversion (`internal/crypto/crypto_cose.go`)

- Validations: `kty=2 (EC2)`, `alg=-7 (ES256)`, `crv=1 (P-256)`, 32-byte X/Y, point on curve.
- Construction: Returns `ecdsa.PublicKey{Curve: P256, X, Y}`.

Rationale:
- Strictly constrain accepted COSE keys to ES256/P-256 for simplicity and security.

Residual risks:
- None significant if inputs are controlled by WebAuthn attestation/registration; ensure that registration path validates attested key consistency once implemented.

Recommendations:
- Optionally detect and reject the point at infinity (implicitly rejected by IsOnCurve, but double-checking via zero X/Y already present in tests).

### Authenticator Data Parsing (`webauthn/ad.go`)

- Parses rpIdHash (32 bytes), flags (1 byte), and signature counter (big-endian uint32). Returns remainder for potential attestation/extension parsing.
- Helpers `HasUP`, `HasUV` extracted for clarity.

Rationale:
- Minimal, spec-compliant extraction needed for assertion verification.

Residual risks:
- None, provided downstream steps perform RP ID hash comparison and counter checks.

Recommendations:
- Add rpIdHash comparison during request handling (Step 14+).
- Track and enforce monotonic `signCount` per credential to detect cloned authenticators.

### ClientDataJSON Parsing (`webauthn/cdj.go`)

- Parses `type`, `challenge` (base64url), and `origin`.
- Validates `type` ∈ {`webauthn.get`, `webauthn.create`}.
- Decodes challenge via tolerant base64url.

Rationale:
- Extract only necessary fields for assertion finishing; keep challenge decoding tolerant to padding variance.

Residual risks:
- Origin validation must be enforced in the higher-level request handler per RP policy.

Recommendations:
- Enforce origin and RP ID constraints in finish endpoints; ensure challenge freshness/TTL.

### Base64url Utilities (`encoding/b64url.go`)

- `Encode` uses unpadded base64url.
- `Decode` is tolerant to padded/unpadded input, normalizes padding when needed.

Rationale:
- Browser/WebAuthn implementations vary in padding; tolerance improves interop while preserving correctness.

Residual risks:
- Tolerance can mask minor formatting inconsistencies; mitigated by subsequent cryptographic checks.

Recommendations:
- Log normalization events in higher-level code paths if needed for telemetry.

### Canonical CBOR (`encoding/cbor.go`)

- Uses fxamacker/cbor with canonical options to ensure deterministic encoding.

Rationale:
- Deterministic encoding is essential if any signed bundles or hashes are derived; although WebAuthn assertion verification here does not directly sign CBOR, other features reference canonical CBOR.

Residual risks:
- None significant; fxamacker is widely used and actively maintained.

Recommendations:
- Pin version and review release notes before upgrades.

---

## Testing and Coverage

Implemented unit tests cover:

- COSE → ECDSA conversion: valid key, invalid parameters, off-curve, zero point.
- WebAuthn signature verification:
  - Happy path with low-S signature.
  - High-S signature rejection (s → N - s).
  - Tampering of `authenticatorData` and `clientDataJSON` causes verification failure.
  - Malformed DER and DER with trailing bytes rejected.
  - Wrong curve and nil key rejected.
- ClientDataJSON: type validation, tolerant challenge decoding, missing fields, malformed JSON.
- AuthenticatorData: correct parsing, flag helpers, big-endian counter.
- Encoding: base64url roundtrips and tolerance; CBOR canonical determinism and roundtrip.

Gaps and suggestions:

- Add tests for RP ID hash comparison and origin enforcement once handlers are implemented.
- Add sign-count anti-cloning checks (monotonicity) tests after persistence wiring.
- Consider fuzz tests for ASN.1 signature parsing and COSE map decoding.

---

## Configuration and Policy

- Curves: Only P-256 accepted; no support for P-384/P-521 or EdDSA here to reduce complexity and surface area.
- Hash: SHA-256 per WebAuthn ES256; no configurability to avoid downgrade risk.
- DER: Strict parsing, no trailing data; low-S enforced.
- Base64url: Encode unpadded; accept padded/unpadded input.
- CBOR: Canonical encoding for determinism; strict decoding in tests.

Policy implications:

- Keep cryptographic algorithm choices fixed unless feature requirements demand expansion, at which point introduce explicit negotiation and per-RP policy gates.
- Treat signature verification failures as security events; instrument logging/metrics with care to avoid leaking sensitive content.

---

## Threat Model Highlights

- Attacker cannot forge ES256 signatures without the authenticator private key.
- Malleability mitigated by low-S requirement and strict DER.
- Replay mitigated by server-managed challenge freshness and signCount monotonic checks (to be implemented in finish handlers).
- Cross-origin/relying party abuse mitigated by origin checks and rpIdHash comparison (to be implemented in finish handlers).

Out of scope here:

- Attestation trust chain validation (registration phase) and authenticator attestation security posture.
- Hardware-backed key protections on clients; assume platform or roaming authenticators enforce their own security.

---

## Recommendations and Next Steps

1. Implement and test RP ID hash comparison and origin enforcement in assertion finish handlers.
2. Store and enforce monotonic `signCount` with alarms on regressions.
3. Add exported sentinel error variables for signature failure classes to improve observability.
4. Add fuzzing for ASN.1 signature parser and COSE → ECDSA conversion.
5. Consider adding optional support for ES384/Ed25519 only if required, gated by explicit policy.
6. Add structured logging around verification failures (rate limited, without sensitive payloads).

---

## References

- W3C WebAuthn Level 2: Assertion Signature Verification and Data Structures
- IETF COSE (RFC 8152/9052): COSE_Key parameters for EC2
- NIST SP 800-56A Rev.3: ECDSA guidance
- Go crypto/ecdsa documentation
- fxamacker/cbor documentation

