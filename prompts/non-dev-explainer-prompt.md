# Non‑Developer Explainer Prompt

## Purpose
Create a clear, non‑developer explainer for a specific subsystem or technology used in the project (e.g., SQLite, cookies, CORS), aimed at technical collaborators who don’t write code. The goal is to describe what it is, why we use it, what data/concerns it involves, how it operates in our app, and practical operating guidance — without diving into source code.

## Inputs
- Requirement ID/Name: e.g., `R-PLAT-3 — SQLite Persistence`.
- Subsystem/Technology: concise label (e.g., SQLite, Sessions, CORS).
- Project context: short sentence on why this exists in the app.

## Output
- A single Markdown explainer placed under the relevant requirement’s `_specs/` folder.
- Filename suggestion: `<subsystem>-usage-explainer.md` (kebab‑case).
- No preamble; just the document.

## Steps
1) Scope and audience: state who this is for and what it covers.
2) What it is: 1–2 bullets in accessible language (no heavy jargon).
3) Why we use it: benefits tied to project needs (simplicity, performance, security, DX).
4) What we store/handle: list key data or responsibilities in domain terms (avoid code types where possible).
5) How it works here: explain runtime behavior in this app (where it lives, how it’s created/used, file paths/URLs if relevant).
6) Performance & safety settings: call out modes, knobs, or policies that affect behavior.
7) Security & privacy: what we log, what we don’t, and how protection is achieved in practice.
8) Operating it day‑to‑day: start/stop, reset, backup/restore, common actions.
9) Limitations & when to upgrade: boundaries and upgrade paths.
10) Errors & observability: how issues surface (status codes/messages), what to check in logs/metrics.
11) Glossary: define minimal terms users will see in UI/docs.
12) Refs: link back to requirement(s), goals, and ADRs.

## Style
- Audience: technical but not writing code; avoid source‑level details.
- Tone: crisp, neutral, practical; present tense, active voice.
- Bullets over paragraphs; every line should carry distinct, useful info.
- Use plain terms first; include specific names only when useful (e.g., `server/demo.db`).
- Avoid code blocks unless showing file paths or CLI invocations that improve clarity.

## Skeleton

### <R-ID> — <Subsystem> Usage Explainer (Non‑Developer)

## Purpose & Audience
- <Who this is for; what it covers>

## What <Subsystem> Is
- <Simple definition>

## Why We Use It Here
- <Benefits tied to project goals>

## What We Store / Handle
- <Key data or responsibilities>

## How It Works in This App
- <Runtime behavior, where it lives, how it’s used>

## Performance & Safety Settings
- <Modes/knobs/policies that matter>


## Security & Privacy
- <Logging, protection, omissions>

## Operating It Day‑to‑Day
- <Start/stop/reset/backup/restore basics>

## Limitations & When to Upgrade
- <Boundaries and upgrade paths>

## Errors & Observability
- <How issues surface; what to check>

## Glossary
- <Term: simple definition>

## Refs
- Refs: requirement <id>; goal <name>; decision <adr>
