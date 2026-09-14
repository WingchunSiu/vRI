# vRI

**A runtime for an eval researcher that learns from experience.**

Evaluation work is more than writing tasks or collecting scores. Someone has
to decide what matters, turn real cases into fair and reproducible tests,
inspect failures, repair tasks and verifiers, and keep the result useful as
models and environments change. Today that work is slow, expert-heavy, and
poorly accumulated across projects.

vRI is building an **Eval Researcher System**: an agent, its harness, tools,
skills, context, memory, environment, and any helper agents it invokes. Its
first workload is benchmark construction and maintenance for individuals and
organizations.

```text
evaluation intent + real work + existing evals
  -> investigate what is worth measuring
  -> build, run, inspect, and repair evaluations
  -> release a benchmark with evidence and limitations
  -> learn from later failures, decisions, and user feedback
  -> improve how future evaluation work is done
```

## System boundary

- **Herdr** can host persistent agents and processes, coordinate helpers, and
  expose live state across local and remote machines.
- **Harbor** can package and execute evaluation tasks in sandboxes and return
  verifier results and trajectories.
- **vRI** makes source material, runs, artifacts, findings, and prior
  experience searchable and reusable. It also records versions, exposure,
  evidence, decisions, and method changes needed to test continual learning.

These are capability boundaries, not a fixed workflow. The eval researcher
chooses what to inspect, which agents or tools to invoke, whether to build or
repair a task, and what is worth retaining as experience.

## Design principles

- **Preserve raw evidence; structure only useful boundaries.** Native session
  logs, trajectories, terminal output, diffs, and Harbor artifacts may remain
  in their original formats. vRI adds enough identity, provenance, scope, and
  version information to find and cite them.
- **Experience is a revisable interpretation.** A lesson or playbook must link
  back to its evidence and may later be supported, narrowed, contradicted, or
  retired. Saving every transcript as permanent memory is not learning.
- **Context is computed for the current task.** Agents query relevant evidence
  and experience instead of rereading every prior session. They may still open
  the full source when needed.
- **Scores require explanation.** A pass or failure may come from model
  capability, harness behavior, environment failure, verifier error, or reward
  hacking. Important conclusions retain the artifacts needed for attribution.
- **Evaluation and acceptance are separate.** The system records what an
  improver saw, freezes candidate methods before protected checks, and keeps
  evidence distinct from human authorization and release.
- **Use strong existing systems.** vRI should improve real agent and evaluation
  workflows, not replace agent interfaces, Harbor, sandboxes, or model APIs.

The near-term product is a useful eval researcher. The continual-learning
claim begins when accumulated experience improves its performance on later
evaluation work. The recursive-self-improvement claim begins only when a
validated change to the researcher system makes it better at producing future
validated changes.

## Status

vRI is pre-stable. A local prototype study produced one real Harbor-backed
evaluation task and four runs, but its raw sessions and machine-specific
artifacts are intentionally not part of this public repository. The Go code is
a storage and query walking skeleton; it does not yet implement the Eval
Researcher System or the Herdr integration.

- [System design](docs/DESIGN.md)
- [Eval researcher proposal](docs/EVAL_RSI_PROPOSAL.md)
- [Initial technical plan](docs/TECHNICAL_PLAN.md)
- [Roadmap](ROADMAP.md)
- [Current v0 storage contract](docs/CORE_DESIGN.md)
- [RSI research notes](docs/research-notes/2026-09-11-dwarkesh-rsi-roundtable.md)

## License

Apache-2.0.
