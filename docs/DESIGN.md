# vRI System Design

> Status: early design. This document defines the product and its invariants.
> [STORAGE.md](STORAGE.md) describes the implemented storage contract, and
> [ROADMAP.md](ROADMAP.md) names the next claim to test.

## 1. Product

vRI is building an **Eval Researcher System** that can perform evaluation work
and improve how it performs that work from experience.

The system includes the primary agent, its model and harness, tools, skills,
context and memory policies, environment, and any helper agents. It should be
able to:

- investigate what a user or organization should measure;
- collect evidence from real workflows, existing evaluations, and domain
  knowledge;
- construct, run, inspect, and repair tasks, environments, rubrics, and
  verifiers;
- explain model, harness, environment, and evaluation failures separately;
- assemble and maintain benchmarks when that is the useful outcome;
- retain useful experience and test changes to its own working method.

Benchmark construction is the first workload because it is important, costly,
and unusually dependent on judgment. A benchmark is one output of the system,
not the definition of the product.

Individuals and organizations use the same core. An individual may want a
small benchmark based on personal workflows for comparing new models. An
organization may use the system to create and maintain internal benchmarks,
debug evaluation failures, support release gates, or investigate product and
model weaknesses. Their data, scale, permissions, and review policies differ;
the underlying eval-research capabilities do not.

## 2. Problem

Evaluation work does not accumulate reliably. Raw sessions and logs may be
saved, but a later agent still has to rediscover where they are, decide what is
relevant, reconstruct what happened, and determine which earlier conclusions
remain valid. Asking an agent to reread all prior trajectories repeats this
cost and leaves no durable link from a claimed lesson to later behavior.

Scores are also weak evidence by themselves. A failure may reflect missing
capability, a harness budget, a broken environment, an overly strict hidden
test, an ambiguous task, or a valid answer that the verifier rejects. A pass
may reflect correct behavior, weak coverage, leaked state, or reward hacking.
The useful unit is therefore not a score but a supported claim about a
particular system, evaluation, and scope.

The product must improve on the practical baseline: a capable user prompting a
strong agent, searching local session files, running Harbor or another eval
tool, and keeping notes by hand.

## 3. First principles

### Preserve evidence before interpreting it

Native transcripts, terminal output, trajectories, diffs, test results,
screenshots, and benchmark artifacts may remain in their provider formats.
Normalizing all of them into one semantic trace would lose information and
couple vRI to fast-changing agent interfaces.

vRI adds only the structure required to locate, cite, compare, protect, and
version important material. Derived summaries and indexes are rebuildable;
they are not the source of truth.

### Let the agent choose the investigation

Evaluation research is open-ended. The agent may inspect one suspicious trace,
sample a failure cluster, ask a domain expert, start helper agents, modify a
verifier, run adversarial solutions, or stop because evidence is insufficient.
vRI supplies capabilities, budget and permission boundaries, and durable
state. It does not impose a universal state machine or agent topology.

### Separate evidence, interpretation, and decision

- **Evidence** is an observed artifact or outcome.
- **Experience** is a revisable interpretation that may help later work.
- **A method** is a reusable way of working, such as a skill, context policy,
  verifier-repair procedure, or delegation strategy.
- **A decision** authorizes an experiment, accepts a benchmark, or releases a
  method within a scope.

An authorization to run an experiment is not approval of its candidate task.
An agent's explanation is not automatically verified experience. A score is
not a release decision.

### Test transfer, not memory volume

The system has learned only when accumulated experience or a changed method
improves later work. More stored transcripts, longer context, more agents, or
more benchmark tasks do not establish continual learning.

### Reuse existing execution systems

vRI should not rebuild terminal multiplexing, agent session restore, sandbox
execution, model serving, or training infrastructure. It composes existing
systems through narrow capability boundaries and records behaviorally relevant
versions.

## 4. System boundary

```text
                         user or organization
                    intent, evidence access, policy
                                  |
                                  v
                       Eval Researcher System
              primary agent + helpers + skills + tools
                     context + memory + environment
                     /                         \
                    /                           \
          Herdr runtime                    eval backends
   sessions, panes, agents, hosts       Harbor, human review,
    prompt/read/wait/resume/events       model APIs, other tools
                    \                           /
                     \                         /
                      vRI durable substrate
          raw references, searchable experience, versions,
             exposure, evidence, decisions, and releases
```

### Herdr

Herdr is a runtime substrate for the Eval Researcher System. It can keep real
terminal processes alive, recognize agent state, resume supported native agent
sessions, start helpers, prompt and read agents, wait on agents or processes,
subscribe to events, and expose work across local and remote hosts.

Herdr does not provide the semantic experience model. Its native session
reference is a useful bridge to the owning agent's transcript store. Terminal
screen history is partial and may be disabled for privacy, so vRI must record
capture completeness and use agent-native logs when available.

The primary eval researcher may run inside Herdr and use it directly. vRI may
also support imported sessions or another runtime; Herdr is the first concrete
integration, not a mandatory user interface.

### Harbor and other eval backends

Harbor owns task packaging, sandbox execution, verifier execution, and its
native job artifacts. The eval researcher may invoke Harbor, monitor jobs,
inspect failures, patch tasks, and rerun them. vRI records the exact task and
runtime identities plus the evidence needed to interpret the result.

This distinction is important: Harbor owns an evaluation execution; the Eval
Researcher System owns the investigation and iteration around it. vRI does not
need to implement Harbor, but it must support that active loop rather than
only ingesting a final score.

## 5. Evidence and experience substrate

The substrate has four layers. They are data roles, not mandatory workflow
stages.

### Raw evidence

