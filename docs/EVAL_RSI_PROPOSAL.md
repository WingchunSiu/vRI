# Proposal: Experience-Grounded Eval Evolution (Eval-RSI)

> Status: early working proposal. This is an independently testable research
> branch, not the first vRI product vertical or part of the core specification.
> If its hypotheses survive testing, its evaluators and portfolios can become
> protected, versioned services in the broader vRI improvement loop.

## 1. Summary

Evaluation is the control signal for model and agent improvement. A benchmark
does more than measure a system: its tasks, environments, verifiers, and
aggregation rules encode what its authors believe is important and how they
want the system to behave.

Today, building a trustworthy agentic benchmark remains expensive and fragile.
The instruction may not match the verifier. The environment may be broken,
mutable, or difficult to reproduce. A weak verifier may reject valid behavior
or admit shortcuts. Aggregate scores can hide these failures, so a score is
often uninterpretable without inspecting the task and the full agent
trajectory. These problems occur in public benchmarks and in internal
evaluation suites at large organizations.

This proposal asks whether agents can help scale the construction, audit,
repair, selection, and continued evolution of evaluations while preserving
external grounding and independent human control.

The proposed system turns evaluation into a governed lifecycle:

```text
real experience and failures
  -> candidate task and verifier
  -> reproducibility and trace audit
  -> independent review and acceptance
  -> versioned evaluation portfolio
  -> model or harness improvement
  -> held-out and future-experience validation
  -> new failures and evaluation revisions
```

The long-term direction is evaluator-mediated recursive improvement: the
system improves not only a model or agent harness, but also the measurement
system that discovers weaknesses and directs subsequent improvement. The
evaluator may evolve, but the system being optimized must not be able to approve
its own evaluator or its own improvement claim.

## 2. Motivation

### 2.1 Evaluation determines the direction of improvement

Benchmark-driven development is not inherently a mistake. How a benchmark is
built partly specifies how its authors want a system to work. The problem is
overfitting to a small, narrow, or invalid set of measurements.

Benchmark selection is also an expression of research and product taste. Some
teams prioritize computer use, others speech, scientific reasoning, coding, or
the belief that a coding agent can mediate most digital work. An index turns
these judgments into a capability taxonomy and a set of weights. For example,
the Artificial Analysis Intelligence Index explicitly assigns category and
per-evaluation weights and evaluates speech, image, and multilingual capability
separately from its primarily English, text-based index. Its rankings should
therefore be read as conditional on that measurement contract, not as a
context-free ordering of intelligence.

If many high-quality and meaningfully independent benchmarks existed, teams
could observe a more complete capability surface and would be less dependent on
one headline score. In practice, frontier evaluations remain sparse, costly,
correlated, and uneven in quality. More benchmark names do not by themselves
solve this problem: several suites may encode the same capability ontology,
share task-generation methods, use the same judge, or reproduce the same
blind spots.

The relevant quantity is therefore effective coverage, not benchmark count.

### 2.2 A score is not trustworthy without task- and trace-level validity

An agentic evaluation combines at least five moving parts:

1. an intended capability or real-world behavior;
2. a task distribution and concrete instructions;
3. an executable environment and resource policy;
4. an agent harness and interaction protocol;
5. a verifier, rubric, and aggregation policy.

A failure in any one of them can be mistaken for a model capability result.
Common failure modes include:

- tests requiring behavior that the instruction did not specify;
- important requirements that the verifier does not check;
- an environment that fails before the agent can attempt the task;
- mutable dependencies, credentials, network services, or time-sensitive data;
- harness budgets or timeouts that reject slow but valid solutions;
- leaked artifacts or shortcuts that permit reward hacking;
- a rubric that rewards presentation while missing functional failure;
- an apparently correct final state produced by invalid or unsafe behavior;
- a valid solution rejected because the verifier assumes one implementation;
- aggregate scores that conceal which of these mechanisms caused the result.

