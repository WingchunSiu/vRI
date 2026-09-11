# Engineering Principles

This file applies to the entire repository.

vRI is an ongoing, pre-stable project. Its design will change as we learn from
real use. Prefer a correct, coherent design over compatibility with an early
abstraction. Breaking changes are acceptable when they materially improve the
product or simplify the system.

## Principles

1. **Start from the problem.** Work from first principles and observed user
   needs. State what problem a change solves, for whom, and how success can be
   checked. Treat assumptions as hypotheses rather than requirements.
2. **Make the whole path work.** Build one realistic flow from input to useful
   outcome before expanding its breadth. Keep that path reliable as the system
   evolves.
3. **Prefer simple solutions.** Add the smallest mechanism that addresses an
   observed failure. Introduce abstractions for real boundaries or concrete
   second uses, not hypothetical flexibility.
4. **Optimize for users.** Functionality includes whether behavior is useful,
   understandable, and recoverable when something fails.
5. **Prefer evidence over confidence.** Check the smallest useful version,
   compare with a realistic baseline, and preserve enough information to
   explain important results and failures.
6. **Test behavior.** Automated tests should cover public behavior, invariants,
   and meaningful failure paths without merely copying the implementation.
7. **Write clear code.** Choose precise names, follow existing style, and use
   comments to explain constraints and tradeoffs rather than restating code.
8. **Keep documentation current.** Update relevant design documents, examples,
   and public contracts when behavior or terminology changes.

Most near-term gains will come from sound engineering, clear feedback, and a
solid end-to-end pipeline. Do not confuse technical novelty or sophisticated
machine learning with product value. Add complexity only after simpler
approaches reach a demonstrated limit.

## Breaking Changes

Do not preserve a poor abstraction solely for backward compatibility during
this pre-stable phase. When making a breaking change:

- explain which assumption changed and why the new design is better;
- update callers, tests, examples, schemas, and documentation together;
- describe effects on persisted state, configuration, and user workflows;
- provide an explicit migration or reset path when user data is involved;
- remove obsolete compatibility layers instead of keeping competing designs.

Breaking changes are permission to improve the design, not permission to lose
data or leave a partial migration.

## Working Agreement

- Read the relevant code and documentation before editing.
- Inspect and preserve unrelated user changes in the worktree.
- Prefer small, cohesive changes and working vertical slices.
- Test in proportion to risk.
- Report what was verified, what remains assumed, and what is still unknown.

The standard for completion is not that code exists. The relevant user path
should work and the result should be easier to understand and maintain.
