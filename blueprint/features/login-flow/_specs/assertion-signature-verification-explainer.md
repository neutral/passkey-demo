# R-FLOW-LOGIN — Assertion Signature Verification Explainer (Non‑Developer)

## Purpose & Audience
- Explain what we verify when a passkey (WebAuthn) assertion arrives from the browser during login or signing, in practical terms for technical collaborators.

## What Gets Verified
- The authenticator (e.g., Touch ID) returns a signature over specific bytes derived from the ceremony:
  - `authenticatorData` (binary header produced by the authenticator)
  - `clientDataJSON` (browser JSON describing the ceremony)
- The server computes a digest: `SHA‑256( authenticatorData || SHA‑256(clientDataJSON) )` and checks the signature with the user’s public key.

## Why It Matters
- Prevents tampering: if any byte of the inputs changes (challenge, origin, flags, counters), the signature breaks.
- Binds intent: the `challenge` inside `clientDataJSON` ties the assertion to a server‑issued, short‑lived nonce.
- Enforces best practices: low‑S signatures (ECDSA) reduce ambiguity; strict DER parsing avoids decoder confusion.

## Algorithm & Policy
- Algorithm: ES256 (ECDSA over P‑256 with SHA‑256).
- Digest: `SHA‑256( authenticatorData || SHA‑256(clientDataJSON) )`.
- Encoding: the signature arrives in ASN.1 DER format (two integers, r and s).
- Checks we enforce:
  - Correct curve/algorithm (P‑256 / ES256).
  - Strict DER parsing (reject malformed encodings).
  - Low‑S signatures (s ≤ curve order/2) to avoid multiple valid encodings.

## How It Works in This App
- We parse `authenticatorData` (rpIdHash, flags, signCount) and `clientDataJSON` (type, challenge, origin).
- We compute the digest and verify the DER signature using the account’s public key captured at registration.
- If any piece (data, key, or signature) is wrong or altered, verification fails and login/signing is rejected.

## Errors & Observability
- Clear errors for malformed DER, wrong algorithm/curve, or signature mismatch.
- Logs include concise reasons, not raw data.

## Glossary
- ECDSA (ES256): Elliptic‑Curve Digital Signature Algorithm on curve P‑256 with SHA‑256.
- DER: A strict binary encoding for structured data (ASN.1) used for ECDSA signatures.
- Low‑S: A convention that constrains the signature’s “s” value to avoid multiple valid encodings.

## Refs
- Refs: requirement R-FLOW-LOGIN; requirement R-FLOW-SIGN; requirement R-PLAT-2; decision webauthn-corrections-and-standardizations
