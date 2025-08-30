# R-SCHEMA-LITE — Lightweight Transaction Schema (CBOR)

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: fr

## Description
- Minimal canonical CBOR bundle for signing: sender COSE EC2 public key, nonce, message, with optional valid_until.

## Depends On
- R-ID-KEY (COSE key as identity)
- R-FLOW-SIGN (signing flow)
- R-PLAT-1, R-PLAT-2

## Scope
- In-scope: schema definition and canonical encoding; server derives challenge and tx_id anchors.
- Out-of-scope: complex typed transactions.

## Acceptance Criteria
- Client produces canonical CBOR bytes B from the schema; server recomputes B identically.
- `challenge = SHA-256("CHALv1" || B)` and `tx_id = SHA-256("TXIDv1" || B)` are stable across runs for same content.

## Schema (CDDL)
```cddl
Bundle = {
  0: CoseKeyEC2Pub,  ; sender_key
  1: uint,           ; nonce
  2: tstr,           ; message
  ?3: uint           ; valid_until (unix seconds)
}

CoseKeyEC2Pub = {
  1: 2,
  3: -7,
  -1: 1,
  -2: bstr .size 32,
  -3: bstr .size 32
}
```

## Flows
- Transaction signing: blueprint/_user-flows/transaction-signing.md

## Interfaces
- Consumed by `/tx/signing/options` and `/tx/signing/finish`.

## Risks
- Non-canonical encodings leading to hash/signature mismatch; bundle drift between client/server.

Refs: goal minimal-cbor-bundle; goal server-derived-challenge-and-txid; decision encoding-and-ceremony-guardrails; spec spec-a; spec spec-b; requirement R-SCHEMA-LITE
