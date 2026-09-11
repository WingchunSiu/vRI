# vRI Initial Technical Plan

> Status: implementation hypothesis for the first vertical. It intentionally
> leaves provider APIs and process topology small until a complete improvement
> loop has been tested.

## 0. Goal of the First Implementation

The first implementation must demonstrate one causal path:

```text
opaque agent performs ordinary work
  -> an episode exposes reusable experience
  -> an improver proposes a method change
  -> a verifier compares candidate and baseline
  -> a human releases the change to a limited scope
  -> a future episode resolves and uses the change
  -> the assignment can be rolled back
```

The first mutable surfaces are external skills, context strategy, and routing.
The agent's inner harness remains opaque. Multi-agent rooms, distributed
execution, managed training, and evaluator evolution are not required to prove
this path.

## 1. Design Stance

The implementation follows **hard invariants, soft workflow**.

The core is opinionated about:

- stable identities and immutable version references;
- causal links from episodes to experience to candidates;
- exact baseline, candidate, evaluator, and environment resolution;
- controlled access to fixed, mutable, and protected components;
- separation of evidence, decisions, releases, and assignments;
- atomic recording of state changes and audit events;
- budgets, human interventions, and rollback.

It is deliberately unopinionated about:

- how an agent reasons or delegates;
- how an improver searches for changes;
- whether the agent uses Bash, Python, an RLM, or another control loop;
- the order in which experiments are proposed and run;
- which verifier expresses product or research judgment.

## 2. Local-First Architecture

```text
                 Human / BYO agents
                         |
              CLI / generated skill / files
                         |
                    vRI local core
       +-----------------+------------------+
       | identities, manifests, event log   |
       | episodes, candidates, evidence     |
       | decisions, releases, assignments   |
       +-----------------+------------------+
                         |
              provider and command adapters
                         |
      Agent / Session / Workspace / Context / Eval /
          Environment / Model / Train / Serve
```

The initial system is a modular monolith and local state directory. A service
is a logical capability and version boundary, not necessarily a network
service. Short commands may operate directly on local state. A worker process
can be added for asynchronous agent runs and experiments without changing the
domain model.

## 3. Versioned Service and Provider Model

A `ServiceRef` identifies a callable capability or artifact relevant to system
behavior. At minimum it records:

- provider and logical role;
- immutable version, digest, or reproducible locator;
- configuration that can affect behavior;
- declared capabilities and access modes;
- provenance and owner.

Providers may be built in, local executables, existing CLIs, containers, or
remote APIs. The first provider contract needs only a small common shape:

```text
describe -> capabilities and schemas
resolve  -> immutable reference and effective configuration
invoke   -> run reference, outputs, logs, cost, and status
```

Optional capabilities such as `inspect`, `fork`, `mutate`, `snapshot`, or
`stream` are discovered rather than assumed.

Model, Data, Train, Serve, Eval, Environment, Harness, and Agent are semantic
roles over this generic mechanism. vRI should not build a dedicated subsystem
for each role until a concrete integration requires it.

### System manifest

A `SystemVersion` is an immutable manifest:

```toml
[system]
agent = "agent.codex@resolved-version"
environment = "environment.host@darwin-arm64"
evaluator = "eval.command@sha256:..."

[envelope]
skills = "artifact.skills@sha256:..."
context_policy = "artifact.context-policy@sha256:..."
routing = "artifact.routing@sha256:..."
```

An improvement run binds the manifest and declares controlled mutability:

```toml
[improvement]
base = "system@sha256:..."
mutable = ["envelope.skills", "envelope.context_policy"]
fixed = ["agent", "environment"]
protected = ["evaluator"]
```

All components remain encapsulated. A mutable component changes only by
producing a candidate reference; an agent does not overwrite the active
version.

## 4. Initial Data Model

Exact SQL schemas should follow an executable specification, but the first
vertical requires these records:

