# vRI Roadmap

> This roadmap orders evidence. A phase is complete only when its user path and
> research claim have been checked; implementation activity alone is not an
> exit condition.

## Product goal

Build an Eval Researcher System that can perform high-quality evaluation work
and improve how it performs that work from experience.

The first workload is benchmark construction and maintenance. The same system
should serve an individual building a benchmark from personal workflows and an
organization building internal evaluations from product requirements,
production failures, domain expertise, and existing suites.

The practical baseline is direct prompting of a strong agent plus native
session history, Git, Harbor or another eval backend, and manual notes.

## Phase 0: Grounded local study and storage skeleton

Status: **review-ready, not complete**.

A local, intentionally untracked study demonstrates one manual path:

```text
real verifier repair
  -> candidate Harbor task
  -> oracle, nop, and two model runs
  -> trajectory-level failure attribution
  -> benchmark candidate and review packet
```

This establishes that real work can produce a runnable task and that trajectory
inspection can explain a score that would otherwise be misleading. It does not
establish that the task represents the user's desired benchmark distribution,
that the comparison is statistically stable, or that vRI produced the
attribution autonomously. User review remains pending.

The Go walking skeleton stores content-addressed objects and typed records,
ingests selected Harbor artifacts, and exposes basic run, review, and release
commands. It was extracted from that study but is not yet the Eval Researcher
System.

**Exit condition:** record the user's actual benchmark decision in the local
study and preserve it as an approved, revised, or rejected case. The
implementation gaps below remain Phase 1 work regardless of that decision.

## Phase 1: Trustworthy experience and evidence substrate

Make the current local core safe enough to support a real agent:

- bind each run to the task identity captured by Harbor at execution time;
- ingest incomplete and infrastructure-failed jobs without requiring a
  successful verifier directory;
- retain arbitrary raw artifacts or stable URI-plus-digest references instead
  of relying on a closed artifact allowlist;
- represent multiple trials and distributions such as pass@k without
  overwriting results;
- keep experiment authorization separate from candidate selection and release;
- exclude generated files from task packages and content identities;
- add full-text and metadata search over source material and derived experience;
- materialize task-specific context views while preserving links to raw
  evidence.

Keep the physical design small: content-addressed files, SQLite records and
relations, and rebuildable indexes. Do not introduce a universal trace schema
or one table per domain noun.

**Exit condition:** a copied or moved store can reproduce a representative
task, run identities, evidence, review state, and decision without relying on
mutable source directories.

## Phase 2: Herdr-managed eval researcher vertical

Run the primary eval researcher inside Herdr and connect it to the durable
substrate and Harbor:

- address the runtime by host, Herdr session, workspace, pane, agent identity,
  and native agent session reference;
- use native Codex, Kimi, Claude, or other transcript stores when available;
  record terminal-only capture as partial;
- let the researcher query relevant prior evidence and experience rather than
  rereading all sessions;
- let it start and coordinate helper agents, run ordinary processes, wait on
  lifecycle events, inspect outputs, and resume long-lived work;
- let it invoke Harbor, inspect failed trials, modify tasks or verifiers, and
  rerun without leaving the research loop;
- retain only consequential inputs, outputs, findings, and decisions; do not
  turn every pane event or agent message into a domain record.

Exercise two uses of the same agent: one personal benchmark task and one
organization-shaped task using product or domain evidence. The second case
tests permissions and review, not a separate product architecture.

**Exit condition:** the system completes a useful eval investigation with less
manual history navigation and coordination than direct prompting plus Harbor,
and an independent reviewer can trace the important conclusion to raw evidence.

## Phase 3: Continual learning of evaluation work

Test whether accumulated experience changes later performance:

- retain supported findings, repair procedures, context queries, delegation
  strategies, and benchmark-design judgments with scope and status;
- compare a stateful researcher with the same system under a stateless or prior
  method condition;
- freeze a candidate method before evaluation on new source material;
- measure task quality, attribution accuracy, expert correction, time, and
  total cost rather than memory volume;
- use later benchmark defects, user decisions, and real workflow outcomes to
  support, narrow, contradict, or retire experience;
- maintain benchmark versions as models saturate tasks and product
  distributions change.

**Exit condition:** retained experience or a changed method improves later eval
work on source material not used to produce or select it, at a matched budget.

## Phase 4: Governed recursive improvement

Let the Eval Researcher System propose changes to its own skills, context and
memory policies, tools, harness, delegation strategy, and curation method.

- evaluate each change against the current strong system and matched extra
  attempts;
- keep acceptance evidence protected from the system that produced the change;
- release changes to narrow scopes with fallback and expiration;
- measure whether a changed system becomes better at producing future validated
  changes, not merely better on the task that triggered it;
- add organization-level access control, remote workers, budgets, and review
  only when the working loop requires them.

**Exit condition:** a later researcher system improves the rate, quality, or
cost of future validated evaluation improvements on unexposed work.

## Phase 5: Test the general vRI hypothesis

Only after the eval-researcher vertical works, try a materially different
target such as a software method, data recipe, or training algorithm. Extract
shared service, exposure, evidence, and release contracts from the two working
applications instead of generalizing the eval workflow in advance.

**Exit condition:** the second target reuses the same small contracts without
forcing domain behavior into the core.

## Measurements

- time and user effort to find relevant prior evidence;
- fraction of important conclusions traceable to sufficient artifacts;
- task, verifier, environment, and attribution defect rates;
- expert-rated importance and portfolio contribution of selected tasks;
- predictive value on held-out or temporally later real work;
- human review time per accepted, useful evaluation;
- gain from retained experience or a new method over the stateless/current
  baseline at matched cost;
- latency or failures imposed on ordinary agent use.

## Narrow or stop if

- indexing and retrieval do not beat direct agent search over native sessions;
- the researcher cannot choose useful directions without the user specifying
  nearly the full benchmark;
- Harbor plus ordinary scripting reaches the same outcome with comparable
  effort and reliability;
- retained experience does not improve later eval work;
- human review requires experts to redo nearly all agent work;
- the evidence model discards important provider-native information or forces
  domains into unnatural workflows;
- a second target cannot reuse the core without extensive special cases.
