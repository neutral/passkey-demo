# R-SCHEMA-LITE — Canonical CBOR Usage Explainer (Non‑Developer)

## Purpose & Audience
- Explain what CBOR is and how/why we use its canonical form for transaction content in this demo, for technical collaborators who aren’t writing code.

## What CBOR Is
- A compact, binary format for structured data (similar to JSON, but smaller and faster to parse).
- Supports maps/arrays/strings/numbers and binary blobs; designed for reliability and cross‑platform use.

## Why Canonical CBOR Here
- Determinism: the same content must always produce the same bytes so hashes and signatures are stable and reproducible.
- Simplicity: canonical rules (sorted map keys, shortest encodings, no indefinite lengths) remove ambiguity between different encoders.
- Security: prevents “different bytes for the same content” tricks that could break signature verification.

## What We Encode
- The “Bundle” that users sign:
  - Sender key (public key) in COSE EC2 representation
  - Nonce (strictly increasing number)
  - Message (free‑form text)
  - Optional valid‑until timestamp
- The Bundle is encoded as canonical CBOR bytes; all hashing and signing is done over these exact bytes.

## How It Works in This App
- The frontend builds the Bundle object and encodes it to canonical CBOR.
- The backend decodes and re‑encodes canonically to confirm the exact bytes, then derives:
  - Challenge: `SHA‑256("CHALv1" || B)` used in WebAuthn
  - Transaction ID: `SHA‑256("TXIDv1" || B)` used for records
- Binary at the API boundary: CBOR bytes travel as base64url strings in JSON.

## Performance & Safety Settings
- Canonical encoding via a mature library with deterministic options.
- Small payloads (a few hundred bytes) — fast to encode/decode; negligible overhead for a demo.
- No indefinite‑length items; shortest integer/length forms; stable map key ordering.

## Security & Privacy
- Canonicalization thwarts ambiguity attacks on signatures.
- Only public/derived values are signed and logged; raw CBOR bytes are not logged.
- CBOR is an encoding, not encryption — transport remains HTTPS.

## Operating It Day‑to‑Day
- No operator action required. If data seems different across systems, compare the CBOR hex for the Bundle to confirm byte‑for‑byte equality.
- To reset: this lives in app logic; no separate service or config to manage.

## Limitations & When to Upgrade
- Not a schema language by itself. For richer, evolving models, consider a documented schema (CDDL is already provided) and versioning fields.
- If payloads grow or interop demands it, add golden vector tests and schema evolution guidelines.

## Errors & Observability
- Malformed CBOR or non‑canonical encodings are rejected with clear errors.
- Determinism checked in tests (same content → same bytes); failures indicate encoder misconfiguration or drift.

## Glossary
- CBOR: Concise Binary Object Representation, a binary alternative to JSON.
- Canonical CBOR: A deterministic subset with strict encoding rules to ensure stable bytes.
- Bundle: The signed transaction content (key + nonce + message [+ optional expiry]).

## Refs
- Refs: requirement R-SCHEMA-LITE; decision encoding-and-ceremony-guardrails; requirement R-PLAT-2; goal server-derived-challenge-and-txid; goal minimal-cbor-bundle
