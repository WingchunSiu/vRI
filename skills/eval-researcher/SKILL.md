---
name: eval-researcher
description: Investigate what an evaluation should measure, build or repair executable checks when useful, attribute outcomes from raw evidence, and retain scoped experience for later eval work. Use for open-ended evaluation research, not routine benchmark execution with an already fixed task and interpretation.
---

# Eval Researcher

Turn an evaluation intent into a supported finding, repair, or benchmark
candidate. Preserve enough evidence for another researcher to check the
conclusion, and retain experience only when it may improve later work.

## Establish the investigation

- Read the user's intent, source scope, budget, permissions, and release
  authority. Keep permission to run an experiment separate from approval of
  its output.
- Query relevant prior material with `vri context query`. Open raw evidence
  when a summary is insufficient; do not treat retrieved experience as fact.
- Form a scoped, falsifiable question and choose the cheapest next action that
  can materially change the answer.

## Choose tools from the question

Use Git history, native sessions, documents, helper agents, or direct analysis
when they provide the needed evidence. Use Harbor when an executable task,
environment, verifier, or model comparison is the useful check. Use Herdr for
persistent processes, helpers, and waiting when those capabilities help the
investigation. Do not impose a fixed tool order.

Inspect important trajectories and artifacts rather than inferring causality
from rewards alone. Distinguish model capability, agent harness, environment,
task, and verifier failures. Preserve uncertainty when the evidence does not
separate them.

## Preserve evidence and experience

- Capture consequential raw files or directories with `vri evidence capture`.
  Record partial or unknown completeness honestly.
- Cite vRI record IDs or object digests in findings. A mutable path alone is
  not a durable citation.
- Propose experience with `vri experience propose` only when the interpretation
  is likely to change a later decision or action. State its scope and link the
  evidence that supports it.
- Use `vri experience revise` to activate, contest, supersede, or retire an
  interpretation. Revisions append new versions; do not rewrite history.

## Report the outcome

Give the user the conclusion, evidence references, material limitations, and
the next decision that needs human judgment. Keep generated artifacts and
scratch work separate from accepted benchmarks or released methods.
