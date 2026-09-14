# Proposal: An Eval Researcher That Learns From Experience

> Status: research and product proposal. The first local study is ready for
> human review, but it is not yet an approved benchmark and does not demonstrate
> continual learning. Its raw evidence remains local. [DESIGN.md](DESIGN.md)
> defines the system boundaries;
> [TECHNICAL_PLAN.md](TECHNICAL_PLAN.md) defines the next implementation slice.

## 1. The problem

A benchmark is a measurement instrument. Its value depends on more than the
number of tasks or the reliability of a runner. Someone must decide:

- which real capabilities and failures are worth measuring;
- whether a task represents the intended work;
- whether its instruction, environment, rubric, and verifier agree;
- whether a pass or failure came from the system under test or from the eval;
- whether a new task adds information to the existing portfolio;
- when a saturated, contaminated, or drifting benchmark should change.

Those decisions require evidence and judgment. Today they are spread across
evaluation engineers, model researchers, product managers, domain experts,
task authors, and users. The workflow is slow, inconsistent, and difficult to
reproduce. Internal benchmarks are not exempt: a domain expert may provide a
valuable but malformed task, a PM may know what users need but not how to build
a fair verifier, and an eval engineer may build a reliable environment for the
wrong measurement target.

Agentic evaluation makes the problem more visible. The same zero can mean a
real capability failure, a missing dependency, a harness timeout, an overly
strict hidden test, an ambiguous instruction, or a valid answer that the
reference did not anticipate. A pass can be a correct solution, weak test
coverage, leaked state, or a reward hack. Aggregate pass/fail scores cannot
resolve these cases; trajectories, artifacts, environment events, and verifier
behavior are part of the evidence.

Frameworks such as Harbor improve packaging, isolation, execution, and result
collection. That is necessary, but it does not answer what matters, whether a
task is semantically valid, or how a portfolio should evolve. A Harbor-shaped
task can still measure the wrong thing perfectly.

## 2. Product thesis

Build a high-quality **eval researcher system**: an execution agent together
with its runtime, tools, skills, context, memory, environments, and helper
agents. Given an evaluation intent, access to relevant experience, and a
budget, it should investigate what is worth measuring and produce defensible
evaluation outcomes.

Possible outcomes include:

- a new or repaired task;
- a personal, team, or company benchmark;
- an explanation of a surprising score;
- a comparison between models or harnesses;
- a recommendation to quarantine or retire a task;
- evidence that the requested claim cannot yet be measured reliably.

The product is not merely a benchmark maintainer. Benchmark creation and
maintenance are major workloads, especially for companies, but the stable
center is the agent that can do evaluation research well. New tools and product
surfaces should be judged by whether they make that agent more capable,
reliable, or able to learn from later evidence.

## 3. Why this is timely

### 3.1 Benchmark selection is a form of taste

Every benchmark encodes choices about domain, task distribution, realism,
constraints, aggregation, and weights. Public indexes make these choices
visible; internal suites make the same choices less visibly. Different teams
reasonably care about speech, computer use, coding, scientific work, or their
own application workflows.

Benchmark-directed development is not inherently a problem. The failure mode
is overfitting decisions to a small, correlated, stale, or invalid set of
measurements. More high-quality and meaningfully different evaluations would
make model capability less dependent on a few headline scores.

The scarce capability is therefore not bulk task generation. It is choosing
promising measurement directions and gathering enough evidence to distinguish
a useful instrument from a plausible-looking one.

### 3.2 Benchmarks will need continuous renewal

Frontier benchmarks saturate, leak, drift, and become targets of optimization.
As systems improve faster, the useful lifetime of a static public benchmark is
likely to shrink. Teams will need harder tasks, more realistic workflows,
stronger reliability checks, and new capability surfaces.

AI already accelerates task proposals, environment construction, verifier
writing, trace inspection, and quality review. Letting an agent coordinate more
of that work is a natural next step. The goal is not to maximize benchmark
churn; it is to reduce the cost of finding and validating the next useful
measurement.

### 3.3 Users need evaluations drawn from their own distribution

Official scores reflect the use cases selected by benchmark authors and model
providers. An individual or company often cares about a shifted distribution.
Today they compensate with informal vibe tests: try a few familiar tasks on a
new model and inspect the result.

An eval researcher can make this behavior systematic without pretending that
a ten-task personal suite is a general leaderboard. It can turn selected local
episodes, corrections, incidents, and workflows into a private regression and
comparison set; preserve uncertainty; and rerun it locally or through an eval
backend. For organizations, the same core agent can turn distributed internal
knowledge into larger governed benchmarks.

### 3.4 Synthetic generation needs an external correction signal

Synthetic tasks are useful for controlled variants and coverage expansion, but
they do not establish real-world relevance on their own. A generator, judge,
and target system drawn from similar models can create a closed loop with
shared blind spots.

The system should treat synthetic data as an experimental instrument. User
workflows, production failures, expert cases, independent sources, and later
outcomes supply the external experience that can correct it.

