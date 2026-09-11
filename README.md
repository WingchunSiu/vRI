# vRI

**A bring-your-own-agent runtime for improving how agent systems work.**

Agents already complete useful tasks, but ordinary use rarely compounds into a
better system. Saving a trajectory or writing a lesson preserves experience; it
does not show that the agent has learned a better method or that the method
helps on tasks it has not seen before.

vRI explores the missing loop:

```text
ordinary agent use
  -> experience and observed weakness
  -> candidate method change
  -> independent evaluation
  -> scoped release and routing
  -> future agent use
```

The agent may be Codex, Claude Code, Prime Agent, a custom CLI, or another
runtime. vRI does not require access to its internal reasoning or harness. It
can improve the larger system around an opaque agent: skills, context
strategies, tools, services, environments, routing, workflows, and eventually
the mechanism that proposes improvements.

## Working thesis

- **Experience is input, not learning.** A trace matters only if it can produce
  a method change whose benefit survives independent, future evaluation.
- **The unit of improvement is the agent system.** A system version composes an
  agent with a versioned improvement envelope rather than treating the model as
  the only mutable component.
- **Agents choose the workflow.** vRI exposes reliable, programmable
  capabilities through CLI, skills, and service adapters instead of imposing a
  universal improvement pipeline.
- **Verification governs accumulation.** Evidence, judgment, release, routing,
  and rollback are distinct. A candidate does not become the new global
  default merely because it won one comparison.
- **Everything may be referenced as a service; vRI does not implement
  everything.** Users bring agents, models, training, serving, evaluation, and
  compute. vRI records their exact versions and controls which surfaces are
  fixed, mutable, or protected in an improvement run.

A successful promotion may add one skill, change routing for one task family,
enable a workflow only within one scope, or retain an older method as a cheap
path or fallback. Improvement is therefore compositional and contextual, not a
single sequence of replacements.

## Initial direction

The first vertical keeps a bring-your-own agent opaque and allows an improver
to change its external skills, context strategy, and environment. The system
must then evaluate the candidate on unseen future tasks, release it to a
defined scope, resolve the correct system version for a later episode, and
retain enough evidence to explain or reverse the decision.

Eval-RSI is a related research branch concerned with discovering weaknesses and
improving the mechanisms that judge improvement. It may later provide versioned
evaluation services to the main loop, but it is not the initial product scope.

## Status

vRI is an ongoing, pre-stable design project. The current documents are working
hypotheses to be tested against real use, not compatibility commitments.

- [System design](docs/DESIGN.md)
- [Initial technical plan](docs/TECHNICAL_PLAN.md)
- [Roadmap](ROADMAP.md)
- [Experience-grounded eval evolution proposal](docs/EVAL_RSI_PROPOSAL.md)

## License

Apache-2.0.
