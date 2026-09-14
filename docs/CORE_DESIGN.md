# vRI v0 Storage Contract

> Status: this document describes the implemented Go storage and query
> skeleton. It is not the product architecture. The local study that motivated
> it is ready for review, not approved or released, and intentionally untracked.
> See [DESIGN.md](DESIGN.md) for
> the eval-researcher system and [TECHNICAL_PLAN.md](TECHNICAL_PLAN.md) for the
> next slice.

## 1. Why this skeleton exists

An eval researcher needs evidence and experience that survive one agent
session. The smallest current mechanism preserves files by content, stores
lightweight interpretations and decisions, links them, imports Harbor results,
and exposes the first context and experience operations to an agent.

This is deliberately narrower than vRI's intended product. Today the agent did
the motivating research manually using ordinary repository and Harbor tools.
The Go code can now capture and search evidence and retain revisable
experience, but it does not host an agent, run Harbor, diagnose a trace, or
improve a method autonomously.

```text
files and source material ──▶ objects + records + relations
Harbor job directories ─────▶ copied artifacts + run metadata
manifest ───────────────────▶ run matrix, review view, release validation
```

## 2. Storage model

A store is local to a project:

```text
.vri/
├── vri.db
└── objects/
```

It contains three tables:

```sql
CREATE TABLE objects (
  digest     TEXT PRIMARY KEY,
  kind       TEXT NOT NULL,
  path       TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE records (
  id         TEXT PRIMARY KEY,
  type       TEXT NOT NULL,
  payload    TEXT NOT NULL,
  actor      TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE relations (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  src_kind   TEXT NOT NULL,
  src_id     TEXT NOT NULL,
  dst_kind   TEXT NOT NULL,
  dst_id     TEXT NOT NULL,
  type       TEXT NOT NULL,
  meta       TEXT,
  created_at TEXT NOT NULL
);
```

- **Objects** are immutable files or directory archives addressed by SHA-256.
- **Records** are append-only JSON documents addressed by ULID.
- **Relations** connect records and objects. The application verifies that both
  endpoints exist before writing a relation.

The tables are generic on purpose. `kind`, `type`, and relation names are
conventions used by current commands, not a closed ontology that every future
artifact must adopt.

## 3. Evidence policy

Raw evidence should remain in its native form whenever possible. A trajectory,
verifier log, screenshot, or unexpected error file does not need to become a
fully normalized record before an agent can inspect it.

The intended policy is:

1. preserve the raw artifact or a stable reference to it;
2. attach the minimum metadata needed for identity, provenance, scope, and
   retrieval;
3. store an interpretation separately so it can be revised without rewriting
   the evidence;
4. add stronger structure only when execution, comparison, access control, or
   release needs it.

The current Harbor importer is stricter than this policy. It copies a fixed
artifact allowlist: job and trial configs/results, files directly inside each
`verifier/` directory, and an optional `agent/trajectory.json`. That allowlist
is a v0 implementation limitation, not a statement that other artifacts are
unimportant.

## 4. Current vocabulary

The local study uses these record types:

| Type | Current purpose |
| --- | --- |
| `source` | Imported source material and provenance |
| `experience` | A scoped, revisable interpretation linked to evidence and a separate body artifact |
| `task_candidate` | A proposed measurement task |
| `selection` | A recorded choice among candidates |
| `harbor_job` | Parsed job metadata, trials, rewards, and artifact digests |
| `claim` | A falsifiable statement linked to supporting or conflicting evidence |
| `benchmark_manifest` | A candidate benchmark definition |
| `decision` | A reviewer outcome and scope |

Current relations include `captures`, `has_body`, `supported_by`, `revises`,
`extracted_from`, `includes`, `evaluates`, `supports`, `contradicts`, and
`decided_by`. New evidence and experience do not need new tables; they may
introduce new record or relation types when an actual query or invariant
requires them.

## 5. Implemented CLI

```text
vri init
vri import <path>
vri evidence capture <path>
vri context query|open|materialize
vri experience propose|revise
vri object put|get|list
vri record put|get|list
vri ingest harbor-job <dir>
vri runs <manifest-file-or-id>
vri review
vri release [--reviewer name] [--outcome value] <manifest-file>
```

Behavior today:

