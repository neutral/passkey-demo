# AuthN Flows vs Session Middleware — Rationale

## Purpose
Explain why registration/login endpoints under `/authn/*` are not wrapped by the new session middleware (Step 24), while `/tx/*` routes are.

## Overview
- The session middleware enforces a logged‑in context via the `sid` cookie and is applied to protected `/tx/*` routes.
- `/authn/*` endpoints (registration/login) work pre‑auth; `login/finish` creates the session, so `/authn/*` remains unauthenticated.

## Rationale
- Pre‑auth endpoints: Registration and login flows use ceremony‑scoped stores (RegSessionStore/LoginSessionStore) and must be callable without `sid`.
- Session creation timing: `login/finish` establishes the server session and sets `sid`; requiring `sid` beforehand would 401 the flow.
- Separation of concerns: `/authn/*` remains ceremony‑scoped; `/tx/*` consumes logged‑in SessionContext via middleware.

## Implications
- `/tx/*` requires an established session; `/authn/*` does not.
- Group middlewares (body/rate limits) are applied to both `/authn/*` and `/tx/*` without changing auth requirements.

## Refs
Refs: requirement R-FLOW-LOGIN; requirement R-FLOW-REG; requirement R-PLAT-2; goal passkey-registration-login-uv; decision webauthn-corrections-and-standardizations
