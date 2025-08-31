# R-PLAT-2 — Common Types Usage Explainer (Non‑Developer)

## Purpose & Audience
- Explain the shared “types” (standard data shapes) used by the backend so non‑developer teammates can understand what flows through APIs and why these shapes exist.

## What “Types” Are
- Plain, named structures that define the shape of data the app sends and receives (e.g., a passkey public key, a login payload, or the content to be signed).
- Think of them like consistent templates: every time we talk about a “Bundle” or a “Credential”, it follows the same template everywhere.

## Why We Use Them Here
- Consistency: one source of truth for field names and formats across registration, login, and signing.
- Safety: fewer mistakes when converting binary data (keys, signatures) to/from network‑friendly strings.
- Traceability: types tie directly back to requirements (identity model, signing bundle) and are easy to audit.

## What We Cover
- COSE EC2 Public Key: the passkey’s public key (binary X/Y coordinates plus algorithm metadata).
- Signing Bundle: the content users sign (sender key + nonce + message + optional expiry) in a compact, deterministic format.
- WebAuthn Payloads: minimal request/response shapes for registration and login (binary fields carried as base64url strings in JSON).

## How It Works in This App
- The backend declares these types once and uses them across handlers and helpers.
- At the API boundary, binary fields (like keys and signatures) are base64url strings; inside the app, they become bytes for verification.
- The signing bundle is encoded as canonical CBOR (a small binary format) so the same content always produces the same bytes for hashing/signing.

## Performance & Safety Settings
- Canonical CBOR: keeps bundle encoding stable and compact.
- Base64url: safe, URL‑friendly way to carry binary data in JSON.

## Security & Privacy
- Public keys and signatures are handled as raw bytes internally but are not logged directly.
- Only short, non‑sensitive identifiers (like thumbprints) appear in logs or UI when needed.

## Operating It Day‑to‑Day
- No direct operator actions required. The value is in predictable request/response shapes and fewer encoding surprises.

## Limitations & When to Upgrade
- These types are intentionally minimal for the demo. If we introduce more features (extra fields, multiple credentials per account), we extend types deliberately and version APIs as needed.

## Errors & Observability
- If a payload is malformed (wrong encoding or missing fields), the backend returns a clear error code; logs include just enough context to troubleshoot.

## Glossary
- COSE EC2: a standard way to represent an elliptic‑curve public key used by passkeys.
- CBOR (canonical): a compact binary format with a deterministic encoding rule set.
- base64url: a safe textual encoding for binary data used in URLs/JSON.

## Refs
- Refs: requirement R-PLAT-2; requirement R-ID-KEY; requirement R-SCHEMA-LITE; goal key-first-identity-cose; goal minimal-cbor-bundle
