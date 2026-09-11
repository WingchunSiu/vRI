# vRI System Design

> Status: early working draft. This document records the current main thesis
> and the boundaries that appear stable enough to test. Object names, APIs, and
> deployment choices remain provisional.

## 0. Design Status

vRI began as a control plane for long-running, multi-agent research campaigns.
That work exposed useful infrastructure concerns, but persistence and
coordination alone do not constitute self-improvement. The current design
centers on the harder loop between ordinary use, method change, independent
verification, and future use.

### Relatively stable decisions

- Bring-your-own agents are replaceable providers. vRI must work when an
  agent's internal harness is opaque.
- Experience, candidate changes, evidence, decisions, releases, and assignments
  are distinct records.
- The unit being evaluated is a versioned agent system, not necessarily a model
  or one agent binary.
- Improvement is scoped and compositional. Accepting a candidate does not imply
  globally replacing every older method.
- Agents should be free to compose capabilities and propose experiments without
  following a fixed workflow.
- Evaluators, environments, agents, models, and other dependencies must be
  referenced by exact versions where they affect an improvement claim.
- vRI should reuse external runtimes and infrastructure through capability
  boundaries rather than implementing the whole AI stack.

### Current working hypotheses

- A useful first product is an outer improvement runtime around opaque agents:
  it changes skills, context strategies, tools, services, routing, and
  environments, then tests whether those changes help on later unseen tasks.
- A versioned service graph can represent both production systems and controlled
  improvement experiments. Each run declares which service slots are fixed,
  mutable, or protected.
- Context should be an agent-queryable computational substrate, not only a
  prompt assembled in advance.
- Verifier evolution is important enough to study independently, but it should
  enter the main system through a protected, versioned evaluation service.

### Still unresolved

- Whether the long-lived organizing object should be called a scope, program,
  campaign, or system family.
- Which experience-to-method transformations generalize beyond one agent,
  repository, or task family.
- How an agent should estimate uncertainty and when that should trigger an
  autonomous experiment.
- Which outcomes from ordinary use are trustworthy enough to become evaluation
  evidence.
- How much release and runtime routing vRI should own versus export to another
  deployment system.
- Whether the first useful implementation needs a background daemon or Web UI.
- Which multi-agent coordination primitives remain necessary once agent
  runtimes can manage their own subagents.

## 1. Problem

Agents produce large amounts of experience: trajectories, tool calls, failures,
human corrections, artifacts, and evaluation results. Most of that experience
does not reliably change how the system works next time.

Saving experience is not learning. If an agent mistakenly edits a generated
file, retaining the trace or writing “avoid generated files” preserves a
description of the failure. It does not establish that:

1. a reusable method changed;
2. the changed method is selected in the relevant future situation;
3. it improves behavior on tasks not used to invent it;
4. it does not cause unacceptable regressions elsewhere.

Existing agent products may provide memory, skills, traces, evaluations,
sandboxes, or orchestration. The missing system-level loop is the controlled
transformation of experience into methods, followed by independent evidence and
scoped accumulation.

## 2. Thesis

**Self-improvement means that experience changes how an agent system works in
the future, and that the change produces validated benefit beyond the
experience that generated it.**

The improvement target is the larger agent system:

```text
agent system
  = agent provider
  + skills and instructions
  + context strategy
  + tools and services
  + workspace and environment
  + routing and fallback policy
```

The agent provider may itself evolve, as Prime Agent's continual harness does,
or remain a closed black box. vRI should support both without requiring one
internal agent architecture.

The long-term recursive hypothesis is stronger: a validated system version may
also become better at proposing, testing, or selecting its own future
improvements. The near-term system should measure that possibility without
claiming open-ended RSI.

## 3. Improvement Loop

The conceptual loop is:

```text
Episode
  -> Experience
  -> Experience-to-method transformation
  -> Candidate change
  -> Independent evaluation
  -> Decision
  -> Scoped release and assignment
  -> Future Episode
```

This is a causal model and audit boundary, not a mandatory workflow engine.
Agents may inspect multiple episodes, branch, delegate, write exploratory code,
or request new experiments in any order.

### Episode

An episode is one bounded use of a resolved agent system on a task. It records
the exact system version, inputs, relevant context references, actions,
artifacts, outcomes, costs, and human interventions.

### Experience

Experience is a typed interpretation derived from one or more episodes. It may
describe a failure, uncertainty, repeated inefficiency, useful strategy,
contradiction, or opportunity. It must retain links to the source episodes.

Experience is neither a durable method nor proof that a proposed lesson is
correct.

### Experience-to-method transformation