Raw material is stored as a content-addressed object or referenced by a stable
URI plus digest when copying is impractical. Examples include:

- native agent sessions and trajectories;
- Herdr pane snapshots and lifecycle events;
- Harbor configs, locks, results, verifier output, logs, and recordings;
- Git commits, diffs, issues, pull requests, and later fixes;
- user corrections, expert judgments, production incidents, and documents.

Unexpected artifacts must remain ingestible. Provider adapters may identify
well-known files, but an allowlist must not make an otherwise useful failed run
disappear.

### Minimal envelope

Cross-system search requires a small amount of common metadata:

```text
source and external identity
time and host
workspace or repository
task or intent when known
agent, model, harness, and environment when known
inputs, outputs, and outcome references
human interventions
permissions, retention, and completeness
```

Every field may be unknown. vRI should preserve partial evidence rather than
invent missing semantics.

### Experience

An experience item is free-form content plus a small durable envelope:

```text
body artifact
supporting or contradicting evidence references
scope in which it may apply
status: proposed, active, contested, superseded, or retired
author and version
optional actions and later outcomes
```

The body may be prose, code, a script, a query, a playbook, or a provider-native
artifact. The agent decides what deserves an experience record. vRI should not
convert every session summary into permanent memory.

### Context views

For each task, the eval researcher queries metadata, full text, and semantic or
domain-specific indexes to build a working set. It may recursively open the
underlying evidence, run code over it, or request a different view.

```text
current intent and workspace
  + relevant prior experience
  + selected raw evidence
  + current benchmark and system state
  -> temporary context view
```

Context views are disposable projections. A reusable query, skill, memory
policy, or context-building program becomes a `MethodVersion` only when it is
deliberately proposed and evaluated.

## 6. Improvement roles and loops

An improvement claim names four roles:

| Role | Meaning | Eval-researcher application |
| --- | --- | --- |
| Target | what changes | benchmark, task, verifier, or researcher method |
| Worker system | what performs the work | Eval Researcher System and helpers |
| Method | reusable way of working | curation, context, delegation, or repair policy |
| Evaluator | what judges the change | task checks, expert judgment, protected cases, later outcomes |

The first loop produces useful evaluation artifacts:

```text
intent and evidence
  -> agent investigation and execution
  -> candidate evaluation artifact
  -> validation and review
  -> scoped release or further iteration
```

The continual-learning loop uses previous work:

```text
prior episodes, findings, corrections, and outcomes
  -> retrieved experience and candidate method change
  -> later evaluation work on new source material
  -> comparison with the previous method or stateless baseline
  -> retain, narrow, or reject the change
```

This resembles a stateful-versus-stateless continual-learning evaluation: vRI
defines comparable conditions and evidence boundaries, not the internal memory
mechanism. The system may use full history, notes, retrieval, a playbook,
fine-tuning, or another method.

Recursive self-improvement is a stronger claim. It requires an accepted change
to the Eval Researcher System to improve its ability to produce future
validated changes on unexposed work. It cannot be inferred from one improved
benchmark or one successful retry.

## 7. Evaluation discipline

Every consequential claim should record:

- the practical baseline and matched budget;
- the exact target, worker, method, evaluator, and environment versions;
- which tasks, traces, answers, verifier feedback, and outcomes the improver
  could access;
- the evidence supporting and contradicting the claim;
- uncertainty, regressions, cost, supported scope, and expiration;
- the human or policy decision that followed.

Information exposure is part of the result:

| Partition | Use |
| --- | --- |
| Discovery | generate hypotheses and changes |
| Selection | choose among candidates |
| Protected acceptance | test a frozen candidate without revealing the test |
| Temporal follow-up | observe later work unavailable during construction |

Using the same case for discovery and reporting can establish a regression fix
or in-sample adaptation. It cannot establish transfer.

For benchmark work, the agent should treat quality as a vector rather than one
score: importance, intent alignment, verifier validity, environment integrity,
anti-hacking robustness, discrimination, portfolio contribution, external
relevance, reproducibility, and interpretability. The required evidence depends
on the decision the benchmark supports.

## 8. Trust and autonomy

Agent autonomy operates inside explicit boundaries:

- the user or organization supplies the top-level intent, data access, budget,
  and protected policies;
- the agent chooses the investigation, tools, helpers, and intermediate
  artifacts;
- resource-consuming actions follow configured authorization rules;
- candidate evaluators and methods cannot serve as their own acceptance proof;
- credentials and protected evaluation data stay with their owning providers;
- raw evidence, exposure, and decisions are immutable or append-only;
- releases are scoped, reversible, and distinct from rejection or requests for
  more evidence.

Only consequential transitions require durable structure. vRI need not record
every thought, terminal keystroke, or agent-to-agent message as a domain object.

## 9. Non-goals

The current design does not include:

- a foundation model or universal agent interface;
- a universal semantic trajectory schema;
- automatic conversion of all history into memory or training data;
- a new task format, sandbox, trainer, serving system, or cluster scheduler;
- a fixed multi-agent topology;
- autonomous authority over top-level organizational goals;
- online model-weight learning as a prerequisite;
- claims of open-ended RSI.

## References

- [Herdr concepts](https://herdr.dev/docs/concepts/)
- [Herdr agent automation](https://herdr.dev/docs/agent-automation/)
- [Herdr socket API](https://herdr.dev/docs/socket-api/)
- [Herdr session state and restore](https://herdr.dev/docs/session-state/)
- [Herdr multi-machine operation](https://herdr.dev/docs/connecting-machines/)
- [Harbor](https://github.com/harbor-framework/harbor)
- [RSI roundtable research notes](research-notes/2026-09-11-dwarkesh-rsi-roundtable.md)
