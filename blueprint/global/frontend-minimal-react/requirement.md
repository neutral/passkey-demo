# R-PLAT-1 — Frontend Minimal React (no external state management)

## Metadata
- State: Draft
- Date: 2025-08-30
- Owners: passkey-demo maintainers
- Type: nfr

## Description
- A minimal React single-page application (SPA) with functional components and no external state management libraries. Uses native React hooks and relies on the backend session for auth state. Implements the demo UI for register, login, and post-login message signing, invoking WebAuthn with userVerification required in all ceremonies. Binary fields are base64url in JSON; bundle encoding is canonical CBOR.

## Depends On
- R-PLAT-2 (single Go service backend)
- R-SEC-UV (user verification required)
- R-FLOW-REG (registration flow)
- R-FLOW-LOGIN (login flow)
- R-FLOW-SIGN (transaction signing flow)
- R-UI-2BTN (two-button primary UI)
- R-OPS-DEV (easy local dev)
- R-ERR (clear error responses and limits)
- R-PORTABLE (works in modern browsers/platform authenticators)

## Scope
- In-scope: React SPA scaffolding; Register/Login/Dashboard views; WebAuthn API calls using server-supplied options; canonical CBOR bundle construction for signing; base64url conversions; simple error display; integration with backend endpoints; local dev origin http://localhost:5173 aligned with RP/origin checks.
- Out-of-scope: SSR or multi-page routing; external state libraries (Redux, MobX, Zustand, etc.); complex design system/theming; internationalization; offline support; message brokers.

## Acceptance Criteria
- Builds and runs locally as an SPA at origin http://localhost:5173.
- No external state management libraries are used; state is limited to React hooks and server session.
- UI exposes two primary actions (Register, Login); after login, shows dashboard with signed messages list and an input to add a signed message.
- WebAuthn is invoked via server-provided options for registration, login, and transaction signing with userVerification set to "required"; origin and rpId validations pass.
- Client constructs canonical CBOR bundle for signing; binary fields exchanged with the server are base64url-encoded.
- Interacts successfully with backend endpoints to complete register → login → sign flows; error states are clearly surfaced.

## Flows
- Registration: blueprint/_user-flows/registration.md
- Login: blueprint/_user-flows/login.md
- Transaction signing: blueprint/_user-flows/transaction-signing.md

## Interfaces
- HTTP API (JSON over HTTP; base64url for binary). Endpoints used by the frontend:
  - POST /authn/passkey/registration/options
  - POST /authn/passkey/registration/finish
  - POST /authn/passkey/login/options
  - POST /authn/passkey/login/finish
  - POST /tx/signing/options
  - POST /tx/signing/finish
  - GET  /tx/list

## Risks
- Browser WebAuthn differences and platform authenticator prompts; origin/rpId mismatch during local dev; base64url/ArrayBuffer conversion errors; canonical CBOR drift; CORS configuration issues; UX confusion on authenticator interactions.

Refs: goal simple-ui-and-storage; goal ui-simplicity-two-buttons; goal webauthn-policy-defaults; requirement R-PLAT-1; spec spec-a; spec spec-b
