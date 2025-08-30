Title: Passkey Demo Goals
Date: 2025-08-30
Owners: passkey-demo maintainers
Status: Draft

---

## Design Tenets

- Simplicity: minimal UI, minimal backend, no queues.
- Key-first identity: account = passkey public key (COSE).
- Deterministic encoding: canonical CBOR for all signed bundles.
- Security by default: user verification (UV) required in all ceremonies.
- Server authority: server supplies WebAuthn options; bind challenges to content.
- Local-first dev: single process API + SQLite; easy to run.

## Non-Goals

- Federation, SSO, or multi-tenant account models.
- Attestation trust-chain verification beyond "none".
- Distributed systems, message brokers, or async pipelines.
- Complex UI/UX flows beyond register, login, and sign.
- Hardware-backed HSM/KMS integrations or multi-algorithm support.

## Goals

- goal passkey-registration-login-uv: Users can register and log in with a platform authenticator; UV is required in every ceremony.

- goal transaction-content-signing: Users can sign transaction content (CBOR bundle) with their passkey; server verifies an ES256 signature bound to the content.

- goal key-first-identity-cose: Use the passkey’s COSE public key as the account identifier; no separate username is required.

- goal minimal-cbor-bundle: Define and use a minimal canonical CBOR transaction bundle (nonce + message [+ optional valid_until]).

- goal simple-ui-and-storage: Provide a minimal React UI (Register, Login, Sign Message) and store accounts/credentials/signed messages in SQLite.

- goal no-external-queues: Run synchronously on a single server; do not introduce external brokers or queues.

- goal webauthn-policy-defaults: Enforce UV, use discoverable credentials (residentKey: "required"), and set attestation: "none".

- goal ui-simplicity-two-buttons: Keep primary UI actions to two buttons (Register, Login) and a post-login sign form.

- goal server-derived-challenge-and-txid: Derive `challenge = SHA-256("CHALv1" || B)` and `tx_id = SHA-256("TXIDv1" || B)` on the server.

## Success Metrics

- End-to-end demo runs locally: register → login → sign message → view verified entry in list.
- UV flag verified server-side in all ceremonies; attempts without UV are rejected.
- CBOR bundle encoding is canonical and stable across runs (hashes match).
- Signatures verify as ES256 with low-S enforcement; signCount strictly increases.
- SQLite persists accounts, credentials, and signed messages; no external brokers used.
- Discoverable credentials are created (residentKey "required"); attestation set to "none".

## Change control

- Changes to these goals require an ADR.
- All new requirements/specs/ADRs must include: Refs: goal <name> for relevant goals.
