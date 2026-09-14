# Research Notes: How Close Are We to Recursive Self-Improvement?

> **Source:** Dwarkesh Podcast, “AI researchers debate how close we are to
> recursive self-improvement” (2026-09-11), featuring Beren Millidge, John
> Schulman, and Charlie O’Neill.
>
> **Document status:** Context for vRI, not evidence or a project decision.
> These are edited, timestamped notes derived from a supplied transcript, not a
> verbatim transcript. Quotations are retained where the wording matters. The
> speakers’ empirical claims and timelines have not been independently
> verified.

The discussion asks what would have to be true for rapid recursive
self-improvement (RSI) to occur, and what could prevent it. Its most useful
contribution to vRI is a sharper separation between optimizing a known
objective and discovering the right objective, together with a practical view
of how deployment experience, evaluation environments, and repeated retraining
might form an improvement loop.

## Why this matters for vRI

The episode reinforces several of vRI’s current working hypotheses without
establishing them:

- **Experience is only an input to improvement.** Deployment traces, accepted
  edits, successful trajectories, and research outcomes do not improve a
  system until they produce a method change that survives independent
  evaluation.
- **Objective specification is a protected surface.** Agents may become very
  effective at optimizing a stated goal before they become reliable at
  proposing goals, judging long-run usefulness, or noticing that an evaluator
  rewards the wrong behavior.
- **Realism and difficulty are different evaluation dimensions.** A hard,
  automatically verifiable puzzle may reveal less about useful agent behavior
  than a less tidy, multi-objective task drawn from real work.
- **Improvement may be compositional before it is continual.** Faster model
  releases, skills, adapters, context strategies, routing, and other external
  changes may deliver practical accumulation before stable online weight
  updates do.
- **Evaluator quality governs accumulation.** Weak reward signals can be
  exploited, and a narrow judge can reduce behavioral diversity. Evaluation
  changes therefore need provenance, independent checks, and controlled
  release just as agent changes do.
- **The near-term question is transfer, not raw candidate throughput.** A large
  number of generated environments or proposed changes matters only when some
  of them improve performance on held-out or future work at an acceptable
  cost.

These points align with the distinction in the [system
design](../DESIGN.md)—experience, candidates, evidence, decisions, and releases
are separate records—and with the roadmap’s insistence on a complete,
independently reviewable improvement loop.

## Episode overview

The guests approach RSI from different positions:

- **Beren Millidge**, CTO of Zyphra, focuses on generalization, plasticity,
  catastrophic forgetting, and the possibility of learning objectives.
- **John Schulman**, chief scientist at Thinking Machines and OpenAI
  co-founder, focuses on judgment, objective specification, realistic training
  distributions, and limits of natural feedback signals.
- **Charlie O’Neill**, head of model training at Baseten, focuses on signal
  generation, curriculum, continual-learning constraints, and the gap between
  executing research and choosing what research to do.

The recurring question is not merely whether an AI can improve performance
against a fixed target. It is whether an AI can repeatedly choose useful
targets, obtain genuinely new information, preserve earlier capabilities, and
judge the long-run consequences of its changes without a human outer loop.

## 00:00 — Steelmanning the case against RSI

Dwarkesh asks what technical explanation would be most plausible if the world
in 2036 had not been radically transformed by vast numbers of superintelligent
systems.

- **Millidge** proposes a Moravec’s-paradox-like outcome: systems master every
  benchmark and environment placed in front of them but never acquire robust,
  general-purpose transfer. Persistent sim-to-real gaps, weak meta-learning,
  or unsolved continual learning could produce this result. He considers it
  unlikely because current RL systems already exhibit some generalization.
- **Schulman** describes a repeating “this is AGI” cycle. A new model initially
  appears astonishing, but after sustained use its poor judgment and weak
  self-checking become obvious. Even if it writes more code than a person, the
  weakest parts of the research and engineering process can prevent a 100×
  productivity gain.
- **O’Neill** asks how close the current transformer-plus-RL recipe is to an
  optimal learning system. Like Moore’s law, a smooth-looking trend may depend
  on many discrete technical breakthroughs. If another such discontinuity is
  required, present RL environments may not be able to discover it.

Dwarkesh offers chess as a counterexample to the intuition that smooth
capability trends imply smooth effects. Engine Elo increased gradually, but
crossing the human range caused a discontinuous practical change: people went
from sometimes winning to effectively never winning. Millidge argues that a
technically quiet 2036 would require progress to asymptote just before a
similar human threshold.

### The central disagreement: where do objectives come from?

O’Neill distinguishes optimizing a known objective from identifying the right
objective. Thinking can update beliefs from available information, but “you
can’t gain any new bits from just thinking.” An AI could greatly accelerate
work once the target is clear—for example, by catching known flaws in an
experimental methodology—but that does not explain how it selects the next
valuable target.

Millidge frames rapid RSI around the same issue: can systems learn their own
objectives, optimize them, and then propose new ones without drifting? Human
autonomy may be easy for biological reasons while remaining technically hard
for AI.