These are not hypothetical edge cases. A 2026 OpenAI audit estimated that
roughly 30% of SWE-Bench Pro tasks were broken, with issues including overly
strict tests, underspecified prompts, low test coverage, and misleading
instructions. The audit needed model attempts, task metadata, failure traces,
multiple investigator passes, and experienced human review. The central lesson
is that benchmark quality cannot be inferred reliably from a task package or an
aggregate score alone.

### 2.3 Internal enterprise evaluations have the same problem with weaker process

Large organizations may possess many internal evaluations without possessing a
reliable evaluation process. The workflow is often defined locally by a product
team, research group, vendor, or data-operations team. Quality then depends on
whether a knowledgeable reviewer happens to inspect the right parts of the
task.

The necessary expertise is distributed across roles:

| Role | What they often know | What may be missing |
| --- | --- | --- |
| Product manager | User value, workflows, business priorities | Task isolation, verifier design, statistical validity |
| Domain expert | Realistic cases, important mistakes, professional standards | Executable packaging, precise output contracts, adversarial testing |
| Evaluation engineer | Harnesses, tests, environments, reproducible runs | Whether the task represents an important domain capability |
| Infrastructure engineer | Containers, credentials, networking, resource controls | Task semantics and rubric validity |
| Researcher or model team | Model behavior and likely failure modes | Production frequency and organizational utility |
| Human reviewer | Nuanced judgment and taste | Time to inspect every task, version, and trajectory |

This creates failure at the handoffs. A domain expert can write a valuable task
with ambiguous acceptance criteria or malformed structured output. A PM can
identify an important workflow without knowing how to isolate it into a fair
benchmark. An evaluation engineer can make a task reproducible while preserving
the wrong semantics. A complex environment can rely on unpinned packages,
mutable APIs, secrets, timing, hidden state, or provider-specific behavior, so a
run cannot be reproduced on another machine or several months later.

Internal results are also frequently consumed by people who did not author the
task and do not inspect the trajectories. A table of pass rates may therefore
drive a launch, training, or prioritization decision without showing whether a
failure came from the model, harness, environment, or verifier.

These are working field observations, not yet measured prevalence claims. An
early goal of this project is to document them systematically across real
internal evaluation workflows.

### 2.4 Existing standards solve an important but incomplete layer

The industry is not starting from zero. Harbor, for example, standardizes a
portable agent task around an instruction, task configuration, environment,
tests, and an optional reference solution. It supports isolated environments,
separate verifier sandboxes, resource policies, versioned packages, agents,
trials, jobs, and a common trajectory representation. Its quality checker also
tests instruction-verifier alignment, dependency pinning, structured-data
schemas, anti-cheating measures, and related implementation properties.

These are valuable execution and interchange primitives. They make a task
easier to package, move, run, and inspect. They do not by themselves decide:

- whether the intended capability is worth measuring;
- whether a task distribution reflects production or future use;
- whether a cleanly formatted verifier measures the intended construct;
- whether a task adds new information to an existing portfolio;
- whether a score should trigger a model or product decision;
- how evidence remains comparable after an evaluator changes;
- who is allowed to accept a benchmark proposed by the optimized system;
- how an organization learns from repeated human review decisions.

Recent Harbor-Index work goes further: it adapts many benchmarks to a common
format and constructs an audited subset through difficulty filtering,
trajectory-grounded AI and human review, false-positive/false-negative analysis,
and an audit-and-fix loop. This is a strong adjacent baseline and evidence that
trace-aware benchmark repair is feasible.

Eval-RSI targets the layer above a task format and beyond a one-time curated
index: an organization-specific, continuously operating system for turning new
experience into evaluation evidence, evolving the evaluation portfolio, and
governing the improvement loop that consumes it.

### 2.5 Synthetic data needs external grounding

Synthetic tasks are useful for controlled variation, rare cases, curriculum,
counterfactuals, and closed worlds with reliable rules. They are not an
unbounded source of information about an open world.

