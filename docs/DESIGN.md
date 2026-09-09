# vRI System Design

> Status: early working draft. The design should change as we test it against
> real research workflows.

## 1. Problem

Agents can already modify artifacts and run experiments. What is still fragile
is the research process around them.

A useful research system must preserve enough state for a person or a different
agent to answer:

1. What are we trying to improve?
2. What do we currently believe, and why?
3. What changed in this candidate?
4. Which experiments were run, under what conditions?
5. What failed or contradicted our expectations?
6. Should we continue, branch, revert, or ask for human judgment?

Chats, terminals, Git commits, and experiment trackers each preserve part of
this state, but none is the shared control plane for the whole process.

## 2. Thesis

**Agents are ephemeral workers. Research state is the durable product.**

vRI should let any compatible agent work on a research campaign while keeping
the campaign understandable, resumable, and reproducible by humans and other
agents.

One common improvement loop is:

```text
goal
  -> hypothesis
  -> experiment
  -> observation
  -> decision
  -> new artifact or next hypothesis
```

This is an observation model, not an enforced workflow. An agent may reorder,
skip, combine, or invent steps. vRI should provide flexible capabilities while
preserving hard invariants around identity, lineage, causality, permissions,
budgets, and durable state.

The work may be run by one agent, multiple agents, a human, or any mixture of
them. Human interventions are part of the recorded history, not invisible side
effects.

## 3. Scope

### vRI owns

- campaign state and its event history;
- artifact identity, versions, and lineage;
- agent tasks, sessions, and handoffs;
- context assembly for each task;
- experiment metadata and evidence;
- human intervention and approval points;
- budgets, stopping conditions, promotion, and rollback semantics;
- adapters connecting external agents, runtimes, and evaluators.
- stable capability contracts between the modules in the durable workflow.

### Users bring

- agents such as coding or research agents;
- model endpoints;
- source repositories and other artifacts;
- tests, evaluators, human review, or other verifiers;
- local processes, worktrees, containers, sandboxes, clusters, or training
  infrastructure.

### Non-goals

At least initially, vRI will not:

- build a new foundation model or general-purpose coding agent;
- decide universally what counts as an improvement;
- replace training frameworks, experiment trackers, or sandbox providers;
- require distributed infrastructure for local workflows;
- claim fully autonomous recursive self-improvement.

## 4. Core Objects

### Campaign

A long-lived improvement effort with a goal, constraints, budget, participants,
and stopping conditions. A campaign is a graph of work, not one chat session.

### Research state

The current structured view of the campaign:

- goals and constraints;
- claims, hypotheses, and confidence;
- known evidence and contradictions;
- open questions and next actions;
- relevant artifacts and experiments;
- decisions and their rationale.

Research state is derived from an append-only event history so that it can be
inspected and rebuilt.

### Task and agent run

A task is a bounded unit of research work. An agent run is one attempt to carry
it out with a specific agent, model, context package, runtime, and budget.

Agent runs may be started, observed, interrupted, resumed, retried, or forked.

### Artifact

Anything the campaign may improve or depend on, including code, prompts, data,
evaluations, harnesses, configurations, reports, and model checkpoints.

Every artifact version should have stable identity, parentage, provenance, and
a reproducible reference to its contents.

### Experiment

A declared comparison or measurement containing:

- the hypothesis and prediction made before the run;
- inputs and artifact versions;
- execution environment and configuration;
- raw outputs and observations;
- costs, failures, and variance where applicable.

### Evidence and decision

Evidence is an observation produced by an evaluator, test, human, or agent. A
decision records how that evidence changed the campaign: continue, reject,
branch, merge, promote, revert, or escalate to a human.

Evidence and decisions must remain separate. A score does not promote a
candidate by itself.

## 5. Architecture

```text
                  Human / agents
                        |
              CLI / API / shared volume
                        |
                     vRI core
          +-------------+----------------+
          | Research state and event log |
          | Service registry and policy  |
          +-------------+----------------+
                        |
       +----------------+-------------------+
       |                |                   |
  Agent/session     Environment         Durable workflow
     services         services              services
       |                |                   |
  Codex / Claude   host / Docker /      context / artifact /
  Herdr / native   sandbox / remote     eval / decision
```

### 5.1 Campaign control plane

The control plane owns durable campaign state, lifecycle, budgets, policies, and
events. It must not encode a mandatory research state machine. Agents and
plugins decide how research proceeds using the capabilities exposed by vRI.

The initial implementation is a modular monolith. A service is a logical
capability boundary, not necessarily a network service or independent process.
Deployment topology may change later without changing research semantics.

### 5.2 Composable execution stack

An agent run is composed from independent dimensions:

- an agent or harness, such as Codex, Claude Code, or a custom CLI;
- a session supervisor, such as Herdr or a native process supervisor;
- an execution environment, such as the host, Docker, a microVM, or a remote
  machine;
