# AuthN Flows vs Session Middleware — Rationale

## Purpose
Clarify why registration endpoints under `/authn/*` are not wrapped by Step 24’s session middleware, whereas `/tx/*` routes are.

## Overview
- Session middleware enforces a logged‑in context via `sid` and applies to protected `/tx/*` routes.
- Registration (`/authn/passkey/registration/*`) operates pre‑auth; `login/finish` creates the session; thus `/authn/*` stays unauthenticated.

## Rationale
- Pre‑auth ceremonies: Registration and login use ceremony‑scoped stores and must succeed without `sid`.
- Session creation happens at `login/finish`; enforcing `sid` earlier would block the flow.
- Minimal‑change principle: Step 24 targets protected routes first to avoid regressions in authn.
- Separation: Keep ceremony logic self‑contained; let `/tx/*` rely on SessionContext.

## Implications
- `/authn/*` remains accessible without `sid`; `/tx/*` requires it.
- Group middlewares (limits/CORS) may wrap `/authn/*` without changing auth.

## Refs
Refs: requirement R-FLOW-REG; requirement R-FLOW-LOGIN; requirement R-PLAT-2; goal passkey-registration-login-uv; decision webauthn-corrections-and-standardizations