| Record | Minimum purpose |
| --- | --- |
| `scope` | Boundary for experience, release, and assignment |
| `service_ref` | Resolved provider, role, version, digest, configuration |
| `system_version` | Immutable manifest of component references |
| `episode` | One task using one resolved system version |
| `experience` | Typed inference linked to source episodes |
| `improvement_run` | Base system, mutable surfaces, improver, budget |
| `candidate_change` | Structured delta and resulting system version |
| `evaluation_run` | Exact candidate, baseline, tasks, evaluator, environment |
| `evidence` | Measurements, uncertainty, failures, and artifact references |
| `decision` | Human or policy judgment over evidence |
| `release` | Approved components, constraints, and fallback |
| `assignment` | Routing from a scope or task predicate to a release |
| `resolution` | Exact system version selected for an episode |
| `event` | Append-only audit record with actor and causal references |

Experience records begin with a deliberately small taxonomy such as `failure`,
`uncertainty`, `inefficiency`, `useful_method`, and `contradiction`. Free-form
details remain artifacts rather than forcing an exhaustive ontology.

Candidate changes should be structured deltas over a manifest. Skill and
context changes may point to ordinary Git or content-addressed artifacts.

## 5. Context Substrate

The first context system should make durable information addressable to an
agent without forcing all history into its prompt.

Agents may:

- list and query episodes, experience, evidence, and artifacts;
- open exact references from Bash or Python;
- materialize a temporary context view;
- write exploratory code that filters, joins, or summarizes history;
- register a reusable query, skill, or context policy as a candidate change.

The distinction between scratch state and method is important:

- a temporary summary helps one run;
- a saved transcript preserves history;
- a versioned context strategy that is selected and validated on future tasks
  is a candidate system improvement.

The first interface can be CLI output and JSON files:

```text
vri context list
vri context query
vri context open <ref>
vri context materialize
```

An RLM-capable agent may call the same primitives from a persistent Python
environment. Other agents can use their existing shell or Python tools. vRI
does not require a new REPL or own context-compaction algorithm.

Context provenance should record which references were made available and
which materialized views were created. Measuring whether an agent actually used
a method may initially require trace evidence or explicit instrumentation; mere
presence in a prompt is insufficient.

## 6. Agent, Session, and Workspace Integration

`AgentProvider` translates a task and resolved system version into the chosen
agent's interface. It may expose only invoke and observe, or optional deeper
mutation surfaces.

Herdr should be reused for the capabilities it already implements:

- start, prompt, wait, read, attach, interrupt, and resume agent sessions;
- persistent terminal state;
- worktree and workspace operations where its interfaces fit;
- event subscriptions when continuous lifecycle events are needed.

A single Herdr provider package may implement several vRI capability contracts.
Calling it a session provider does not mean discarding its other useful
features; it means that terminal and worktree semantics do not become vRI's
improvement domain model.

The first integration should prefer Herdr's stable JSON CLI. Direct socket
integration is justified only by a measured need for streaming or lower
latency. A simple native process provider remains useful as a baseline and for
tests.

Prime Agent can initially be treated like any other opaque agent. A deeper
adapter may later import its structured continual-harness edits as
`CandidateChange` records or expose its prompt, memory, skill, and subagent
surfaces as mutable capabilities. That integration is optional.

## 7. Experience-to-Method Transformation

The first improver may itself be a BYO agent invoked with:

- selected source episodes and experience;
- the base system manifest;
- the allowed mutable surfaces;
- available evaluation and environment services;
- budget and scope;
- existing releases and known regressions.

It returns one or more candidate deltas, rationale, expected outcome, and
suggested evaluation. It may also return “no justified change.”

The improver may proactively propose an experiment when uncertainty or missing
evidence is high. The core records the proposal separately from authorization
and execution. Initially a human authorizes resource-consuming work.

Refinement should run asynchronously relative to normal agent use. Failure or
latency in an improver must not block closing or starting an ordinary episode.

## 8. Evaluation, Decision, and Release

The first evaluator adapter runs arbitrary user-owned commands and records:

- exact task set and held-out status;
- baseline and candidate system versions;
- evaluator and environment references;
- raw outputs, scores, variance, cost, and failures;
- agent and human interventions.

Where practical, candidate-generation tasks and acceptance tasks are disjoint.
Important claims should include regression and cost evidence rather than only a
single best score.