Richard Sutton and Khurram Javed argue that any fixed simulator is a severe
approximation of a much larger world and must be corrected continually through
experience. For evaluation, the corresponding risk is an epistemically closed
loop: a model generates tasks from its current assumptions, a related model
judges them, and the target system optimizes against the resulting synthetic
distribution. The loop can create unlimited data while adding little
information not already present in the models.

This proposal therefore treats synthetic generation as an expansion mechanism,
not as the final source of validity:

```text
real trace or expert-observed failure
  -> replayable seed task
  -> synthetic neighboring cases and adversarial variants
  -> independent verifier and human audit
  -> frozen evaluation snapshot
  -> validation against later real experience
```

A benchmark can be viewed as a lossy, frozen, executable compression of
experience. It becomes more trustworthy when its provenance reaches back to a
real need or failure and when its predictions are later checked against new
experience.

## 3. Problem statement

Build a system that helps humans and agents create, audit, repair, select, and
evolve agentic evaluations such that:

- task outcomes correspond to the intended capability claim;
- environments and verifier behavior are reproducible;
- false passes, false failures, and reward hacks are visible at trace level;
- changes to tasks, environments, harnesses, verifiers, or weights are explicit;
- old and new evaluator versions are connected by bridge experiments;
- evaluation portfolios preserve both coverage and meaningful diversity;
- synthetic tasks remain grounded in real experience or independent rules;
- human taste and disagreement are recorded rather than hidden;
- the target system cannot approve its own evaluator or promotion decision;
- improvement claims can be reproduced and audited by another reviewer.

## 4. Thesis

Agents can substantially reduce the mechanical cost of evaluation engineering:
format checking, environment diagnosis, trace triage, verifier analysis,
adversarial probing, task variation, and patch generation.

They are less reliable at deciding which capabilities matter, which failures
deserve scarce evaluation budget, and whether a measurement should change an
important decision. These judgments depend on domain knowledge, organizational
utility, forecasting, and research taste.

The useful near-term architecture is therefore asymmetric:

- agents scale inspection, construction, and candidate generation;
- humans define utility, adjudicate uncertain semantics, and approve protected
  evaluator changes;
- hidden or independent evidence checks whether accepted changes improve real
  predictive validity;
- the system records enough history to learn which human and agent judgments
  were reliable over time.

The long-term hypothesis is that parts of evaluation taste can become learnable
from structured records of proposals, pairwise judgments, disagreement,
confidence, predictions, and later outcomes. Until that is demonstrated,
"taste" remains a human-controlled input rather than an automated scalar reward.

## 5. Research questions and hypotheses

### RQ1: Can agents find real evaluation defects from full trajectories?

**Hypothesis:** A trace-aware investigator with access to the task,
environment, verifier, reference solution, and multiple model attempts will
identify high-severity defects more accurately than a reviewer that sees only
the task package or final score.

### RQ2: Can agent assistance reduce expert audit cost without weakening recall?

**Hypothesis:** Agent triage, evidence collection, and candidate patches will
reduce expert review time while preserving expert recall on false passes, false
failures, environment failures, and exploitable verifier gaps.

### RQ3: Does real-experience grounding improve predictive validity?

**Hypothesis:** At equal authoring and audit budgets, evaluations seeded from
real user workflows and failure traces, then expanded with audited synthetic
variants, will predict future production failures better than purely authored
or purely synthetic task sets.

### RQ4: Can evaluator changes be made without destroying longitudinal evidence?

**Hypothesis:** Versioned evaluator components plus mandatory bridge runs can
separate real system improvement from score movement caused by a changed task,
environment, harness, verifier, or aggregation policy.

### RQ5: Can evaluation taste be operationalized and partially learned?

**Hypothesis:** Models will learn some expert judgments when the target is a
specific prediction, such as task validity, marginal portfolio coverage, or
future failure prediction. They will be less reliable on unconstrained judgments
of broad research importance, especially across organizations and domains.

### RQ6: Does an evolving evaluation portfolio produce broader improvement?

