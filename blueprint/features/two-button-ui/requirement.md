# R-UI-2BTN — Two-Button UI + Dashboard

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: fr

## Description
- Minimal UI emphasizing two primary actions (Register, Login). After login, show a dashboard with a list of signed messages and an input to add a signed message.

## Depends On
- R-PLAT-1 (frontend minimal React)
- R-FLOW-REG, R-FLOW-LOGIN, R-FLOW-SIGN

## Scope
- In-scope: minimal styling; clear buttons; post-login form and list; simple error toasts/messages.
- Out-of-scope: design system, theming, advanced routing.

## Acceptance Criteria
- Home screen contains only Register and Login buttons.
- After login, dashboard renders signed messages list and input to submit new message which triggers signing flow.
- Errors surface clearly and non-intrusively.

## Flows
- Registration, Login, Transaction signing user flows apply.

## Interfaces
- Frontend-only presentation; consumes the flows’ HTTP endpoints.

## Risks
- UX confusion with authenticator prompts; empty states.

Refs: goal ui-simplicity-two-buttons; goal simple-ui-and-storage; spec spec-a; spec spec-b; requirement R-UI-2BTN
