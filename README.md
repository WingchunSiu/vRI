# vRI

**A human-supervised research runtime for agents.**

vRI helps people and AI agents run long-lived research and artifact-improvement
work without losing the reasoning, experiments, and state that produced each
version.

Today, capable agents can already write code, change prompts, analyze data, and
run experiments. The harder problem is coordinating those agents over time:

- context disappears when an agent session ends;
- parallel agents duplicate work or overwrite one another;
- hypotheses, experiments, and decisions are scattered across chats and shells;
- humans cannot easily inspect, interrupt, reproduce, or resume the process;
- apparent improvements are difficult to trace back to evidence.

vRI treats agents as replaceable workers and **research state as the durable
system of record**.

## What vRI aims to provide

- Bring-your-own agent, model, evaluator, and compute.
- Flexible research behavior without a prescribed agent loop.
- Persistent, inspectable agent sessions with human intervention.
- Structured context and handoffs between agents.
- Versioned artifacts, experiments, evidence, and decisions.
- Branching, comparison, replay, and rollback of research work.
- Pluggable verification rather than a built-in claim about what is "better."

vRI is not intended to be another foundation model, coding agent, training
framework, or sandbox implementation. It is the control plane connecting those
systems into a coherent research workflow.

## Status

vRI is at the design stage. The current documents describe a working thesis,
not a frozen specification:

- [System design](docs/DESIGN.md)
- [Initial technical plan](docs/TECHNICAL_PLAN.md)
- [Roadmap](ROADMAP.md)

## License

Apache-2.0.