**Hypothesis:** Improvements guided by a diverse, independently governed, and
experience-refreshed portfolio will survive held-out and future evaluations more
often than improvements guided by a small fixed benchmark set.

## 6. Scope

### In scope

- ingesting existing benchmark tasks and full agent trajectories;
- supporting portable task formats such as Harbor rather than replacing them;
- structural, semantic, environment, verifier, and trace-level audit;
- human adjudication workflows and disagreement records;
- generation and repair of tasks, environments, rubrics, and verifiers;
- provenance from real experience to derived synthetic variants;
- evaluator and portfolio versioning;
- frozen, hidden, live, and temporal evaluation slices;
- bridge runs and improvement receipts;
- organization-specific capability taxonomies and evaluation portfolios;
- campaigns that improve models, harnesses, or evaluators under explicit trust
  boundaries.

### Non-goals

- defining one universal measure of intelligence;
- replacing Harbor, sandboxes, experiment trackers, or training frameworks;
- claiming that every important capability has a programmatic verifier;
- fully automating domain-expert or research-taste decisions at the outset;
- allowing the optimized agent to approve its own benchmark or promotion;
- treating a larger number of benchmarks as sufficient evidence of coverage;
- autonomous model-weight training as the first implementation milestone.

## 7. Evaluation lifecycle

### 7.1 Capture intent before implementation

Every evaluation begins with a capability intent rather than a task directory:

- target user or stakeholder;
- real decision the evaluation will inform;
- desired behavior and unacceptable behavior;
- source experience, incident, workflow, or expert rationale;
- expected failure modes;
- validity conditions and known exclusions;
- expected cost and frequency of execution.

This separates the normative question of what matters from the engineering
question of how to measure it.

### 7.2 Compile intent into an executable task

An agent or evaluation engineer translates the intent into:

- instruction and input artifacts;
- pinned environment and resource requirements;
- agent interaction policy and budget;
- verifier and rubric;
- reference or oracle behavior where possible;
- health checks, negative controls, and adversarial probes;
- explicit sources of nondeterminism and external state.

Harbor can serve as the initial executable task format.

### 7.3 Validate mechanically

Before model trials, the system checks:

- schema and file references;
- image builds and dependency pinning;
- environment health and deterministic reset;
- oracle success and no-op failure;
- agent/verifier isolation and hidden-artifact leakage;
- repeatability across clean runs and, where possible, providers;
- reward production and error classification;
- secret, network, time, and external-service dependencies.

Mechanical validity is necessary but does not establish semantic validity.

### 7.4 Audit semantics through trajectories

Run several models or harnesses and inspect successful and failed traces. The
audit classifies each outcome using an explicit taxonomy:

- valid pass;
- valid capability failure;
- false pass or reward hack;
- false failure caused by the verifier;
- broken or unavailable environment;
- harness- or budget-induced failure;
- ambiguous or underspecified task;
- unsafe or policy-violating path to a passing state;
- unresolved disagreement.

Agent investigators must cite concrete task files, verifier conditions, runtime
events, and trajectory steps. A human reviewer accepts, rejects, or revises the
finding and records confidence and rationale.

### 7.5 Repair with bridge evidence

An accepted finding may produce a task, environment, verifier, or rubric patch.
The system then runs both versions on a bridge set of model-harness
configurations. The change record answers:

- which defect the patch is intended to fix;
- what score changes were predicted before running;
- which prior passes or failures changed classification;
- whether the patch introduced regressions or new exploits;
- whether model rankings changed and why;
- whether old and new scores remain comparable.

Evaluator changes never silently overwrite historical evidence.

### 7.6 Curate a portfolio

Tasks enter an evaluation portfolio through an explicit decision that records:

- capability and stakeholder coverage;
- importance and expected decision value;
- marginal information beyond existing tasks;
- difficulty and saturation risk;
- validity and exploit evidence;
- runtime and human-review cost;
- provenance and independence from the optimized system;
- owner, review date, and expiration or refresh condition.