Schulman predicts that defining the objective may be the last human job:
deciding how assistants should behave and what helpfulness means. He separates
alignment into objective specification and objective achievement and expects
the former to remain necessary for a long time.

## 00:18:39 — Distillation, prompt distributions, and fast followers

Schulman argues that distillation works as a counterforce to model-provider
concentration. Capabilities learned through RL may be representable by a small
amount of information and therefore be comparatively easy to distill. The hard
part is obtaining a broad, realistic prompt distribution; access to outputs or
chains of thought alone is not enough.

The speakers discuss router services as a potentially valuable source of
coding prompts and responses. Millidge notes that a follower can sometimes
generate synthetic traces from a frontier model more easily than the original
lab could collect the real-world behavior used to create the capability. He
also observes that multiple labs may be able to buy similar specialist data.

O’Neill gives model comparisons as anecdotal evidence that access to a strong
teacher and good RL environments does not guarantee a superior deployed model.
Schulman proposes two axes for environments:

1. **Difficulty:** how hard the task is to solve.
2. **Realism:** how closely the task matches multi-turn, multi-objective work.

Benchmark optimization tends to emphasize difficult but puzzle-like tasks.
Larger models may gain an advantage by generalizing from narrow, hard tasks to
realistic ones. Post-training failures can remain invisible on benchmarks.

## 00:28:06 — Training automated AI researchers

Schulman expects automated research training to combine feedback that captures
researchers’ taste with many multi-step research environments. Each iteration
would patch the most important observed weakness rather than reenact the full
history of machine learning from first principles.

O’Neill calls this “diffing the bugs”: labs operating near the frontier will
turn recent bugs and discoveries into new training environments because they
cannot afford to make a system rediscover everything through self-play.

Millidge stresses that environments can still exceed their human designers. A
designer can specify an outcome that no person can reach, allowing the learner
to search beyond human demonstrations. Schulman qualifies this by describing
real research as more than hill climbing: researchers form an intuition, build
a simplified task that might show signs of life, and only then restore realism.

## 00:33:51 — Can long-horizon RL elicit AGI?

The labs’ implied bet is that RL across a very large set of environments and
domains will produce an agent that persists, manages context, collaborates,
and completes week- or month-long work as a drop-in remote worker.

The speakers describe a progression from coding to finance, spreadsheets,
presentations, and the long tail of knowledge work. Domain-specific RL may be
theoretically redundant for a sufficiently capable on-the-job learner, but it
can make inference much more efficient. Broad domain training may also transfer
general qualities such as taste and long-horizon work habits.

### Deployment learning and the “hive mind”

Deployment produces an enormous amount of experience. Millidge argues that
this experience already feeds later model generations after filtering,
labeling, and synthesis. O’Neill points to product loops in which user actions
serve as reward signals and improved checkpoints are deployed frequently.

Schulman’s warning is directly relevant to vRI: natural feedback often lacks a
trustworthy reward function. A signal such as whether a user accepted an edit
is convenient but superficial and can be reward-hacked. Observed behavior is
not automatically valid evaluation evidence.

## 00:45:24 — The sim-to-real gap and continual learning

The discussion distinguishes cumulative technical work from non-stationary
real-world work. Once a research result such as attention or mixture-of-experts
routing is established, it becomes a durable step in a training recipe. In a
law firm or other organization, relationships, processes, and tacit knowledge
keep changing.

Schulman adds that current models show limited diversity of thought and weak
long-horizon judgment. Much of what people call taste is behavior that remains
useful over time—for example, knowing which software architecture will still
be maintainable years later. A very large context window would not solve this
by itself; the model would also need to be trained to learn from that context.

O’Neill describes a practical continual-learning failure mode. Repeated small
SFT or on-policy distillation updates can cause catastrophic forgetting and
general capability loss. RL can add capabilities through a relatively small
policy change, but it is less effective at adding knowledge.

Millidge therefore sees plasticity and catastrophic forgetting, rather than
raw capacity, as the central bottleneck. Today’s practical loop is often:
specialize, collect traces, retrain a new base model, and release again. If that
cycle compresses from months to weeks, days, or hours, it begins to approximate
continual learning at the system level even if no single model learns online.

## 01:00:33 — Data, architecture, and missing signal

Dwarkesh cites a small-scale experiment attributing more compute-efficiency
gain to data than architecture. O’Neill notes that this does not account for
the much larger gains sometimes attributed to the entire training recipe and
suggests post-training may explain a substantial part of the difference.

His more fundamental claim is about **signal scarcity**. During pre-training,
useful signal already existed in web data and the challenge was filtering it.
At the capability frontier, the decisive proof, experiment, or behavior may not
exist anywhere in the corpus. New information has to be generated through
people, environments, experiments, or deployment, and the supply of information
beyond the current frontier may be limited.

Millidge connects this to weak RL exploration: if a model cannot find a useful
trajectory in a modest rollout budget, it receives no learning signal. A
curriculum must keep the distance between successive rungs small enough. The
speakers cite examples in which relatively weak models can imitate expert
behavior once shown a suitable trajectory, while direct RL fails when the
starting capability is too far below the target.

