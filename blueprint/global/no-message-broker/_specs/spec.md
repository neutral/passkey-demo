# R-NO-BROKER — Architecture Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Single-process synchronous request handling; direct persistence to SQLite; no brokers.

## Interfaces
- Same REST surface; no pub/sub endpoints.

## Data / Models
- Unchanged relative to DB schema.

## Algorithms
- Immediate verification and storage within HTTP request lifecycle.

## Security / Privacy
- Simpler threat surface without broker credentials.

## Errors / Observability
- Standard HTTP error handling; linear tracing through handler stack.

## Testing Strategy
- Confirm absence of broker config/clients; E2E flow latency acceptable.

## Open Questions
- None.

Refs: spec spec-a; spec spec-b; goal no-external-queues; requirement R-NO-BROKER
