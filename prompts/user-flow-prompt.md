Use this template to generate clear, consistent user flow documents in any project. Paste it into your tool to produce Markdown flows that follow the structure and tone below.

## User Flow Prompt Template
Generate a canonical user flow document in Markdown. Use a crisp, technical, neutral tone; short paragraphs; present tense; active voice. Output only the document (no preamble or commentary). Follow this exact structure:

1) H3 heading: <Flow Title>
2) Refs line: Refs: requirement <ids>; spec/design <names>; decision/ADR <names>; goal/objective <names>; ticket <id>  (omit categories you don’t have)
3) Goal: 1–2 sentences stating the objective and scope.
4) Flow Summary: 1–3 sentences summarizing the end‑to‑end flow.
5) Actors: list explicit roles (e.g., User, Client, Server, Authenticator, Worker).
6) Preconditions: environment/session prerequisites and gating conditions.
7) Step-by-step: numbered steps; each step starts with the actor (e.g., Client, Server, UI, Service, Worker, Authenticator); include specific actions, inputs/outputs, and interfaces like `POST /...`, `GET /...`, `myMethod(args)`, CLI commands, or UI interactions; keep steps verifiable and focused.
8) Request/Response Examples: minimal JSON for each client→server step, with field names and encodings.
9) Important Details: concise bullets with normative rules, algorithms, constants, data contracts, id/nonce/counter policies, concurrency and ordering, retries/timeouts, and guardrails.
10) Security Notes: bullets covering authentication/authorization, input validation, origin/host checks, encryption, signature/nonce/counter requirements, privacy considerations.
11) Errors and Observability: bullets listing failure cases, detection/validation points, returned status codes/messages, retry behavior, and what to log/measure/trace.
12) Postconditions: durable state and invariants after success; counters advanced; sessions set.
13) Outputs: bullets describing success results and state transitions (IDs, records created/updated, counters advanced, side effects).
14) Verification: short manual or curl sequence to validate end‑to‑end.
15) Separator line: a single line containing `---` at the end.

Style rules:
- Use backticks for interfaces, IDs, constants, and code-like tokens.
- Prefer precise, testable language (“must/required/enforce/verify”); avoid vague phrasing.
- Keep steps actor-centric and actionable; one main action per step.
- Keep bullets tight; no fluff; every line must be verifiable.
- If data models are relevant, name fields and formats explicitly.

Tone and micro‑style constraints:
- Sentence length: 10–25 words; avoid run‑ons and stacked clauses.
- Tense/voice: present tense, active voice (“Server verifies…”, not “will be verified”).
- Step format: start each step with the Actor name, then the action.
- Interfaces: include exact method and path (e.g., `POST /v1/items`), plus key fields and encodings.
- Outputs: state what is returned or persisted for each client→server step.
- Terminology: use consistent entity names; avoid pronouns without clear antecedents.
- Normative language: use “must/require/enforce/verify/reject” for rules; avoid marketing adjectives.

Step-by-step micro‑pattern (apply to every step):
- Pattern: `<Actor> <action> via `<method path>` with <fields/encodings>; receives <result/fields>`.
- If additional sub-results or validations are needed, add 1–3 sub‑bullets under the step.
- One main action per step; keep 1–2 sentences per step before sub‑bullets.

Good vs. bad examples (for style anchoring; do not include in output):
- Good: `Client requests options via POST /auth/login/options with { origin, rpId }; receives { challenge, rpId, allowCredentials }.`
- Bad: `The server sends a challenge to the user so they can login.` (vague actor, no interface, no fields)
- Good: `Server verifies signature; updates signCount; returns 200.`
- Bad: `Everything is checked and saved.` (non‑verifiable)

Skeleton (fill in placeholders):
### <Flow Title>

Refs: requirement <req-a, req-b>; spec/design <doc-x>; decision/ADR <adr-y>; goal/objective <goal-z>; ticket <ABC-123>

Goal: <one or two sentences stating objective and scope>

Flow Summary: <one to three sentences summarizing the end-to-end flow>

Actors:
- <User>
- <Client>
- <Server>
- <Authenticator/Service/Worker>

Preconditions:
- <Environment/session prerequisites and gating conditions>

Step-by-step:
1. <Actor> requests/produces <artifact> via `POST /...` with <key params>; receives <options/result>.
2. <Actor> validates/derives <data>; stores <state>; returns <options/result>.
3. <Actor> invokes `<Interface or Method>` with <inputs>; obtains <fields>.
4. <Actor> submits <fields> to `.../finish` (or equivalent).
5. <Actor> verifies <challenge/origin/flags/auth/constraints>; updates <state>; returns <result>.

Request/Response Examples:
- Request: `POST /.../options` { <example JSON> }
- Response: 200 { <example JSON> }
- Request: `POST /.../finish` { <example JSON> }
- Response: 200/201 { <example JSON> }

Important Details:
- <Normative rule or policy>
- <Algorithm/constant/derivation>
- <Data contract: fields, formats, encoding>
- <Concurrency/order/retry/timeout semantics>

Security Notes:
- <AuthZ/AuthN requirements; input validation; origin/host checks>
- <Nonce/counter/signature requirements; replay/downgrade protections>
- <PII/secret handling; encryption at rest/in transit>

Errors and Observability:
- <Fail case>: <how detected/where> → <status/code/message>; retry: <yes/no/backoff>.
- Metrics: <key counters/latency/error rates>; Logs: <fields to log>; Traces: <spans/attributes>.

Postconditions:
- <Durable state and invariants; counters advanced; sessions set>

Outputs:
- <Success code/result>; IDs: <e.g., record_id, tx_id>; State: <what changed>; Side effects: <notifications/storage/queues>.

Verification:
- <Manual or curl sequence demonstrating end-to-end validation>

---

Quality checklist (author must self‑validate; do not include in output):
- Every step starts with an Actor and includes a concrete interface or method.
- Each client→server step specifies key request fields, encodings, and returned fields.
- Errors and Observability list at least three meaningful failure cases with status codes.
- Postconditions describe durable state and IDs; Verification provides an executable validation path.
- Tone is neutral and technical; no future tense; sentences 10–25 words; consistent terms.