The parameter-scaling discussion highlights a changing optimization regime.
When data, rather than compute, is the binding constraint, the most
data-efficient model size and sparsity may differ from the most
compute-efficient configuration. The guests do not claim that current scaling
choices will remain optimal.

## 01:18:03 — Why does RL work as well as it does?

The puzzle is that an RL episode may provide roughly one bit of reward, yet
post-trained models can improve qualitatively.

Millidge attributes much of the apparent gain to mid-training. High-quality
synthetic reasoning data warm-starts a model close to the eventual policy, and
RL then provides a high-signal correction. SFT asks the learner to reproduce
every detail of a teacher’s trajectory, including incidental choices; RL can
communicate only whether the outcome was good. A small number of bits can still
eliminate a large portion of the hypothesis space.

O’Neill distinguishes **horizontal generalization** from **horizon
generalization**. Training on mathematics does not necessarily make a model a
better coder, but training across many environments may teach it to use more
tokens productively and sustain progress for longer. Many task-level phase
transitions, averaged together, can look like a broad qualitative jump.

## 01:24:54 — Move 37, diversity, and entropy collapse

The speakers separate two meanings of creativity:

- Solving a difficult search problem under many constraints, where AI systems
  can discover surprising solutions.
- Maintaining a wide distribution of styles, hypotheses, and approaches.

Schulman argues that RL can improve the first while harming the second.
Post-trained systems develop repeated stylistic habits, and widespread
distillation from the same frontier models can create a monoculture across
open-weight models.

Millidge sees entropy collapse as primarily a data and evaluator problem rather
than an unavoidable property of RL. A simplistic verifier rewards a narrow,
hackable mode; a richer judgment signal could preserve more useful diversity.

## 01:28:32 — Rapid-fire timelines

The estimates are speculative and depend strongly on definitions.

| Milestone | Charlie O’Neill | Beren Millidge | John Schulman |
| --- | --- | --- | --- |
| Drop-in remote worker across white-collar work, with general computer use and month-long operation | About 1 year if browser restrictions and communication access are relaxed | About 3 years for full generality; organizations will adapt their workflows and reach 80–90% sooner | Some lower-quality work is already covered; a useful version within about 1 year |
| 10× productivity uplift for AI researchers | 5–10 years; the bottleneck is absorbing information and choosing the Bayesian-optimal next step | A visible path in about 2 years; coding is already highly accelerated, and a few autonomous experimental loops would matter greatly | About 2 years, while rejecting a single scalar estimate across all research work |
| Better than top human experts at all computer-mediated work | 5–10 years; automating AI research may itself be “ASI-complete” | About 5 years for heavily targeted domains, longer for the neglected long tail | 3–4 years for computer-based work; embodied engineering and long-horizon workplace learning take longer |

The structure of the disagreement is more informative than the numbers. The
guests are relatively close on a useful remote worker and differ most on
automated research. Their uncertainty concentrates on whether systems can
absorb new information, acquire taste, and select the next useful objective.

## Consolidated insights

### 1. Objective invention is a different capability from objective optimization

All three guests converge on this distinction from different directions.
Within a well-specified objective, large acceleration appears plausible.
Reliable objective invention without a human outer loop requires additional
capabilities: seeking new information, forming useful abstractions, judging
long-run effects, and noticing specification failures.

For vRI, this suggests treating an agent-proposed objective as a candidate that
requires evidence and authorization, not as the next step of an automatically
trusted loop.

### 2. Deployment distributions may matter more than access to a teacher

Distillation can reproduce behavior only over the prompts and situations it
sees. A realistic deployment distribution can therefore be a stronger asset
than nominal access to a more capable model. This increases the value of
recording task distribution and context, not just successful trajectories.

### 3. System-level accumulation may arrive before online model learning

Frequent retraining, adapters, skills, context strategies, and routing can make
the deployed system improve continuously even when one set of model weights
cannot. This supports vRI’s decision to define the improvement target as a
versioned agent system rather than only a model.

### 4. Mid-training and curriculum complicate attribution

If RL succeeds only after synthetic mid-training and carefully spaced
curriculum steps, attributing a gain to “RL” is misleading. Evidence should bind
the full recipe: starting model, data, environments, rollout policy, evaluator,
and selection rule.

### 5. The scarce resource may be informative signal

Compute and raw data are not enough when the missing knowledge has not yet been
produced. Improvement systems need mechanisms for generating information:
experiments, interaction with reality, diverse critique, and well-designed
environments. Candidate generation without new signal will eventually saturate.

### 6. Diversity is an evaluation target, not a cosmetic property

A system can score better while becoming more brittle or monocultural. When
diverse hypotheses matter, evaluations should measure coverage and correlated
failure rather than relying only on average task reward.

## Caveats

This was a discussion among practitioners, not an empirical report. Claims
about named models, training-data practices, benchmark results, efficiency
decompositions, and laboratory behavior were presented as personal
observations or second-hand reports. The cited data-versus-architecture result
was described as a small-scale experiment whose scale dependence is unknown.
The timelines are subjective estimates.

The most defensible use of these notes is therefore to generate testable
hypotheses for vRI—not to use the episode as evidence that those hypotheses are
true.
