# vRI Roadmap

> This roadmap orders learning, not just implementation. Each phase should earn
> the complexity of the next one.

## Product hypothesis

Research teams are increasingly limited not by an agent's ability to execute an
individual task, but by the loss of context, provenance, and control across many
agent sessions and experiments.

vRI is useful if it makes a real research workflow easier to resume, inspect,
coordinate, and reproduce than the practical baseline of chat, shell, Git
worktrees, and an experiment tracker.

## Phase 0: Workflow study and executable spec

Use several real artifact-improvement projects to document:

- where research context is lost;
- where agents duplicate or conflict with work;
- which human interventions are necessary;
- what must be known to reproduce a result;
- which parts are already solved well by Git, shells, and existing trackers.

Create a small set of example campaign records by hand. These become the first
schema and CLI acceptance tests.

Build a disposable Herdr adapter spike that exercises workspace creation, agent
start, prompt, wait/read, attach, and session recovery. Use the observed
integration friction to decide whether Herdr should be the default local session
provider or an optional integration.

**Exit condition:** we can describe one painful workflow and show exactly which
vRI objects and actions would remove that pain.

## Phase 1: Local vertical slice

Build the smallest end-to-end runtime for one repository and one or more CLI
agents:

- initialize and inspect a campaign;
- expose an in-process service registry with built-in providers;
- create optional tasks without imposing a fixed research workflow;
- run an arbitrary CLI agent through Herdr or a simple local provider;
- read agent state and attach or intervene from the terminal;
- isolate code changes with Git worktrees;
- record an append-only event history in SQLite;
- checkpoint artifacts and link them to tasks and agent runs;
- expose a shared filesystem room with per-agent inboxes and cursors;
- generate a task-specific, versioned context package;
- record hypotheses, predictions, observations, and decisions;
- run user-provided verification commands;
- export a reproducible campaign report.

Use a CLI and a simple local daemon. No graphical interface is required for the
initial system.

**Exit condition:** a new agent can resume a stopped campaign without receiving
the full transcript, and a human can explain how the final artifact was made.

## Phase 2: Multi-agent research workflows

Add the coordination primitives needed by observed workloads:

- bounded parallel tasks and branches;
- structured handoffs and dependency outputs;
- shared findings, contradictions, and open questions;
- conflict detection for overlapping artifact changes;
- compare, merge, retry, and fork operations;
- per-agent and per-task budgets;
- context selection policies and context quality measurements.

Avoid prescribing one universal agent-team topology. Different schedulers should
operate on the same campaign protocol.

**Exit condition:** multiple replaceable agents can work concurrently without
making the campaign less understandable or reproducible.

## Phase 3: Environment and storage providers

Generalize execution beyond local worktrees:

- container and sandbox providers;
- remote machines and queues;
- snapshot, fork, pause, and resume capabilities where supported;
- model, training, serving, and evaluation service references;
- artifact backends for large datasets and model checkpoints;
- credential and secret boundaries.

Keep session supervision, execution environment, workspace, and context as
separate capability axes. The capability model should expose what each provider
can do without forcing all implementations into the strongest sandbox
abstraction.

**Exit condition:** the same campaign can move between local and isolated or
remote runtimes without changing its research semantics.

## Phase 4: Verification and promotion protocols

Strengthen how campaigns make improvement claims:

- versioned evaluator adapters;
- baseline and candidate comparison;
- repeated trials, variance, cost, latency, and regression evidence;
- protected or hidden evaluation boundaries;
- configurable promotion and rollback policies;
- human approval gates;
- an exportable improvement receipt linking artifacts, evidence, interventions,
  and decisions.

This phase improves verification infrastructure; it does not assume that a
general verifier exists.

**Exit condition:** an independent reviewer can reproduce the comparison and
audit why a candidate was promoted.

## Phase 5: Scaled and recursive campaigns

Only after the earlier model works locally:

- distributed scheduling and failure recovery;
- quotas, tenancy, and organization-level policy;
- large search trees and adaptive resource allocation;
- campaigns that improve agents, harnesses, evaluators, or research policies;
- reuse of an improved researcher in subsequent campaigns;
- longitudinal measurement of whether the system becomes a better improver.

**Exit condition:** scaling increases validated research throughput without
destroying evidence quality, human control, or reproducibility.

## Early success metrics

- Time for a different agent or human to resume a campaign.
- Fraction of work duplicated across agents.
- Fraction of artifact versions with reproducible provenance.
- Human time spent reconstructing state versus making research decisions.
- Rate of apparent gains that survive a stronger or held-out evaluation.
- Cost and quality difference between full-history prompts and selected context
  packages.

## Explicitly deferred

- a custom foundation model;
- a universal verifier;
- a new sandbox or cluster scheduler;
- a graphical UI;
- autonomous model-weight training as the first use case;
- elaborate social roles for agent teams;
- a workflow language before the core objects stabilize.