The system should expose overlap and correlated model behavior rather than
assuming that each additional benchmark adds an independent dimension.

### 7.7 Maintain frozen and live views

Two evaluation views serve different purposes:

- **Frozen snapshots** support regressions, retention, reproducibility, and
  historical comparison.
- **Live experience streams** capture distribution change, new user workflows,
  and previously unknown failures.

Promotion should normally require acceptable performance on both: no material
regression on trusted frozen evidence and improved performance on independently
held-out or later experience.

## 8. Core objects

### CapabilityIntent

The stakeholder, decision, desired behavior, source experience, exclusions, and
rationale that explain why an evaluation should exist.

### EvalTask

An executable instruction, inputs, environment, harness policy, verifier,
reference behavior, and metadata. It may reference a Harbor task package rather
than duplicating its contents.

### EvalComponentVersion

A stable identity and digest for each independently changing component:

- task content;
- environment;
- harness and model settings;
- verifier and rubric;
- generator;
- aggregation and portfolio policy.

### EvalRun

A trial linking all component versions to raw outputs, full trajectories,
environment events, verifier evidence, cost, timing, and final score.

### AuditFinding

A typed claim about task or run validity, with cited evidence, severity,
confidence, investigator identity, human adjudication, and resolution.

### EvalPatch

A proposed component change linked to one or more findings, a pre-registered
prediction, bridge runs, review status, and migration notes.

### EvalPortfolio

A versioned collection of tasks and policies with a declared stakeholder,
intended decisions, coverage model, weights, cost budget, hidden boundaries, and
refresh policy.

### TasteJudgment

A pairwise or structured human judgment over tasks, findings, patches, or
portfolio changes. It records criteria, rationale, confidence, disagreement,
discussion, prediction, and later outcome where observable.

### ImprovementReceipt

An exportable statement linking a promoted model or harness to its baseline,
candidate, exact evaluator versions, bridge evidence, held-out results,
interventions, caveats, and approval decision.

## 9. Trust and recursion boundaries

Eval-RSI permits evaluator improvement but not unrestricted self-approval.

At minimum, the system must distinguish:

1. **Target layer:** the model, agent, harness, or policy being optimized.
2. **Eval-evolution layer:** agents that inspect traces and propose evaluation
   tasks, patches, and portfolio changes.
3. **Acceptance layer:** protected evidence and authorized humans or independent
   systems that decide whether evaluator and target changes are trusted.

The same underlying model may participate in the first two layers for
efficiency, but it must not control the final acceptance evidence. Protected
boundaries may include hidden tasks, secret verifier state, future temporal
slices, independent judge panels, production outcomes, and explicit human
approval.

If an agent can modify both the objective and the evidence used to approve that
objective, the loop measures self-consistency rather than improvement.

## 10. Relationship to vRI and Harbor

### vRI

vRI is the general improvement runtime. It treats agents as replaceable
providers and records the loop from episodes and experience through candidate
method changes, independent evidence, human or policy decisions, scoped
releases, assignments, and future use.

Eval-RSI is a domain-specific research branch built on those primitives. It
studies how the mechanisms for finding weaknesses and judging improvement can
themselves evolve. It adds evaluation semantics, audit taxonomies, portfolio
reasoning, bridge evidence, and stronger trust policies.

The integration boundary is a versioned evaluation service. A vRI improvement
run may call an Eval-RSI portfolio while treating its hidden tasks, verifier
state, and acceptance policy as protected. Eval-RSI may propose a new evaluator
version, but that version requires separate meta-evidence and cannot approve
itself. The branch can be tested independently of whether vRI succeeds as a
general-purpose runtime.

### Harbor

Harbor is a natural initial execution and interchange substrate. It defines
portable task packages, connects agents and environments, records trials and
trajectories, runs verifiers, and distributes versioned datasets.

Eval-RSI should integrate with Harbor rather than invent another task format:

```text
vRI
  system versions, episodes, candidates, evidence, releases, assignments
    |
Eval-RSI
  intent, audit, repair, taste, portfolio, evaluator evolution
    |
Harbor
  task package, environment, agent trial, verifier execution, trajectory
    |
Docker / cloud sandbox / internal runtime
```

The boundary is intentionally porous. If Harbor adds a native quality check or
task primitive, Eval-RSI should consume it. Eval-RSI is responsible only for the
continuous evidence and governance lifecycle that remains above individual task
execution.

## 11. Initial users and workflows

### Internal AI product team

Turn production traces, escalations, and human corrections into a private,
versioned evaluation portfolio. Preserve sensitive provenance and require domain
and technical approval before promotion.

### Benchmark maintainer

Run trace-aware audits across models and harnesses, triage likely broken tasks,
review cited evidence, accept patches, and publish bridge results with a new
benchmark version.

### Model or agent-harness team

Investigate why a score changed, distinguish model limitations from harness or
environment failures, propose improvements, and obtain an auditable promotion
receipt.

### Evaluation operations or data team

Convert contributions from PMs and domain experts into standardized tasks,
detect formatting and environment issues, route semantic questions to the right
reviewer, and track recurring authoring failure modes.

## 12. Minimum viable research program

The first milestone is a trace-first benchmark debugger, not autonomous
benchmark generation.

### 12.1 Pilot dataset

Select one public Harbor-compatible agentic benchmark and, if access permits,
one small internal evaluation suite. Sample approximately 20 to 40 tasks with
different environments and verifier styles. Preserve the original task versions.

### 12.2 Trial matrix

Run at least two model-agent configurations and, where feasible, vary one
harness or budget setting. Collect full trajectories, environment logs, verifier
artifacts, timing, and costs. Repeat a subset to measure nondeterminism.

### 12.3 Expert reference audit

Have a domain reviewer and an evaluation engineer independently label each
sampled outcome using the audit taxonomy. Reconcile disagreements while
recording the discussion and confidence rather than erasing the initial labels.

### 12.4 Agent-assisted audit conditions

Compare:

1. score and final output only;
2. task-package review without trajectories;
3. a generic LLM review with task and trajectory;
4. a tool-using investigator agent with the complete environment and evidence;
5. investigator agent plus human adjudication;
6. unaided expert review.

The comparison should use identical tasks and blinded outcome ordering where
practical.

### 12.5 Repair and bridge study

Allow the investigator to propose patches for accepted high-severity findings.
Run original and patched evaluator versions across the same model-harness matrix
and a small hidden neighboring set. Measure which score and ranking changes are
explained by corrected validity rather than new task difficulty.

### 12.6 Experience-grounding study

For a narrow internal workflow, compare four task sources under a matched human
audit budget:

- tasks written directly by domain experts;
- tasks generated synthetically from a broad prompt;
- tasks reconstructed from real failures or workflows;
- real-experience seed tasks expanded with synthetic variations.

Use a temporal split: build the evaluations from earlier experience and test
whether their scores predict later real failures. This is the decisive test of
the grounding hypothesis.

## 13. Metrics

### Task and verifier validity

- human-adjudicated false-pass and false-failure rate;
- high-severity defect precision and recall;
- reward-hack discovery rate;
- instruction-verifier agreement;
- oracle success and negative-control failure;
- unresolved ambiguity and reviewer disagreement.

### Reproducibility

- clean rerun success rate;
- result variance across runs, hosts, and supported providers;
- environment build and health-check failure rate;
- fraction of dependencies and external state with reproducible identity;
- time required for an independent reviewer to reproduce a finding.

### Audit efficiency

- expert minutes per task and per accepted finding;
- fraction of agent findings accepted or materially revised;
- severe defects missed by agent triage;
- time from observed anomaly to reviewed patch;
- reviewer agreement before and after evidence-grounded discussion.

### Decision usefulness

