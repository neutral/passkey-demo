# R-PLAT-2 — Single Go Service Backend

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- A single Golang HTTP service implementing all endpoints for registration, login, and transaction signing, serving as the sole backend.

## Depends On
- R-PLAT-3 (SQLite)
- R-SEC-UV, R-ID-KEY
- R-FLOW-REG, R-FLOW-LOGIN, R-FLOW-SIGN
- R-NO-BROKER

## Scope
- In-scope: REST endpoints, session management, WebAuthn verification, SQLite access, static file serving (SPA) as needed for dev.
- Out-of-scope: microservices, brokers, async pipelines.

## Acceptance Criteria
- Implements endpoints:
```text
POST /authn/passkey/registration/options
POST /authn/passkey/registration/finish
POST /authn/passkey/login/options
POST /authn/passkey/login/finish
POST /tx/signing/options
POST /tx/signing/finish
 GET  /tx/list
```
- Runs locally as a single binary; integrates with SQLite; passes E2E flows.

## Flows
- Registration, Login, Transaction signing user flows.

## Interfaces
- JSON over HTTP with base64url for binary fields.

## Risks
- CORS/origin setup for dev; DB connection management; binary encoding bugs.

Refs: goal simple-ui-and-storage; goal no-external-queues; goal server-derived-challenge-and-txid; decision webauthn-corrections-and-standardizations; spec spec-a; spec spec-b; requirement R-PLAT-2
