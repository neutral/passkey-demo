Use this template to extract a single Functional Requirement (FR) from a larger body of source text and produce a standards‑compliant requirement.md for the blueprint.

## FR Requirement Prompt Template

Generate a Functional Requirement document (requirement.md) in Markdown by extracting only the information relevant to one FR from the provided source text. Use a crisp, technical, neutral tone; short paragraphs; present tense; active voice. Output only the document (no preamble or commentary). Follow this exact structure and field names:

1. H1: `# <R-ID> — <Short Name>` (e.g., `# R-FLOW-LOGIN — Login Flow`)
2. H2 Metadata block with bullets:
   - `State: <Draft|Reviewed|Approved>`
   - `Date: <YYYY-MM-DD>`
   - `Owners: <names>`
   - `Type: fr`
3. H2 Description: 2–4 bullets describing the FR scope at a high level; write testably.
4. H2 Depends On: bullet list of related requirements/constraints extracted from the text.
5. H2 Scope: two bullets labeled “In-scope” and “Out-of-scope”.
6. H2 Acceptance Criteria: 4–8 bullets written as verifiable outcomes and checks.
7. H2 Flows: bullets linking relevant user flows if present (paths or titles).
8. H2 Interfaces: bullets listing APIs, UIs, CLIs, or methods directly exercised by this FR.
9. H2 Risks: bullets calling out key risks and pitfalls.
10. Final line: `Refs: goal <names>; requirement <ids>; spec <names>; decision <adr-names>` (omit categories you don’t have).

Additional rules and style constraints:

- Use backticks for interfaces, IDs, constants, and code-like tokens.
- Prefer precise, testable language (“must/require/enforce/verify”); avoid vague phrasing.
- Keep bullets tight; every line must be verifiable.
- Derive names in kebab‑case for folders/files; keep the human title concise.
- If exact IDs are not present, infer a consistent ID pattern from context.

Skeleton (fill in placeholders using extracted facts only):

# <R-ID> — <Short Name>

## Metadata

- State: Draft
- Date: <YYYY-MM-DD>
- Owners: <owners>
- Type: fr

## Description

- <High‑level requirement statement 1>
- <High‑level requirement statement 2>
- <High‑level requirement statement 3>

## Depends On

- <R-PLAT-x / R-SEC-... / other FRs/NFRs>

## Scope

- In-scope: <concise list of covered behaviors/areas>
- Out-of-scope: <concise list of exclusions>

## Acceptance Criteria

- <Criterion 1>
- <Criterion 2>
- <Criterion 3>

## Flows

- <path-or-title to flow, if available>

## Interfaces

- <`POST /...` or Component/View or Method signature>

## Risks

- <Risk 1>
- <Risk 2>

Refs: goal <goal-a>; requirement <req-b>; spec <spec-x>; decision <adr-y>

Pathing note (do not include in output):

- Save as `blueprint/features/<kebab-name>/requirement.md`.
