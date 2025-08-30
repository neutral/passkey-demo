# R-NO-BROKER — No Message Broker

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- Eliminate external brokers/queues (e.g., JetStream/NATS). All operations are synchronous within the single backend service.

## Depends On
- R-PLAT-2 (backend service)

## Scope
- In-scope: synchronous request processing; direct DB writes.
- Out-of-scope: async pipelines, distributed messaging.

## Acceptance Criteria
- No broker dependencies in code or deployments; flows complete synchronously and persist to SQLite.

## Flows
- Applies to all flows.

## Interfaces
- N/A beyond existing REST endpoints.

## Risks
- Throughput limits (acceptable for demo).

Refs: goal no-external-queues; spec spec-a; spec spec-b; requirement R-NO-BROKER
