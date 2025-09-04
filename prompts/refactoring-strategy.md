# Refactoring Strategy Prompt

## Purpose
- Provide a repeatable, lightweight process to discover, evaluate, and document refactoring opportunities without prematurely changing code.
- Align refactor ideas with goals, requirements, decisions, and the active implementation plan for traceability and safe sequencing.

## When To Use
- You’re surveying a codebase for maintainability improvements (duplication, cross‑cutting concerns, testability) while honoring a “minimal change” posture.
- You plan to capture proposals under `blueprint/_refactor/` for future consideration, not to implement immediately.

## Inputs
- Code: source files and their `*.desc.md` descriptions.
- Blueprint: `blueprint/goals.md`, `blueprint/(features|global)/**/requirement.md`, related `_specs/spec.md` and explainers, `_decisions/*.md`, `_user-flows/*.md`.
- Plan: `blueprint/implementation.md` (active steps, Done list, quality gates).
- Tests: existing unit/integration tests and patterns for validation.

## Process (Reflective + Actionable)
1) Orient
- Read `blueprint/goals.md` and decisions to internalize design tenets and constraints (e.g., canonical encoding, UV required, server‑derived challenges).
- Skim relevant requirements/specs and the active implementation step to avoid proposing refactors that conflict with in‑flight work.

2) Inventory Hotspots
- Look for repetition and cross‑cutting seams using fast search:
  - Session stores, error handling/mapping, randomness utilities, router wiring, DB query strings, data validation.
- Scan `*.desc.md` to learn responsibilities, invariants, and relationships; follow those to callers/consumers.
- Note where behavior is consistent but implementation is duplicated.

3) Trace To Artifacts
- For each candidate, identify which goals/requirements/NFRs it supports (e.g., R‑ERR for error envelopes, R‑PLAT‑2 for service cohesion).
- Check ADRs for accepted patterns to standardize on (e.g., encoding guardrails, policy checks).

4) Risk & Scope Triage
- Classify each idea by blast radius and dependency graph:
  - Low: leaf utilities or additive modules (preferred first).
  - Medium: shared stores/wiring with stable public APIs.
  - High: changes to verification/crypto or core request flows (defer; require ADR).
- Favor incremental migration over big‑bang rewrites; introduce shims where helpful.

5) Propose (Doc‑First)
- For each idea, write a focused doc under `blueprint/_refactor/` with:
  - Purpose, Context, Proposal (APIs/modules), Migration Plan (incremental), Risks & Mitigations, Testing Strategy, Acceptance Criteria, Refs.
- Keep proposals concrete but technology‑agnostic enough to allow evolution.

6) Prioritize
- Use a simple matrix: Impact (maintainability, reliability) × Effort × Risk.
- Prefer items that reduce duplication and improve testability without altering behavior.

7) Plan Integration
- Ensure proposals don’t block current steps; if they help upcoming steps, reference where they would slot in (e.g., after Step 24 middleware).
- If a change affects public contracts or policies, consider drafting an ADR first.

## Heuristics & Principles
- Minimal surface change: preserve public APIs while refactoring internals.
- Centralize cross‑cutting concerns: error envelopes, session TTL policy, randomness, route registration, prepared statements.
- Improve test seams: inject `now`, `rand`, and thin DB interfaces in builders; keep handlers thin.
- Prefer composition and small, cohesive packages; avoid deep hierarchies.
- Maintain deterministic behavior and security invariants (e.g., canonical CBOR, UV required).

## Output Requirements
- Do not modify code during the survey.
- Produce one markdown file per idea under `blueprint/_refactor/` using the template below.
- Keep “Refs” lines accurate to goals/requirements/specs/decisions.

## Refactor Doc Template (Minimal)

Title: <Short, action‑oriented name>

## Purpose
- <Why this refactor matters; user‑visible benefits avoided>

## Context
- <Where duplication/pain exists; current patterns; constraints>

## Proposal
- <APIs/modules to add or rehome; responsibilities; example signatures>

## Migration Plan (Incremental)
1) <Step>
2) <Step>
3) <Step>

## Risks & Mitigations
- <Risk> — <Mitigation>

## Testing Strategy
- <Unit+integration coverage; invariants to assert; -race if applicable>

## Acceptance Criteria
- <Concrete, verifiable outcomes; no behavior change unless ADR signed‑off>

## Refs
- Refs: goal <name>; requirement <name>; spec <name>; decision <name>

## Examples (Use Judiciously)
- Unify in‑memory session stores behind a generic TTL store.
- Standardize error responses via a common envelope + mapping helpers.
- Centralize `randBytes` in a single `randutil` package.
- Centralize router building in `server/cmd/api` for cohesion.
- Introduce light DB interfaces and prepared statements via small repos.

## Anti‑Patterns To Avoid
- Big‑bang rewrites across multiple layers without tests.
- Refactors that couple to in‑flight steps or change public contracts without ADRs.
- Broad “cleanup” commits lacking traceability or acceptance criteria.

## Verification (Meta)
- Each proposal includes a clear migration path and test plan.
- Implementation only proceeds when explicitly prioritized and scheduled; otherwise remains as documentation.
