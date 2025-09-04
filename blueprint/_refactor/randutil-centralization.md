# Centralize Cryptographic Random Byte Generation

## Purpose
- Eliminate duplicated `randBytes` helpers by providing a single, testable utility for cryptographic randomness.

## Context
- `randBytes` currently lives in multiple files (registration, login, tx) with identical logic.

## Proposal
- Add `internal/util/randutil`:
  - `func Bytes(n int) ([]byte, error)` using `crypto/rand`.
  - Optional: `func MustBytes(n int) []byte` for tests.
  - Interface injection for tests when determinism is needed (`type Reader interface{ Read([]byte)(int,error) }`).

## Migration Plan
1) Introduce `randutil` and unit tests (entropy length, error surfacing).
2) Replace usages in reg/login/tx with `randutil.Bytes`.
3) Remove duplicated helpers.

## Risks & Mitigations
- Minimal risk; keep API small and explicit.

## Testing Strategy
- Unit: ensure correct lengths and error propagation; fuzz for edge lengths (0, 1, 32, 1024).

## Acceptance Criteria
- All call sites use `randutil.Bytes`; no duplicate helpers remain.

## Refs
- Refs: requirement R-PLAT-2; requirement R-OPS-DEV; decision encoding-and-ceremony-guardrails