- `init` creates the SQLite store and object directory.
- `import` copies source material into the object store and writes a `source`
  record.
- `evidence capture` preserves a file or directory and records a minimal source
  envelope with provenance and completeness.
- `context query` searches record metadata and captured textual content;
  `open` resolves a record or object to raw content; `materialize` creates a
  disposable working set with a manifest back to its source references.
- `experience propose` requires at least one scope and one existing evidence
  reference and stores its body separately; `revise` appends a new version with
  a status and a `revises` relation instead of mutating the earlier
  interpretation.
- `object` and `record` expose the generic storage primitives.
- `ingest harbor-job` parses a completed Harbor job, copies its known semantic
  artifacts, writes one `harbor_job` record, and links it to any matching task
  object already in the store.
- `runs` renders a task-by-job reward matrix from ingested jobs. It accepts a
  manifest file or an already stored manifest record ID.
- `review` renders candidates, claims and evidence links, run summaries, and
  manifest limitations from records already present. It does not independently
  inspect trajectories or derive causal attribution.
- `release` accepts a manifest file, verifies current task objects and Harbor
  job directories, and records a review decision. Only `approve` creates the
  `decided_by` release binding; `reject` and `more-evidence` remain non-release
  decisions.

Commands that return structured results generally support `--json`; `record
put --json` instead uses that flag for its input payload, and `object get` emits
the object itself.

## 6. Implemented invariants

- File and directory identity is content-addressed.
- Directory digests use sorted `(relative path, file digest)` entries.
- Relation endpoints must exist when a relation is added.
- Re-ingesting the same Harbor source path is idempotent.
- Release refuses a missing task object, changed job-directory digest, or
  manifest run without a matching ingested job record.
- Synthetic tests reconstruct multiple jobs as a comparison matrix.
- Experience cannot be proposed with a missing or empty evidence set.
- Removing an original source does not prevent captured evidence from being
  queried, opened, or materialized.
- Experience revisions preserve the previous record and body.
- Default context queries return the head of an experience revision chain and
  omit heads marked `superseded` or `retired`; history remains addressable.

These invariants are covered by behavior tests. They establish only what the
implementation checks; they do not establish task importance, semantic
validity, or benchmark approval.

## 7. Known gaps

These are observed correctness gaps, not speculative future features:

1. **Task identity:** Harbor ingest recomputes a digest from the current task
   source path. A later edit can detach the run from the exact task version
   Harbor executed. The runtime lock or packaged task snapshot should be the
   authority.
2. **Partial jobs:** ingest requires job-level `result.json` and a readable
   verifier directory. Interrupted or infrastructure-broken runs may contain
   the most useful evidence and must remain ingestible as incomplete evidence.
3. **Multiple trials:** the reward matrix maps one task ID to one reward per
   job, so repeated trials overwrite one another instead of supporting
   pass@k, uncertainty, and failure distributions.
4. **Mutable source directories:** release re-hashes live Harbor directories
   and therefore depends on their continued presence and byte-for-byte state.
5. **Artifact selection:** the fixed allowlist can omit unexpected evidence and
   includes incidental files such as `__pycache__` when they appear directly in
   selected directories.
6. **Retrieval scale:** context query currently scans the small local store and
   has no ranking or rebuildable FTS index. Context materialization is present,
   but method comparison and automatic context-policy evaluation are not.

7. **Researcher runtime:** the repository includes an initial eval-researcher
   skill, but no Herdr launcher or adapter and no completed autonomous vertical.

Fixing the first five makes the existing Harbor evidence boundary trustworthy.
The scan-based retrieval and initial researcher skill now make a thin
Herdr-managed investigation the next product slice; it should drive further
adapter and storage work from observed failures.

## 8. Local-study relationship

The motivating study contains source provenance, a curation log, a Harbor
task, four job directories, an evidence packet, a candidate manifest, and a
pending human review. It was useful for discovering the three-table shape and
the gaps above, but its raw artifacts contain local paths and potentially
private execution context. They remain outside the public repository.

The study does not prove that its task represents the user's desired model
comparison. Authorization to spend compute on the Harbor runs was not explicit
approval of the candidate or its inclusion in a benchmark. Until review is
recorded, it remains a local review-ready candidate.