A decision consumes evidence but does not directly rewrite an active system.
It may create a release containing:

- one added or updated skill;
- a task-family routing rule;
- a scope constraint;
- a cheaper default with a stronger fallback;
- shadow or comparison traffic;
- an expiration or review condition.

Assignment resolution runs before each episode and records the resulting exact
`SystemVersion`. Rollback disables or changes an assignment while preserving
the release and its evidence.

Eval-RSI remains a separate research branch. Its artifacts can later be exposed
as versioned evaluator or portfolio services. An evaluator proposed by the
optimized system cannot approve itself.

## 9. State and Files

Use SQLite for normalized current projections and an append-only event history.
A state-changing command updates projection rows and appends its event in one
transaction.

Use:

- Git for source, skill, and configuration artifacts where natural;
- content-addressed files for manifests, context views, logs, and outputs;
- URI plus digest references for large or remote objects.

A tentative local layout is:

```text
.vri/
├── vri.db
├── objects/
├── manifests/
├── runs/
└── views/
```

Human-readable reports and agent-facing views are projections. They are not
independent sources of truth and do not require a permanent shared-room,
inbox, or cursor abstraction.

## 10. CLI and Process Shape

The first interface is CLI-first and scriptable:

```text
vri init
vri system show|diff|resolve
vri episode start|finish|show
vri experience record|list
vri improve propose|run
vri candidate show|diff
vri eval run|compare
vri decide
vri release create|disable
vri assignment set|resolve
vri status
```

These names are sketches, not a public compatibility promise. A generated skill
can teach arbitrary coding agents the same surface. An SDK or MCP adapter may
wrap it after the object model proves useful.

A background worker becomes useful for proactive experiments, long evaluations,
and scheduled improvement. It should be optional in the first explicit loop and
must communicate through the same durable commands. A Web UI is deferred until
observing evidence, comparison, routing, and approvals through generated
reports proves inadequate.

## 11. Initial Stack

- Go for the durable core and CLI;
- SQLite for state and events;
- local content-addressed storage for small objects;
- Git for source-oriented artifacts;
- TOML for human-authored manifests and provider configuration;
- JSON for command and external-provider request/results;
- existing Bash and Python environments for agent-programmable context work;
- Herdr and a native process runner as early agent/session providers;
- arbitrary commands as the first evaluator adapter.

Provider boundaries should be tested in-process or through existing CLIs before
designing a plugin registry, binary ABI, or remote control protocol.

## 12. Validation Plan

The first study compares vRI with the realistic baseline of normal agent use,
Git, worklogs, hand-written skills, and manual evaluation.

It should test:

1. whether experience can produce a concrete method change rather than another
   note or summary;
2. whether the change improves held-out future tasks at matched cost;
3. whether the system routes the method only to its intended scope and records
   actual resolution;
4. whether the old method remains usable as fallback and rollback is lossless;
5. whether another human can reproduce the evidence and explain the decision;
6. whether normal agent use continues when improvement work fails or is slow.

The generic service-graph claim earns confidence only if the same manifest,
controlled-mutability, evidence, and release model later handles a second
vertical such as a stable training factory with a mutable data or algorithm
slot.

Early measurements include:

- held-out task success, regression, cost, and variance;
- fraction of experience that yields tested candidates;
- candidate acceptance and later rollback rates;
- transfer across tasks, time, and scopes;
- whether released methods were actually selected and used;
- human time spent verifying and routing improvements;
- improvement latency that leaks into the ordinary-use path.

If the vRI path does not outperform the baseline, the response is to narrow the
design rather than add orchestration features.

## References

- [Herdr](https://github.com/herdrdev/herdr)
- [Herdr agent automation](https://herdr.dev/docs/agent-automation/)
- [Prime Agent](https://github.com/PrimeIntellect-ai/prime-agent)
- [Prime Agent RLM](https://github.com/PrimeIntellect-ai/prime-agent/blob/main/packages/coding-agent/docs/rlm.md)
- [Prime Agent continual refinement](https://github.com/PrimeIntellect-ai/prime-agent/blob/main/packages/coding-agent/src/core/refinement/refinement.ts)
