# vRI Initial Technical Plan

> Status: implementation plan for the first useful Eval Researcher System.
> [DESIGN.md](DESIGN.md) defines the product and invariants.
> [CORE_DESIGN.md](CORE_DESIGN.md) documents the smaller storage skeleton that
> exists today.

## 1. Claim to test

The next implementation should show that an eval researcher can use indexed
experience, Herdr, and Harbor to complete one evaluation investigation with
less manual history navigation and coordination than a user directly prompting
an agent and running Harbor.

```text
high-level evaluation intent
  -> retrieve relevant real work and prior experience
  -> investigate with tools and helper agents
  -> build, run, inspect, and repair an evaluation
  -> produce a supported finding or benchmark candidate
  -> retain useful experience for a later investigation
```

This claim does not require model training, a Web UI, a universal trajectory
format, or proof of recursive self-improvement.

## 2. User path

The user keeps their normal agent interfaces and opts selected data into vRI.
They may start the eval researcher explicitly or leave it in a persistent Herdr
workspace.

The first path is:

1. provide an evaluation intent, source scope, budget, and permissions;
2. let the researcher query relevant sessions, repositories, Harbor jobs, and
   prior experience;
3. review a proposed direction only when the choice is consequential or
   uncertain;
4. let the researcher run and iterate with Harbor and helper agents;
5. review one compact evidence packet and any benchmark or method change;
6. approve, reject, narrow, or request more evidence.

The system must distinguish permission to spend resources from approval of a
task, benchmark, or method.

## 3. Local architecture

```text
                         Eval Researcher Agent
                     skills, tools, context policy
                                  |
                +-----------------+-----------------+
                |                                   |
          Herdr adapter                       vRI context API
     agents, panes, hosts, events       query, open, materialize,
       prompt, read, wait, resume          retain experience
                |                                   |
                +-----------------+-----------------+
                                  |
                         local vRI substrate
                 SQLite + content-addressed files
                    + rebuildable search indexes
                                  |
                       Harbor and other tools
```

Start as a local modular monolith. The primary researcher is a normal supported
agent running in a Herdr pane. Short vRI commands use the local store directly.
A worker process is justified only when event subscriptions, long-running work,
or scheduled triage require it.

## 4. Experience and evidence substrate

### Raw artifacts

Accept any regular file or directory as an artifact. Store small material by
content digest. Reference large or provider-owned material by URI plus digest,
size, media type, provider, permissions, and retention policy.

Provider adapters may expose well-known artifacts, but ingest must also retain
unexpected files and incomplete runs. A failed Harbor trial with no verifier
directory is still useful evidence.

### Minimal records

Keep the current `objects`, `records`, and `relations` tables until a concrete
query or invariant requires another table. Treat record types as conventions,
not a closed ontology.

The first agent-facing concepts are:

- `source`: an external session, repository state, document, or run;
- `episode`: a bounded attempt with known system and outcome references;
- `experience`: free-form content linked to evidence, scope, status, and
  version;
- `method`: a reusable skill, query, context policy, repair procedure, or
  delegation strategy;
- `claim`: a statement linked to supporting and contradicting evidence;
- `decision`: authorization, acceptance, rejection, or request for more
  evidence;
- `release`: an approved benchmark or method version within a scope.

Payloads remain extensible JSON and artifact references. Missing fields stay
unknown. Do not parse provider-native data unless a query or validation check
needs it.

### Search and context

Build rebuildable indexes over metadata and extracted text. Add semantic search
only after measuring whether metadata and full-text retrieval miss relevant
experience.

Expose a small agent surface:

```text
vri context query <query> [--scope ...]
vri context open <ref>
vri context materialize <ref...> --out <dir>
vri experience propose --body <file> --evidence <ref...>
vri experience revise <id> --status <status> [--body <file>]
```

`materialize` creates a disposable directory containing selected content and a
manifest back to source references. The agent may edit scratch files inside
that directory; they become durable only through an explicit import or
experience proposal.

## 5. Herdr adapter

Use Herdr as the first managed runtime, not merely as a source importer.

### Identity

Address live work with:

```text
host reference
Herdr server or named session
workspace, pane, and agent identity
native agent session reference when available
```

Herdr IDs are scoped to one server, so every durable reference includes the
host and session namespace. Runtime state and native conversation identity are
separate.

### Capabilities

The adapter should support only capabilities Herdr reports:

- list and inspect agents, workspaces, panes, and foreground directories;
- start a supported agent in an existing pane;
- prompt, read, and wait on an agent;
- split panes and run ordinary commands such as Harbor or tests;
- subscribe to agent status, pane output, and process lifecycle events;
- resume supported native sessions after a server restart;
- operate on the intended local or remote host.

The eval researcher may use the Herdr CLI or skill directly. vRI should not
wrap every Herdr command before the first vertical requires it.

### Capture

Use Herdr's integration-reported native session identity to locate the owning
agent's transcript when available. Record terminal snapshots and screen history
as partial evidence. Do not enable persistent pane history automatically; it
may contain prompts, command output, credentials, or private data.