## 4. What the eval researcher does

The researcher receives goals rather than a mandatory workflow. Depending on
the case, it may:

1. inspect prior episodes, artifacts, outcomes, and existing evaluations;
2. propose competing hypotheses about what should be measured;
3. select high-information cases under a budget;
4. abstract a real episode into a reusable task without erasing its provenance;
5. build or repair the environment, instruction, rubric, and verifier;
6. run oracle, no-op, target, and adversarial attempts through Harbor or another
   backend;
7. inspect traces and artifacts to attribute outcomes;
8. revise, split, quarantine, or reject tasks;
9. compare the candidate with the current benchmark portfolio;
10. preserve conclusions, counterexamples, and reusable methods for later work.

These are capabilities, not a fixed pipeline. A broken dependency may require
several repair-and-rerun iterations. An open-ended task may need paired human
judgment rather than hidden tests. A clear production regression may need no
broad capability taxonomy at all. The agent should choose the investigation;
the system should preserve what makes that choice inspectable and reusable.

## 5. Evidence without a rigid ontology

The system needs more than a long-lived chat session, but it should not force
every useful observation into one elaborate schema.

Four distinctions are enough to guide the design:

- **raw evidence**: trajectories, logs, files, screenshots, verifier output,
  source episodes, and user corrections;
- **experience**: a revisable interpretation such as “this failure came from a
  lost virtual-environment path” or “this verifier is unsafe under dependency
  failure”;
- **method**: a reusable procedure, heuristic, skill, or tool change;
- **decision**: selection, release, quarantine, rejection, or escalation.

Raw evidence can remain in its native format. Minimal envelopes provide stable
identity, provenance, timestamps, scope, and links. More structure is added
only where execution, search, comparison, access control, or release requires
it. Experience may be prose, a playbook entry, a code change, or a compact
record. Context is assembled for the current investigation rather than copied
wholesale from all prior sessions.

This is materially different from asking an agent to reread a directory of old
trajectories each time. vRI should provide durable references, provenance,
content identity, retrieval, relations between evidence and conclusions,
versioned decisions, and reproducible handoff. Herdr can keep agents and helper
processes alive across local or remote hosts and coordinate their work. Harbor
can execute evaluations. vRI supplies the shared evidence and experience
substrate across those sessions and systems.

The detailed boundary is in [DESIGN.md](DESIGN.md). The current three-table
store is an implementation starting point, not the product ontology.

## 6. What “valid” means

No single run proves that a benchmark is valid. Different claims require
different evidence:

| Claim | Useful evidence |
| --- | --- |
| The task can execute | clean replay, dependency and service checks |
| The verifier rewards intended success | oracle or known-good behavior |
| The verifier rejects intended failure | no-op and known-bad behavior |
| The task resists shortcuts | adversarial attempts and state inspection |
| A failure belongs to the tested system | trace-level attribution and reruns |
| The task represents important work | source episodes, expert judgment, or downstream outcomes |
| The task adds portfolio value | discrimination, failure overlap, and decision impact |
| A benchmark predicts real behavior | temporally later or otherwise held-out outcomes |

Harbor runs can establish several technical claims: an environment launches, a
verifier behaves under probes, a target receives a fair attempt, and failures
can be reproduced. They cannot alone establish importance, coverage, or
predictive validity. Conversely, a realistic task proposal without a clean run
is not yet a usable eval. The researcher must connect both forms of evidence.

## 7. Continual learning and RSI

The initial useful product does not need to prove recursive self-improvement.
The claims should be earned in order:

1. **Useful execution:** the researcher completes one real evaluation job.
2. **Persistent assistance:** prior evidence reduces repeated work or improves
   explanations on later jobs.
3. **Continual learning:** accumulated experience improves performance on new
   evaluation work at comparable cost.
4. **Recursive improvement:** the system changes its own evaluation-research
   method, and that change improves later improvement work on held-out cases.

The analogy to Continual Learning Bench is useful: compare a stateful system
with a stateless version under the same task sequence, base model, tools, and
budget. The difference must appear on later, unexposed evaluation work—not only
as better recall of an earlier answer.

For vRI, the learning target is the eval researcher itself. It might learn that
a class of verifier traps must fail closed, that a particular artifact is
needed to attribute browser failures, or that one task family adds little
information. A durable note is not yet learning. The experience counts only if
retrieving or operationalizing it improves a later outcome.

RSI is a stronger claim. It requires a method change—such as a revised skill,
investigation policy, context strategy, helper-agent topology, or evidence
rule—and a comparison showing that the new method produces better later
evaluation work. Merely accumulating more context, generating more tasks, or
raising scores on the benchmark used to make the change is insufficient.

## 8. First research program

### 8.1 Question

Can an eval researcher using indexed prior evidence and a persistent runtime
produce a more trustworthy evaluation outcome than a capable coding agent
given raw session files and Harbor directly?

### 8.2 First vertical

Use one bounded coding-agent evaluation problem with real provenance. The
researcher should:

- recover the relevant source episode and intent;
- build or repair a Harbor task;
- run technical validation and at least two representative systems;
- explain important pass/fail outcomes from trajectories and artifacts;
- preserve a rerunnable candidate and its unresolved limitations;
- use the resulting experience on a second, related but unseen case.

The current local study covers much of the first case. Its human review remains
pending, and it does not cover retained-experience transfer.

### 8.3 Baselines

- **Direct agent:** the same base agent receives raw session paths, repository
  access, and Harbor, but no vRI retrieval or retained experience.
- **Stateless researcher:** the same harness and tools, with evaluation history
  removed before the second case.
- **Generation-only:** an agent proposes tasks from the high-level intent
  without real source episodes.
- **Human workflow:** where feasible, record the expert time and defects from
  the current process rather than treating it as an undefined gold standard.

All comparisons should match source access, model, tool permissions, compute,
and human-review budget as closely as the question requires.

### 8.4 Outcomes

Measure outcomes that can change a real decision:

- serious task, environment, and verifier defects found before release;
- correctness and usefulness of trace-level attribution;
- rerun success from preserved artifacts and versions;
- expert-rated importance and representativeness;
- incremental information relative to the existing portfolio;
- expert interventions, elapsed time, compute, and model cost;
- performance on a later case with and without retained experience;
- downstream agreement with later real-world behavior, when available.

Do not collapse these into a single “benchmark quality” score before the
individual measurements are understood.

## 9. Autonomy and governance

The eval researcher needs autonomy over reversible investigation: what to
inspect, which helper to invoke, which probe to run, and how to revise a
candidate. Human control remains appropriate for access to private data,
material external cost, publication, release gates, destructive operations,
and unresolved semantic judgments with real consequences.

Human review should be evidence-directed. The goal is not approval ceremony;
it is to surface the few questions for which human taste, domain knowledge, or
authority changes the result. Run authorization is not benchmark selection,
and technical validation is not release approval.

## 10. Risks and falsifiers

The proposal should be narrowed or rejected if:

- the researcher prefers measurable but unimportant questions;
- polished reports do not reduce serious benchmark defects;
- stored evidence adds ceremony but does not change later actions or decisions;
- users must restate nearly the entire task and capability map each time;
- stateful gains disappear under matched context, cost, and source access;
- learned methods fail to transfer to new cases;
- private episodes cannot be used with acceptable retention and redaction;
- experts must redo almost all of the work to trust the output;
- Harbor plus an ordinary prompted agent achieves the same result with similar
  effort, reliability, and continuity.

Important open questions remain: how to measure taste, how much personal data
is enough, which experiences should become reusable methods, how to avoid
shared model blind spots, and when task-family diversity represents real
coverage rather than cosmetic variety.

## 11. Current status

One local study derived from real work includes source provenance, candidate
selection notes, a Harbor task, four runs, an evidence packet, and a pending
review packet. It demonstrates that a concrete evidence-to-rerunnable-task path
can be assembled. The raw study is intentionally untracked because it contains
machine-specific and potentially private execution evidence. It does not
establish that the task should enter a benchmark, that the current storage
design is sufficient, or that the researcher learns.

The Go implementation is a local storage and query skeleton extracted from
that study. The next step is not more schema. It is one Herdr-managed eval
researcher completing a second realistic investigation while using the stored
evidence and experience substrate.

## 12. Related work

- [Self-Improving Agents: Learning How to Work](https://furong-huang.com/blog/self-improving-agents-learning-how-to-work/)
  motivates improving an agent's way of working rather than only its weights.
- [Recursive Self-Improvement in AI](https://arxiv.org/abs/2607.07663)
  surveys systems that improve components of their own improvement process.
- [Rethinking the Evaluation of Harness Evolution for Agents](https://arxiv.org/abs/2607.12227)
  motivates held-out tasks and budget-matched baselines for harness changes.
- [TASTE: Can AI Models Judge AI Safety Research Proposals?](https://alignment.anthropic.com/2026/taste/)
  studies model prediction of expert research judgment and disagreement.
- [Artificial Analysis benchmarking methodology](https://artificialanalysis.ai/methodology/intelligence-benchmarking)
  illustrates that capability selection and weighting form a measurement contract.
- [Separating signal from noise in coding evaluations](https://openai.com/index/separating-signal-from-noise-coding-evaluations/)
  documents task, environment, and verifier defects in agentic evaluations.
- [Harbor task format](https://github.com/harbor-framework/harbor/blob/main/docs/content/docs/tasks/index.mdx)
  provides an execution substrate rather than a theory of what should be measured.
- [Harbor Adapters and Harbor-Index](https://arxiv.org/abs/2609.04298)
  is an adjacent baseline for standardization and trajectory-grounded repair.
- [Oak Lab mission](https://oaklab.ai/mission) and the
  [Sutton–Javed interview](https://sequoiacap.com/podcast/rich-sutton-and-khurram-javed-why-ai-models-stop-learning-and-how-to-start-it-again)
  motivate learning continuously from external experience.
