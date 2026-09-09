# vRI Initial Technical Plan

> Status: initial planning document. It records current hypotheses and should be
> revised after the first real integration and workflow tests.

## 1. Design Stance

vRI should provide **flexible agents and opinionated backend invariants**.

Agents must have room to choose their own research process. vRI should not force
every run through a fixed sequence of planning, hypothesis generation,
experimentation, and evaluation. It should make these activities easy to record
without requiring them to occur in one prescribed order.

The backend is opinionated about the properties required for durable research:

- stable identity;
- causality and lineage;
- atomic state changes and event recording;
- explicit actors and human interventions;
- permissions and budgets;
- reproducible references to context, artifacts, and evaluations.

The rule is **hard invariants, soft workflow**.

## 2. Local-First Architecture

```text
                       Human / agents
                             |
                   CLI / API / shared files
                             |
                        vRI process
              +--------------+--------------+
              |                             |
          vRI kernel                 workflow modules
      identity / events /         context / artifact /
      transactions / policy       eval / findings /
      service registry            decisions / runs
              |                             |
              +-----------+-----------------+
                          |
                capability providers
           Herdr / host / Git / filesystem /
                    command evaluator
```

The first version is one local process and one local state directory. Services
are modules behind explicit interfaces, not separate microservices.

## 3. Service Model

A service is a logical capability contract. A provider implements that
contract. A plugin packages a provider, policy, or event hook.

These concepts do not imply a process boundary:

- initial providers may be compiled into the `vri` binary;
- an adapter may call an existing executable such as Herdr or Git;
- a later provider may live behind a local socket or remote API;
- moving a provider out of process must not change campaign semantics.

Initial service boundaries:

| Service | Responsibility | Initial provider |
| --- | --- | --- |
| `CampaignService` | Durable campaign lifecycle | built in |
| `AgentService` | Translate work into an agent interface | arbitrary CLI |
| `SessionService` | Start, send, wait, read, attach, resume, stop | Herdr; local fallback |
| `EnvironmentService` | Compute, dependencies, network, isolation | host |
| `WorkspaceService` | Working directory, branch, diff, checkpoint | Git worktree |
| `ContextService` | Shared room, inbox, context snapshots | local filesystem |
| `ArtifactService` | Artifact identity, blobs, and lineage | Git plus local blobs |
| `EvalService` | Invoke and record user-owned verification | command |
| `PolicyService` | Budgets, approval, promotion, stopping | built in, minimal |

The service definition owns request and result types. Consumers depend on the
definition, not on a specific provider.

## 4. Execution Is a Composition

Herdr, a local host, and Docker are not competing runtime modes. They occupy
different axes:

```text
Agent:        Codex / Claude Code / custom
Session:      Herdr / native supervisor / other
Environment:  host / Docker / microVM / remote machine
Workspace:    directory / Git worktree / mounted volume / snapshot
Context:      shared filesystem / room history / generated bundle
```

A local run may be:

```yaml
agent: codex
session: herdr
environment: host
workspace: git-worktree
context: shared-filesystem
```

A later isolated run may keep Herdr while changing only the environment:

```yaml
agent: codex
session: herdr
environment: docker
workspace: mounted-worktree
context: mounted-shared-filesystem
```

Capabilities such as snapshot and fork are reported by the selected provider;
they are not assumed to exist everywhere.

## 5. Herdr Integration

Herdr already provides a mature local terminal and agent-session substrate:

- persistent terminal panes and human attach;
- workspace and Git worktree operations;
- agent start, prompt, read, wait, and status;
- lifecycle detection for common CLI agents;
- native agent session references and resume;
- a JSON CLI, local socket API, and event subscriptions.

Herdr does not own vRI's durable research context, findings, experiment history,
artifact provenance, or campaign decisions. Its agent-to-agent communication is
terminal/session coordination rather than a durable research room.

Therefore Herdr should implement `SessionService` and optionally parts of
`WorkspaceService`. It must not become the vRI domain model or the only possible
session provider.

The first integration should use Herdr's JSON CLI for simple operations. A raw
socket subscriber should be added only when continuous lifecycle events are
needed. vRI should not vendor or fork Herdr unless later evidence shows the
public interfaces are insufficient.

The decisive spike is:

```text
create workspace
  -> start agent
  -> send work
  -> wait and read
  -> attach as a human
  -> recover or resume
```

If this composition remains natural, Herdr can be the preferred local session
provider. If vRI repeatedly works around Herdr's abstractions, it remains an
optional adapter.

## 6. Durable Shared Context

The local context model is a shared research room materialized into a filesystem
volume. It borrows the useful property of a group chat: humans and agents share
one continuing body of context and can resume where another participant left
off.

The filesystem makes context legible to arbitrary CLI agents without a custom
SDK:

```text
.vri/
├── vri.db
├── shared/
│   ├── brief.md
│   ├── findings/
│   ├── decisions/
│   ├── evals/
│   ├── inbox/
│   │   └── <agent-id>/
│   └── artifacts/
├── contexts/
│   └── <run-id>.md
└── blobs/
```

Responsibilities:

- `vri.db` stores canonical coordination state and the event history;
- `shared/` contains human-readable shared material;
- `contexts/<run-id>.md` is the immutable context snapshot given to one run;
- `blobs/` contains content-addressed logs and outputs;
- each agent has an inbox and a cursor into the shared event stream.

The full history remains available for pull-based search. Context providers may
push a relevant slice to an agent, but no summarizer is trusted to be the only
route to older information.

Agents may edit ordinary workspace files directly. Mutations that should enter
the durable research graph are registered through vRI commands so actor,
causality, and provenance are preserved. Generated filesystem views should not
be treated as an independent source of truth.

Start with paths, explicit references, tags, SQLite full-text search, and normal
filesystem search. Do not add a vector database before real context-retrieval
failures show that it is necessary.

## 7. State and Storage

Use SQLite for both normalized current-state projections and an append-only
event history. A command updates its projection rows and appends its audit event
in the same transaction.

This is not a commitment to pure event sourcing. The event history exists for
inspection, causality, recovery, and future projections; normal relational
queries serve the current state.

Use:

- Git and worktrees for source artifacts;
- content-addressed local files for logs, context snapshots, and small outputs;
- URI plus digest metadata for large datasets, model weights, or remote objects.

The local database and object files can later map to PostgreSQL and object
storage without changing object identity or service contracts.

## 8. Agent-Facing Interface

The CLI is the first user and agent interface. The tentative surface is:

```text
vri init
vri run
vri send
vri publish
vri artifact
vri eval
vri context
vri status
vri attach
vri stop
```

The CLI talks to the local vRI process. Agents can use it from a shell without
special integration. An MCP surface may expose the same capabilities later, but
MCP is an adapter rather than the internal domain protocol.

There is no graphical interface in the initial version. Human interaction uses
the CLI and Herdr's existing attach experience.

## 9. Plugin Shape

DeepSeek Harness demonstrates the value of service definitions, replaceable
providers, typed events, and append-only run history. vRI should borrow those
principles without adopting a large general-purpose harness as its core.

Initial providers should be built in or implemented as thin adapters. When an
external plugin protocol becomes necessary, begin with a small manifest and
JSON request/result contract, for example:

```toml
id = "session.herdr"
provides = ["session"]
command = ["vri-provider-herdr"]
```

Do not begin with an in-process binary ABI, hot reloading, a package registry,
or a complex plugin dependency graph.

The kernel itself is not replaceable. It owns global identifiers, event
ordering, transactions, causal references, artifact identity, permissions, and
provenance. Providers extend capabilities within those invariants.

## 10. Initial Stack

- Go for the core process and CLI;
- SQLite for state and events;
- the local filesystem for context and blobs;
- Git CLI and worktrees for code artifacts;
- TOML for configuration and provider manifests;
- local socket plus JSON for local control;
- Herdr as the first session provider;
- arbitrary commands as the first agent and evaluator adapters.

This stack optimizes for a small installation, inspectable state, and the
ability to replace providers later.

## 11. Local and Managed Evolution

The local and managed versions should share the same objects and service
semantics:

| Capability | Local | Managed |
| --- | --- | --- |
| State | SQLite | PostgreSQL |
| Blobs | local filesystem | object storage |
| Session | Herdr or native | managed agent session |
| Environment | host or Docker | sandbox or remote compute |
| Context | local shared volume | mounted or synchronized volume |
| Interface | local CLI/socket | CLI/SDK over authenticated API |

Managed infrastructure should be introduced only after the local workflow is
useful. Distribution is a provider and deployment concern, not a reason to
change the research model.

## 12. Immediate Validation

The first implementation work should answer three questions:

1. Does Herdr remove most session and terminal lifecycle work without leaking
   terminal-specific details into the vRI domain model?
2. Can a second agent resume useful work from the shared volume and a bounded
   context snapshot without receiving the full transcript?
3. Can a human reconstruct why an artifact changed from the event, finding,
   evaluation, and decision records?

These checks are more important than implementing a broad plugin system or
additional infrastructure providers.

## References

- [Herdr](https://github.com/herdrdev/herdr)
- [Herdr agent automation](https://herdr.dev/docs/agent-automation/)
- [Herdr socket API](https://herdr.dev/docs/socket-api/)
- [DeepSeek Harness](https://github.com/deepseek-ai/deepseek-harness)
- [DeepSeek Harness service design](https://deepseek-harness.github.io/deepseek-harness/en/develop/practice/)