Record consequential coordination only: delegated task, input refs, resulting
artifact or finding, agent identity, and outcome. Do not persist every
agent-to-agent message as a domain record.

## 6. Harbor adapter

Harbor remains the first evaluation backend. Correct the current adapter before
adding another backend:

- read the task identity and Harbor version from the job lock captured at run
  time instead of hashing a mutable task path during ingest;
- preserve job config, lock, result, all trial identities, rewards,
  exceptions, verifier artifacts, trajectories, and useful logs;
- ingest partial and failed trials;
- represent repeated trials without overwriting them and compute requested
  summaries such as pass@k as projections;
- separate Harbor's native digest from any vRI package digest;
- make a stored job self-contained or explicitly reference durable remote
  artifacts so release does not depend on its original directory;
- let the researcher inspect and rerun jobs through ordinary Harbor commands.

Do not build a second sandbox or task format. A later backend adapter is
justified only by a real second use.

## 7. Eval researcher workspace

Give the primary agent a normal filesystem workspace with:

```text
intent.md                 # user or organization objective and boundaries
context/                  # materialized evidence and experience
work/                     # scratch analysis, scripts, and candidate artifacts
runs/                     # durable references to external executions
output/                   # findings, benchmark candidates, method proposals
```

The layout is a useful default, not a required reasoning protocol. The agent
may create its own files, tools, and helper topology. It should write important
claims with evidence references and surface focused questions when stakeholder
judgment is needed.

The first default skill should teach the agent to:

- query before scanning all history;
- open raw evidence when summaries are insufficient;
- distinguish system, harness, task, environment, and verifier failures;
- use Herdr for helpers and long-running processes;
- use Harbor for executable checks;
- propose experience only when it may change later work;
- state claim scope and unresolved uncertainty.

## 8. Delivery sequence

### A. Expose the minimum researcher core

Status: first skeleton implemented. The Go CLI can capture arbitrary files and
directories, query and materialize textual context, and propose or revise
evidence-backed experience. An initial eval-researcher skill defines the agent
method without imposing a fixed workflow.

Keep these as capabilities called by a normal agent. Do not turn the CLI into a
universal investigation state machine or wrap every Harbor and Herdr action.

### B. Run one complete investigation in Herdr

Stand up the primary researcher in a persistent workspace. Give it a high-level
intent and selected source access, then let it query context and directly use
Git, native sessions, Harbor, and helper agents. Record its hypotheses, tool
choices, human interventions, cited evidence, finding, and retained experience.

Compare the path with the same agent given a direct prompt, native history, and
Harbor at a matched budget. This first vertical, not backend completeness, is
the next implementation milestone.

### C. Harden the boundaries the vertical exercises

Fix execution-time task identity, incomplete-job ingest, multi-trial
representation, artifact portability, and generated task files as the real
investigation reaches them. Add behavioral tests for each observed failure.
Spike only the Herdr capabilities required by the vertical and preserve native
session identity and capture completeness.

### D. Improve retrieval from observed misses

Import selected evidence from the local study and a small set of native agent
sessions without normalizing their trajectories or committing private raw
material. Replace scan-based search with a rebuildable full-text index when
corpus size, latency, or missed evidence demonstrates the need. Add semantic
search only after lexical and metadata retrieval have a measured limitation.

### E. Test retained-experience transfer

Run a second investigation from new source material. Compare the system with
and without selected prior experience, or compare two frozen context or repair
methods. Record whether the experience improved quality, time, cost, or expert
correction.

## 9. Verification

The vertical is useful only if it checks behavior beyond happy-path storage:

- relevant evidence can be found without opening every source session;
- raw artifacts remain available and match their recorded digests;
- partial and failed runs are queryable;
- run, task, model, harness, environment, and exposure identities are exact;
- a reviewer can reproduce a finding from cited artifacts;
- rejection does not create a release and later approval remains possible;
- stateful gains survive new source material and a matched baseline;
- the system reduces user or expert work enough to justify its machinery.

## 10. Deferred

- a universal trace or memory schema;
- mandatory launch through `vri`;
- automatic ingestion of every user session;
- a fixed multi-agent workflow;
- a vector database before retrieval evidence requires one;
- a Web UI before terminal review becomes limiting;
- a provider registry for hypothetical backends;
- training or serving infrastructure;
- organization-wide autonomous release authority;
- online weight learning or open-ended RSI.

## References

- [Herdr agent automation](https://herdr.dev/docs/agent-automation/)
- [Herdr socket API](https://herdr.dev/docs/socket-api/)
- [Herdr session state and restore](https://herdr.dev/docs/session-state/)
- [Herdr multi-machine operation](https://herdr.dev/docs/connecting-machines/)
- [Harbor](https://github.com/harbor-framework/harbor)
- [System design](DESIGN.md)
- [Eval researcher proposal](EVAL_RSI_PROPOSAL.md)
