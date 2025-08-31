# Technical Spec Prompt

## Purpose
Use this template to extract the Technical Specification for a single requirement from a larger body of text and produce a standards‑compliant `_specs/spec.md`.

## Technical Spec Prompt Template

Generate a technical spec document (spec.md) in Markdown for one requirement by extracting only the relevant implementation details from the provided source text. Use a crisp, technical, neutral tone; short paragraphs; present tense; active voice. Output only the document (no preamble or commentary). Follow this exact structure and field names:

1. H1: `# <R-ID> — <Short Name> — Technical Spec`
2. H2 Metadata block with bullets:
   - `Status: <Draft|Reviewed|Approved>`
   - `Date: <YYYY-MM-DD>`
   - `Owners: <names>`
3. H2 Overview: 2–4 bullets summarizing what this spec covers.
4. H2 Interfaces: list concrete APIs, components, methods, or CLI; include methods and paths like `POST /v1/...`.
5. H2 Data / Models: name entities, fields, types, and encodings (e.g., base64url, CBOR).
6. H2 Algorithms: bullets for derivations, transformations, and verification procedures; include minimal code or formula blocks when necessary.
7. H2 Security / Privacy: bullets for authn/authz, origin/host checks, counters, nonces, signature policies, PII handling.
8. H2 Errors / Observability: status codes, failure cases, logging/metrics/tracing.
9. H2 Testing Strategy: outline unit/E2E/manual checks specific to this spec.
10. H2 Open Questions: explicit unknowns or decisions deferred.
11. Final line: `Refs: requirement <id-or-name>; decision <adr-names>; spec/design <related-docs>; goal <names>` (omit categories you don’t have).

Additional rules and style constraints:

- Use backticks for interfaces, IDs, constants, and code-like tokens.
- Prefer precise, testable language (“must/require/enforce/verify”); avoid vague phrasing.
- Keep bullets tight; every line must be verifiable.
- Use small fenced code blocks for algorithms when helpful (e.g., hash anchors, SQL, CDDL).

Skeleton (fill with extracted facts only):

# <R-ID> — <Short Name> — Technical Spec

## Metadata

- Status: Draft
- Date: <YYYY-MM-DD>
- Owners: <owners>

## Overview

- <Scope statement 1>
- <Scope statement 2>

## Interfaces

- <`POST /...` or Component/Method>

## Data / Models

- <Entity/fields/encodings>

## Algorithms

- <Algorithm/derivation/validation rule>

```text
<optional formula/code>
```

## Security / Privacy

- <Rule/policy>

## Errors / Observability

- <Failure → status/code/message; metrics/logging>

## Testing Strategy

- <Unit/E2E/manual checks>

## Open Questions

- <Unknowns to resolve>

Refs: requirement <req-a>; decision <adr-x>; spec/design <doc-y>; goal <goal-z>

Pathing note (do not include in output):

- Save as `blueprint/(features|global)/<kebab-name>/_specs/spec.md`.
