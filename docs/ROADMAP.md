# vRI Roadmap

This roadmap tracks the next uncertainty to reduce. It is not a feature
inventory or calendar. A step is complete only when its user path and research
claim have been checked.

## Current state

The repository has two grounded pieces:

- A local study turned one real verifier repair into a Harbor task, four runs,
  trajectory-level attribution, and a review-ready benchmark candidate. The
  candidate still awaits human review and does not demonstrate autonomous
  research or learning.
- The Go skeleton preserves content-addressed evidence, records and relations;
  queries and materializes context; and retains scoped, evidence-backed,
  revisable experience. The eval-researcher skill describes the initial agent
  method.

The system has not yet run a complete investigation as a persistent Eval
Researcher in Herdr. That is the current product gap.

## Now: one Herdr-managed investigation

### Claim

An Eval Researcher using vRI context and experience can complete one useful
evaluation investigation with less manual history navigation and coordination
than the same agent using native sessions and Harbor directly.

### Build only the path this test needs

1. Run a normal supported agent in a persistent Herdr workspace with a clear
   intent, source scope, budget, permissions, and the eval-researcher skill.
2. Let it query and open selected prior evidence, form a falsifiable question,
   and choose its own investigation.
3. Let it call Git, native session stores, helper agents, and Harbor directly.
   Harbor is an experimental tool, not the workflow controller.
4. Capture consequential inputs and outputs, inspect important trajectories,
   and produce a scoped finding with durable evidence references.
5. Retain experience only when it is expected to change later work.

Fix a storage, Harbor, or Herdr boundary only when the real path reaches it.
Do not add a universal agent state machine, provider registry, semantic trace
schema, vector database, or Web UI for this step.

### Check

Use the same base agent, source access, tools, permissions, and approximate
budget for the practical baseline. Give the baseline native history and Harbor
but no vRI retrieval or retained experience.

Record:

- time and user effort needed to locate relevant evidence;
- agent and human interventions;
- important task, environment, harness, and verifier defects;
- whether the conclusion can be reproduced from cited artifacts;
- elapsed time, compute, and model cost;
- unresolved uncertainty and the decision that followed.

This step succeeds when the researcher produces a useful, independently
traceable result and measurably reduces manual navigation or coordination. A
polished report alone is not success.

## Next: harden observed evidence boundaries

The current implementation already exposes several concrete correctness gaps:

- bind a Harbor run to the task digest and Harbor version captured at execution
  time rather than a mutable source directory;
- preserve incomplete and infrastructure-failed jobs;
- retain unexpected artifacts instead of relying on a closed allowlist;
- represent repeated trials without overwriting outcomes;
- make accepted run evidence portable without its original directory;
- add indexed retrieval only when corpus size, latency, or missed evidence
  demonstrates that scan-based search is insufficient.

The store is trustworthy for the exercised path when it can be copied or moved
and still reproduce the relevant task, run identity, evidence, experience, and
decision.

## Then: test retained-experience transfer

Freeze the researcher method and run a second, related investigation on source
material not used to create or select the retained experience. Compare the
stateful researcher with the same system under a stateless or previous-method
condition at a matched budget.

Retained experience counts as learning only if it improves later work: quality,
attribution, expert correction, time, or cost. More stored context or better
recall of the first case is not enough.

## Later, only if the evidence supports it

- Let the researcher propose changes to its own skills, context policy, tools,
  or delegation method, then evaluate frozen candidates on protected work.
- Call this recursive improvement only when a validated system change improves
  the production of future validated changes.
- Try a materially different target only after the eval-researcher vertical
  works; extract shared contracts from two real uses instead of generalizing in
  advance.

## Redirect or stop if

- retrieval does not beat direct search over native sessions;
- the user must specify nearly the entire investigation for the agent;
- Harbor plus an ordinary prompted agent reaches the same outcome with similar
  effort and reliability;
- retained experience does not improve later work under a fair comparison;
- experts must redo most of the investigation to trust it;
- evidence capture loses important provider-native information or makes
  privacy and retention unacceptable.