- a workspace, such as a directory, Git worktree, snapshot, or mounted volume;
- a context provider, such as a shared filesystem or generated context package.

These dimensions are composable, not mutually exclusive. For example, Herdr may
supervise an agent running on the host in a Git worktree, or a process whose
execution environment is a Docker container.

An agent adapter translates a common task request into a specific agent's
interface. A session service exposes start, send, wait, read, attach, resume,
interrupt, and terminate. An environment service exposes create, execute,
snapshot, fork, and destroy when those capabilities exist.

### 5.3 Context manager

Context management is not merely conversation summarization. Its job is to
compile a task-specific context package from durable research state.

A context package may include:

- the task and acceptance criteria;
- relevant hypotheses, evidence, and unresolved contradictions;
- selected artifact versions and diffs;
- outputs from prerequisite tasks;
- constraints, budget, tools, and reporting requirements.

The local implementation uses a shared context volume as a durable group room.
Messages, findings, decisions, and artifact references are available to every
authorized participant. Each agent also has an inbox, a cursor into the event
stream, and a versioned context snapshot for its run.

The full campaign history remains queryable without being injected into every
prompt. Agents receive a relevant slice by default but may search or read the
shared filesystem themselves. Context packages are versioned inputs to agent
runs so their effect can be compared.

### 5.4 Artifact workspace

The artifact layer provides version, diff, branch, merge, checkpoint, and
rollback operations. Git worktrees are a strong first backend for source code;
other artifact types can use content-addressed storage and external references.

The system records relationships between artifact versions, experiments, and
the agent or human actions that produced them.

### 5.5 Agent coordination

Agents coordinate through a durable shared room containing tasks, artifacts,
observations, findings, decisions, and messages. Free-form agent-to-agent chat
can exist, but it must not be the only record of important work.

The scheduler should initially support simple dependencies and bounded parallel
work. Complex autonomous team topologies can wait until real workloads require
them.

### 5.6 Verification interface

vRI does not solve the general verifier problem. It provides a standard way to
invoke and record whatever verification the user trusts:

- deterministic tests;
- benchmark or evaluation suites;
- statistical comparisons;
- external services;
- agent critics;
- human review.

Evaluators are themselves versioned dependencies. A mutable evaluator cannot be
treated as independent proof of improvement without a separate, trusted check.
Promotion policies may require particular evidence or explicit human approval.

### 5.7 Human control

Humans must be able to inspect the campaign at multiple levels:

- see active tasks, agents, budgets, and recent events;
- attach to a live agent session;
- add context or redirect a task;
- pause, stop, retry, or fork work;
- compare artifact versions and evidence;
- approve or reject important decisions.

Every intervention is recorded so the claimed level of autonomy remains
auditable.

### 5.8 Storage

The initial implementation can use:

- SQLite for current projections and an append-only event history, updated in
  the same transaction;
- Git and worktrees for code artifacts;
- content-addressed local storage for context packages, logs, and outputs;
- a human-readable shared filesystem projected from durable state;
- explicit external references for large datasets or model weights.

The database is the canonical coordination state. Agents may freely modify
ordinary workspace files, but changes that enter research history should be
registered through vRI so provenance is recorded. Storage interfaces should be
replaceable without making distributed storage a requirement for the local
version.

## 6. Runtime Composition and Isolation

Session supervision, workspace management, and isolation are separate concerns.
Herdr is a possible session supervisor and workspace integration; it is not an
alternative to a host or Docker execution environment.

Isolation is a workload property, not the central abstraction:

- A worktree may be sufficient for trusted code changes.
- A local process may be sufficient for analysis or documentation.
- Containers may be appropriate for dependency conflicts.
- A sandbox or microVM may be necessary for untrusted code, model training, or
  reproducible system environments.
- Snapshot and fork become valuable when environment setup is expensive or a
  campaign explores many branches from the same state.

vRI should define small capability interfaces and adopt existing providers
rather than building a specialized sandbox before a workload proves it
necessary. The initial technical choices are described in
[TECHNICAL_PLAN.md](TECHNICAL_PLAN.md).

## 7. Trust Boundaries

The researcher being evaluated must not silently control every source of truth.
Campaign policies should be able to protect:

- hidden or held-out evaluations;
- evaluator implementations and versions;
- promotion rules;
- credentials and private data;
- resource and spending limits;
- immutable provenance records.

These boundaries can be enforced by different runtime backends, but their
meaning belongs in the vRI control plane.

## 8. Open Questions

- What is the smallest structured research state that agents consistently use?
- When does structured handoff beat simply giving an agent the full transcript?
- Which coordination primitives reduce duplicate work without adding overhead?
- How should conflicting observations and beliefs be represented?
- Which artifact types need native support after source code?
- How much policy should be declarative versus implemented by an agent?
- What evidence is sufficient for automatic promotion in real workflows?
- At what scale do worktrees stop being sufficient and stronger isolation pays
  for itself?

These are hypotheses for experiments, not assumptions to encode permanently.