- stability of model rankings across evaluator repairs;
- ability to predict later production failures or expert judgments;
- rate of apparent model gains that survive hidden, sibling, and temporal evals;
- marginal portfolio coverage relative to runtime cost;
- number of real model or harness decisions changed by valid evidence.

### Taste learning

- held-out agreement with experts on pairwise task or patch choices;
- calibration of confidence;
- generalization across task authors, organizations, domains, and time;
- ability to predict which evaluations remain useful or become saturated;
- performance relative to simple heuristics such as difficulty, novelty, or
  current-model failure rate.

## 14. Baselines

The project should compare against serious existing practice:

- raw benchmark scores with manual spot checks;
- Harbor schema, quality-check, oracle, no-op, and reproducibility checks;
- the Harbor-Index audit-and-fix methodology;
- task-only LLM judging;
- trajectory-aware generic LLM judging;
- unaided domain-expert review;
- unaided evaluation-engineer review;
- static frozen portfolios with the same evaluation budget.

The claim is not that Eval-RSI beats an absent process. It must add value beyond
the strongest practical combination of standardized task packaging and expert
review.

## 15. Risks and failure modes

### Automated review creates false confidence

An investigator can produce plausible explanations unsupported by the actual
task or trace. Findings must cite replayable evidence, and high-impact decisions
must retain human or independent review.

### The evaluator and target collude unintentionally

Shared model priors, generators, judges, or training data can create correlated
blind spots. The system must record model provenance and preserve independent or
future evidence.

### Live evaluation destroys comparability

Continuous changes can make trend lines meaningless. Frozen snapshots and
bridge runs are required; score series must be segmented by evaluator version.

### Real experience is noisy, biased, and private

Production frequency does not equal importance. Rare safety failures may matter
more than common benign errors. Traces may contain secrets or personal data.
Ingestion therefore needs sampling policy, redaction, access control, retention,
and explicit stakeholder utility.

### Standards become bureaucracy

A schema can make task creation slower without improving validity. Required
fields should be justified by observed failure modes, and the system should
permit provisional tasks and incomplete evidence with visible status.

### Taste becomes a disguised universal metric

Organizations have different goals. Taste models and portfolio weights should
be scoped to a stakeholder and decision context, with disagreement retained.

### Synthetic expansion overwhelms real signal

Generating many variants can overweight one seed failure and create an illusion
of coverage. Portfolios should track shared provenance and cap correlated task
families.

## 16. What would falsify or narrow the proposal

The direction should be narrowed if:

- trace-aware agent review does not reduce expert time or misses severe defects;
- standardized task packaging plus existing quality checks already captures
  nearly all actionable failures;
- bridge runs do not help reviewers separate evaluator drift from target
  improvement;
- real-experience-seeded tasks do not predict later real outcomes better than
  cheaper synthetic or manually authored baselines;
- portfolio selection judgments fail to generalize beyond individual reviewers
  and cannot be made useful through explicit stakeholder scoping;
- evaluation teams will not adopt the additional provenance and review workflow;
- the dominant bottleneck is access to domain experts or protected outcomes,
  rather than tooling and coordination.

Even under these outcomes, a smaller trace-audit or reproducibility tool may
remain useful without supporting the full evaluator-evolution thesis.

## 17. Staged roadmap

### Stage 0: Workflow study

- interview or observe internal evaluation authors, PMs, domain experts,
  evaluation engineers, and result consumers;
- map current handoffs and collect concrete examples of broken or irreproducible
  tasks;
- manually produce complete audit records for a small set of trajectories;
- refine the failure taxonomy and minimum evidence model.

**Exit condition:** describe one public and one internal workflow where the
proposed objects remove a demonstrated source of invalid or unactionable
evidence.

### Stage 1: Trace-first audit vertical

- ingest Harbor tasks, trial results, and trajectories;
- run mechanical and environment checks;
- let investigator agents produce cited findings;
- support human adjudication, disagreement, and patch proposals;
- export a versioned audit report.

**Exit condition:** agent-assisted review preserves severe-defect recall while
materially reducing expert investigation time against the best available
baseline.

