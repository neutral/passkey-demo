# R-UI-2BTN — UI Spec

## Metadata
- Status: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers

## Overview
- Simple SPA layout with two primary actions and a post-login dashboard for signing and listing messages.

## Interfaces
- Components: Register button, Login button; Dashboard view (list + input).

## Data / Models
- Uses session presence to toggle dashboard view; message model is free-text submitted into signing bundle.

## Algorithms
- Glue to invoke registration, login, and signing flows via the backend endpoints; display results and errors.

## Security / Privacy
- No sensitive data persisted in local storage; rely on HttpOnly cookies.

## Errors / Observability
- Inline error banners or toast; console logs in dev only.

## Testing Strategy
- Manual flow verification for UI states before/after auth; error states.

## Open Questions
- Navigation behavior post-login; minimal routing or single-screen swap.

Refs: spec spec-a; spec spec-b; goal ui-simplicity-two-buttons; requirement R-UI-2BTN