An improver converts experience into candidate changes. It may be a human,
agent, script, search process, or combination. Candidate surfaces include:

- adding or revising a skill;
- changing context selection or context-computation strategy;
- updating routing for a task family;
- adding a tool or service;
- changing an environment, workflow, prompt, or subagent specification;
- modifying data, training, serving, or evaluation components;
- changing the improver itself.

The transformation should preserve its inputs, reasoning or predictions where
useful, budget, generated candidates, and failures. No single transformation
algorithm is part of the core.

### Active experiments

Improvement need not wait for the next user request. An agent may use its
uncertainty, conflicting evidence, expected value, or missing coverage to
propose experiments. In the first version, proposing and authorizing work
remain separate actions so resource use and autonomy are visible.

## 4. Agent Boundary and Improvement Envelope

vRI distinguishes two harness layers:

- The **inner harness** is owned by the agent provider: its reasoning loop,
  compaction, tool routing, subagent implementation, and internal memory.
- The **improvement envelope** is the versioned system around the agent:
  external skills, context surfaces, tools, services, workspace, environment,
  routing, budgets, and feedback.

An adapter declares which capabilities and mutable surfaces a provider exposes.
Examples:

| Provider type | Typical mutable surfaces visible to vRI |
| --- | --- |
| Closed agent | skills, context, tools, environment |
| CLI coding agent | instructions, skill files, workspace, tool services |
| Continual harness | prompt, memory, skills, subagent specifications |
| Training system | data recipe, algorithm, model, trainer configuration |

Deeper provider-specific improvement is optional. The portable vRI loop must
still work at the envelope level.

## 5. Versioned Service Graph

“Everything as a service” means that important components are addressable
through stable contracts and immutable references. It does not mean that every
component is a network microservice or that vRI implements all of them.

A system version is an immutable manifest of service and artifact references:

```text
SystemVersion
  model        -> ServiceRef
  data         -> ServiceRef or ArtifactRef
  train        -> ServiceRef
  serve        -> ServiceRef
  eval         -> ServiceRef
  environment  -> ServiceRef
  harness      -> ServiceRef
  agent        -> ServiceRef
  envelope     -> versioned skills, context, tools, and routing
```

The core should not require every role or hard-code domain-specific behavior.
Roles describe how generic versioned capabilities participate in a system.

### Controlled mutability

Every component remains encapsulated. An improvement run assigns one of three
access modes:

- **fixed:** callable but unchanged, providing a controlled experimental factor;
- **mutable:** forkable or replaceable through a declared extension point;
- **protected:** callable but not inspectable or changeable by the system being
  evaluated.

For example, a stable training factory may accept a mutable data recipe while
holding the model, algorithm, and environment fixed. A later run can hold data
fixed and open the algorithm slot. Joint optimization should begin only after
single-surface effects and important interactions can be measured.

Service references must capture a version or digest and the configuration that
can affect results. A mutable remote alias is not a sufficient control.

## 6. Core Records

The following are candidate core concepts. Exact schemas remain provisional.

### Scope

The boundary within which experience, policies, releases, and assignments
apply: for example a user, repository, organization, task family, or production
service.

### SystemVersion

An immutable manifest resolving an agent and all relevant improvement-envelope
and service references. Every episode and evaluation binds to one.

### Episode and Experience

An episode records what happened. Experience records what an actor inferred
from one or more episodes. Neither mutates the active system.

### ImprovementRun and CandidateChange

An improvement run binds source experience, a base system version, fixed,
mutable, and protected surfaces, an improver, objective, budget, and generated
candidates. A candidate change is a structured delta from the base manifest.

### EvaluationRun and Evidence

An evaluation run binds an exact candidate, evaluator, environment, task set,
and resource policy. Evidence preserves observations and uncertainty. Evidence
does not decide deployment.

### Decision

A human or policy judgment over evidence. It may reject, request more evidence,
approve limited use, or make a candidate eligible for release.

### Release, Assignment, and Resolution

A release packages approved capabilities and constraints. An assignment policy
selects a release for a scope and task context. Resolution records which exact
system version was used by an episode.

Promotion therefore need not replace an old workflow. It may:

- add a skill without removing existing skills;
- route one task family to a specialized method;
- enable a candidate only for a repository or team;
- retain the older method as a cheaper path or fallback;
- run a candidate in shadow or comparison mode.

Rollback changes assignment or release state; it does not erase history.

## 7. Architecture