### Stage 2: Evaluator repair and evidence continuity

- add evaluator-component identity and versioning;
- execute patch bridge matrices;
- segment score histories by evaluator version;
- export improvement receipts and migration notes;
- enforce protected acceptance policies.

**Exit condition:** an independent reviewer can explain whether a changed score
came from the target system or the evaluation system.

### Stage 3: Experience-to-eval pipeline

- ingest and redact production or internal-workflow traces;
- cluster and prioritize candidate failures;
- compile selected incidents into replayable Harbor tasks;
- generate provenance-linked synthetic variations;
- validate against temporal holdouts.

**Exit condition:** experience-grounded evaluations predict future real failures
better than matched-cost alternatives.

### Stage 4: Portfolio and taste assistance

- define organization-specific capability maps and decision contexts;
- expose task-family redundancy and model-behavior correlations;
- collect pairwise expert judgments, rationale, confidence, and outcomes;
- train or prompt assistants to forecast task validity and marginal value;
- keep portfolio changes under human approval.

**Exit condition:** assisted portfolio decisions improve held-out decision value
or reduce expert effort without collapsing meaningful disagreement.

### Stage 5: Governed evaluator co-evolution

- allow campaigns to propose both target-system and evaluator improvements;
- allocate search budget based on validated weaknesses;
- maintain frozen, hidden, live, and temporal evidence boundaries;
- measure whether successive campaigns become better at discovering and fixing
  previously unknown failures.

**Exit condition:** evaluator evolution increases validated, transferable
improvement without degrading evidence quality, reproducibility, or human
control.

## 18. Open questions

- What is the smallest audit taxonomy that is reliable across domains?
- Which evidence can an agent approve mechanically, and which requires domain
  judgment?
- How should a task represent its intended construct and stakeholder utility?
- What is the right unit of evaluator versioning: task, component, dataset, or
  portfolio?
- How can private production outcomes validate an evaluation without exposing
  sensitive data or hidden answers?
- How should related synthetic variants be weighted so they do not masquerade as
  independent coverage?
- What bridge matrix is sufficient when evaluator changes are expensive?
- How should uncertainty and human disagreement affect promotion decisions?
- Can evaluation taste transfer across organizations, or is it inherently local?
- What trusted boundary remains when agents eventually perform most review work?

## 19. Related work and sources

- [Recursive Self-Improvement in AI](https://arxiv.org/abs/2607.07663)
  frames evaluation, verification, and direction-setting as central constraints
  on broader recursive improvement.
- [Separating signal from noise in coding evaluations](https://openai.com/index/separating-signal-from-noise-coding-evaluations/)
  reports a trace- and expert-driven audit of SWE-Bench Pro and a concrete defect
  taxonomy.
- [TASTE: Can AI Models Judge AI Safety Research Proposals?](https://alignment.anthropic.com/2026/taste/)
  studies model and expert judgments of research proposals, including discussion,
  confidence filtering, and residual human disagreement.
- [Artificial Analysis intelligence benchmarking methodology](https://artificialanalysis.ai/methodology/intelligence-benchmarking)
  is an explicit example of capability selection, exclusion, aggregation, and
  evaluator versioning as a measurement contract.
- [Harbor task format](https://github.com/harbor-framework/harbor/blob/main/docs/content/docs/tasks/index.mdx)
  provides a portable instruction, environment, verifier, and solution package
  for agentic tasks.
- [Harbor Adapters and Harbor-Index](https://arxiv.org/abs/2609.04298)
  provides a strong adjacent baseline for benchmark standardization,
  trajectory-grounded audit, and curated task repair.
- [Oak Lab mission](https://oaklab.ai/mission) and the
  [Sutton-Javed interview](https://sequoiacap.com/podcast/rich-sutton-and-khurram-javed-why-ai-models-stop-learning-and-how-to-start-it-again)
  motivate continual learning from experience and the limits of treating a
  fixed synthetic world as a substitute for external grounding.
