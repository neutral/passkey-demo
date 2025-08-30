Use this template to extract a single Non‑Functional Requirement (NFR) from a larger body of source text and produce a standards‑compliant requirement.md for the blueprint.

## NFR Requirement Prompt Template

Generate a Non‑Functional Requirement document (requirement.md) in Markdown by extracting only the information relevant to one NFR from the provided source text. Use a crisp, technical, neutral tone; short paragraphs; present tense; active voice. Output only the document (no preamble or commentary). Follow this exact structure and field names:

1. H1: `# <R-ID> — <Short Name>` (e.g., `# R-PLAT-2 — Single Go Service Backend`)
2. H2 Metadata block with bullets:
   - `State: <Draft|Reviewed|Approved>`
   - `Date: <YYYY-MM-DD>`
   - `Owners: <names>`
   - `Type: nfr`
3. H2 Description: 2–4 bullets describing the NFR at a high level; write testably.
4. H2 Depends On: bullet list of related requirements/constraints extracted from the text.
5. H2 Scope: two bullets labeled “In-scope” and “Out-of-scope”.
6. H2 Acceptance Criteria: 4–8 bullets written as verifiable outcomes and checks.
7. H2 Flows: bullets linking relevant user flows if present (paths or titles), or state “Applies to all flows” if appropriate.
8. H2 Interfaces: bullets listing APIs, platform settings, deployment knobs, or modules relevant to this NFR.
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
- Type: nfr

## Description

- <High‑level non‑functional constraint 1>
- <High‑level non‑functional constraint 2>

## Depends On

- <Related requirements or platform constraints>

## Scope

- In-scope: <behaviors/areas enforced by this NFR>
- Out-of-scope: <exclusions>

## Acceptance Criteria

- <Criterion 1>
- <Criterion 2>

## Flows

- <Flow references or “Applies to all flows”>

## Interfaces

- <APIs/Configs/Deployment knobs or components>

## Risks

- <Risk 1>
- <Risk 2>

Refs: goal <goal-a>; requirement <req-b>; spec <spec-x>; decision <adr-y>

Pathing note (do not include in output):

- Save as `blueprint/global/<kebab-name>/requirement.md`.