```text
                    Humans and agents
                           |
                CLI / skills / SDK adapters
                           |
                     vRI control core
       +-------------------+-------------------+
       | identity, lineage, events, policy     |
       | episodes, experience, candidates      |
       | evidence, decisions, releases         |
       +-------------------+-------------------+
                           |
                versioned provider adapters
                           |
      Agent / Model / Data / Train / Serve / Eval /
              Environment / Workspace / Context
```

Use, improvement, and evidence/governance are separate logical concerns, not
necessarily separate processes. The initial implementation should remain a
modular local system.

### Agent-programmable capabilities

vRI should expose small, reliable primitives through Bash, skills, Python, or
other adapters. The agent decides how to compose them; vRI records the resulting
call graph and artifacts. This preserves a soft workflow while keeping
identity, budgets, permissions, provenance, and evidence hard.

### Context as a computational substrate

Context is not only a generated prompt. Episodes, artifacts, experience, and
evidence should be addressable data that an agent can search, filter, join,
summarize, or analyze with ordinary code.

An RLM-like agent may do this inside a persistent REPL. Another agent may use
Bash and Python against a mounted context store. vRI should standardize the
references and record resulting context views without requiring either agent
architecture.

A promoted context improvement is a reusable method: a query, skill, routing
rule, or context policy. A scratch summary or stored transcript is not by itself
such an improvement.

## 8. Verification and Trust

Verifier quality constrains the whole loop. vRI does not provide a universal
definition of improvement; it provides a governed way to invoke and record
tests, benchmarks, statistical comparisons, external outcomes, agent critics,
and human review.

At minimum:

- evaluation evidence is versioned and separate from decisions;
- candidates are tested on tasks not used to construct them where possible;
- regression, cost, and uncertainty remain visible;
- the target cannot silently modify protected evaluators or promotion rules;
- evaluator changes require independent meta-evaluation or human authority;
- exact inputs, environments, and interventions remain auditable;
- a released method can be disabled or routed away without deleting evidence.

Eval-RSI studies how weakness discovery and evaluation mechanisms can evolve
under these boundaries. It can later supply an `EvalService`, evaluation
portfolio, or candidate evaluator changes to the main loop.

## 9. Scope

### vRI owns

- durable identities, version manifests, lineage, and event history;
- episode, experience, candidate, evidence, decision, and release records;
- controlled-mutability and trust semantics;
- scope, assignment, resolution, and rollback semantics;
- adapters and capability discovery for external services;
- enough scheduling and budget control to run improvement experiments.

### Users and providers bring

- agents and their internal harnesses;
- models, data, trainers, serving systems, and evaluators;
- repositories, tasks, and domain judgment;
- local processes, Herdr, worktrees, containers, sandboxes, clusters, or cloud
  infrastructure.

### Non-goals

Initially vRI will not:

- build a foundation model or universal agent;
- implement every service in the AI stack;
- equate trace storage, memory, or self-critique with learning;
- prescribe one research or multi-agent workflow;
- assume a universal scalar verifier;
- automatically approve evaluator or improver changes;
- claim open-ended recursive self-improvement.

Multi-agent execution remains a useful system capability for search,
independent critique, and parallel experiments. It is not the first product
thesis and should reuse agent runtimes or session providers where possible.

## 10. First Vertical

The first discriminating implementation keeps one bring-your-own agent opaque:

1. resolve a base system version containing the agent, skills, context policy,
   tools, environment, and evaluator;
2. run ordinary tasks and record episodes;
3. turn selected experience into candidate skill, context, or routing changes;
4. evaluate candidates against the current system on held-out tasks;
5. record a human decision and create a scoped release;
6. resolve that release for a later episode and measure whether the gain
   persists;
7. disable or roll back the assignment without losing lineage.

The realistic baseline is the existing workflow of agent sessions, Git,
worklogs, hand-written skills, and manual evaluation. vRI must show a measurable
advantage over that baseline, not merely reproduce it with more objects.

## 11. Open Questions

- What is the smallest method-change representation that covers skills,
  context, routing, and workflows without becoming a generic configuration
  language?
- How should useful experience be selected without creating a noisy,
  self-reinforcing memory loop?
- How is actual use of a promoted method observed rather than inferred from its
  presence in a context window?
- Which tasks and temporal gaps are sufficient to test transfer?
- How should several compatible improvements compose, conflict, or inherit
  fallback behavior?
- What uncertainty signal justifies proactive experimentation?
- When does joint service optimization produce enough value to justify weaker
  causal attribution?
- Which Herdr and Prime Agent capabilities should be consumed directly through
  adapters?
- When do a daemon, Web UI, and managed service materially improve the loop?

These questions should be answered with vertical experiments before their
solutions become permanent abstractions.
