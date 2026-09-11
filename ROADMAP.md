# vRI Roadmap

> This roadmap orders learning, not just implementation. Each phase must earn
> the complexity of the next one.

## Product hypothesis

Agent use can become a compounding improvement process if a system can:

1. turn grounded experience into candidate changes to how the agent system
   works;
2. verify those changes independently on future or held-out tasks;
3. compose and route accepted methods within appropriate scopes;
4. preserve fallback, rollback, provenance, and human judgment;
5. eventually improve the mechanism that proposes and evaluates changes.

The realistic baseline is ordinary agent sessions plus Git, worklogs,
hand-written skills, and manual evaluation. Persistent traces or cleaner
multi-agent handoffs alone are not sufficient product evidence.

## Phase 0: Manual loop and executable specification

Run the proposed loop by hand on real work:

- select one opaque agent and one recurring task distribution;
- record exact system versions and episodes;
- identify experience that may justify a reusable method change;
- produce candidate skill, context, or routing changes;
- compare candidates with the baseline on unseen tasks;
- make explicit scoped-release and fallback decisions;
- run later episodes through the selected release;
- document where existing tools already suffice.

Represent the minimum records in fixtures and CLI acceptance tests before
stabilizing schemas.

**Exit condition:** at least one change derived from experience produces a
reproducible improvement on future tasks, and every required object and human
decision can be named without relying on a chat transcript.

## Phase 1: Explicit local improvement loop

Build one local, CLI-first vertical:

- register immutable service and artifact references;
- resolve an opaque BYO agent plus its improvement envelope into a
  `SystemVersion`;
- record episodes and source-linked experience;
- invoke a BYO improver with declared mutable surfaces and budget;
- store candidate deltas without overwriting the active version;
- run user-provided baseline and candidate evaluations;
- record evidence separately from a human decision;
- create a scoped release with routing and fallback;
- resolve the release for a subsequent episode;
- disable or roll back an assignment without losing lineage.

Use SQLite, Git, content-addressed files, and command adapters. A native process
runner is the baseline; reuse Herdr for session and workspace capabilities where
its interfaces fit. No Web UI or distributed runtime is required.

**Exit condition:** a complete cross-generation loop works locally and an
independent reviewer can reproduce why the later episode received its resolved
method.

## Phase 2: Agent-programmable context and active experiments

Let agents work with experience rather than merely receive summaries:

- query episodes, experience, artifacts, and evidence from Bash or Python;
- materialize provenance-linked context views;
- register reusable context strategies as candidate methods;
- measure whether released methods were selected and actually used;
- let an agent propose experiments from uncertainty, contradictions, or missing
  evidence;
- separate proposal, authorization, execution, and decision;
- run improvement jobs asynchronously without blocking ordinary use.

An RLM may use these capabilities from a persistent REPL, but vRI must also work
with agents that expose only shell, files, or a simple request API.

**Exit condition:** a context or experiment strategy learned from prior
experience transfers to held-out tasks more reliably than stored notes or a
static context package.

## Phase 3: Versioned service graph and controlled mutability

Test the broader “everything addressable as a service” hypothesis:

- define small provider contracts and capability discovery;
- record fixed, mutable, and protected components per improvement run;
- integrate external model, data, train, serve, eval, and environment services
  without implementing those systems in vRI;
- add a second vertical such as a stable training factory with a mutable data
  recipe or algorithm plug-in;
- measure single-surface changes before testing important interactions;
- bind service versions, configuration, environment, and cost to evidence.

**Exit condition:** two materially different verticals use the same system
manifest, candidate, evaluation, and release semantics without domain-specific
changes to the core.

## Phase 4: Parallel search and managed operation

Add scale only where observed workloads require it:

- bounded multi-agent candidate generation and independent critique;
- worktrees, containers, sandboxes, or remote environments through providers;
- conflict and duplicate-work detection;
- adaptive experiment budgets and cancellation;
- durable background workers and recovery;
- organization scopes, credentials, quotas, and audit policy;
- PostgreSQL and object-storage providers;
- a review UI if CLI reports no longer support judgment efficiently.

Reuse agent-runtime orchestration rather than prescribing a universal team
topology. Multi-agent work is one search strategy over the same durable
improvement protocol.

**Exit condition:** parallel or managed execution increases validated
improvement throughput without weakening attribution, evidence quality, or
human control.

## Phase 5: Governed recursive improvement

Open deeper mutation surfaces cautiously:

- version and evaluate the improver itself;
- compare improvement policies on held-out improvement tasks;
- compose several compatible releases and measure interactions;
- support joint service optimization with fixed anchor evaluations;
- admit evaluator changes only through independent meta-evaluation;
- measure persistence, transfer, cost, regression, and multi-generation gain;
- allow wider autonomous experiment budgets only when prior calibration
  supports them.

Eval-RSI is the focused research branch for weakness discovery, evaluator
repair, portfolio evolution, and protected acceptance. Its validated outputs
may become `EvalService` implementations or service versions in this phase.

**Exit condition:** a later system is demonstrably better at producing
validated future improvements, not merely better at the task used to modify it.

## Early success metrics

- Held-out and temporally later task improvement at matched cost.
- Regression and rollback rate after release.
- Fraction of accepted methods actually selected and used.
- Transfer across repositories, task families, agents, and time.
- Human judgment time per validated improvement.
- Candidate throughput versus validated-improvement throughput.
- Fraction of results with reproducible system, evaluator, and environment
  resolution.
- Ordinary-use latency or failures caused by background improvement.

## Falsification and narrowing conditions

Narrow or change the design if:

- stored experience rarely yields concrete candidate methods;
- proposed methods do not beat simple hand-written skills or worklogs on unseen
  tasks;
- scoped routing adds more complexity than value;
- provider abstraction repeatedly hides important domain semantics;
- independent evaluation dominates the workflow but cannot be made affordable;
- users prefer existing agent-native continual harnesses without needing an
  outer control layer;
- a second vertical cannot reuse the core model without extensive special
  cases.

The response to these results is not to add more orchestration or services.

## Explicitly deferred

- a custom foundation model or universal agent;
- a universal verifier;
- a new sandbox, trainer, serving stack, or cluster scheduler;
- automatic global replacement of accepted candidates;
- a mandatory RLM or agent-internal harness;
- an elaborate workflow language or agent-team topology;
- a Web UI before the review workflow demonstrates its need;
- open-ended autonomous RSI claims.
