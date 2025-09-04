# AuthN Flows vs Session Middleware — Rationale

## Purpose
Explain why registration/login endpoints under `/authn/*` are not wrapped by the new session middleware (Step 24), while `/tx/*` routes are.

## Overview
- The session middleware enforces a logged‑in context via the `sid` cookie and is applied to protected `/tx/*` routes.
- `/authn/*` endpoints (registration/login) must work pre‑auth and some create the session itself, so they remain unchanged.

## Rationale
- Pre‑auth endpoints: Registration and login flows use ceremony‑scoped stores (RegSessionStore/LoginSessionStore) and must be callable without `sid`.
- Session creation timing: `login/finish` establishes the server session and sets `sid`; requiring `sid` beforehand would 401 the flow.
- Minimal change: Step 24’s goal is to unify auth for protected routes with minimal disruption; changing `/authn/*` would broaden scope and risk regressions.
- Separation of concerns: `/authn/*` remains ceremony‑scoped; `/tx/*` consumes logged‑in SessionContext via middleware.
- Compatibility: Tx handlers include a cookie fallback to keep tests and non‑middleware wiring functional.

## Implications
- `/tx/*` requires an established session; `/authn/*` does not.
- Future steps (limits, envelopes) can wrap `/authn/*` with non‑auth middlewares (rate limits/CORS) without requiring `sid`.

## Refs
Refs: requirement R-FLOW-LOGIN; requirement R-FLOW-REG; requirement R-PLAT-2; goal passkey-registration-login-uv; decision webauthn-corrections-and-standardizations

