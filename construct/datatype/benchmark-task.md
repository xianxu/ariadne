---
type: type
name: benchmark-task
description: An immutable, replayable benchmark task frozen from a live issue — a spec snapshot + base SHA + grading rubric that different coding agents are run against by the `sdlc bench` harness. Triggers on "freeze this issue into a benchmark task", "/xx-datatype benchmark-task", or editing a workshop/benchmarks/tasks/*.md file. Distinct from an issue (mutable work record) — a benchmark-task is frozen and never changes.
---

# benchmark-task

A **benchmark-task** is the immutable definition of one controlled experiment for
the `sdlc bench` harness (#119). It is frozen from a live backlog issue at a point
in time and never changes thereafter, so any coding agent can be replayed against
the exact same starting conditions even after `main` has advanced.

Lives in: `workshop/benchmarks/tasks/<id>.md` (repo-local; the `sdlc bench`
harness itself is base-layer, but each repo curates its own task library).

## Frontmatter

- `type: benchmark-task`
- `id` — slug, usually `<issue>-<short-slug>`; matches the filename
- `repo` — repo the task is frozen from
- `source_issue` — the live issue number it was frozen from
- `base_sha` — the immutable commit agents branch from (the reproducibility anchor)
- `created` — ISO date

## Body

- `## Spec` — the issue's `## Spec`, copied **verbatim**; the constant prompt
  context handed to every agent.
- `## Config` — a fenced ` ```json ` block: `{ "setup": [...], "rubric": {...} }`.
  - `setup` — shell commands to bring a fresh worktree to green before the agent runs.
  - `rubric` — `objective` checks (measured) + `subjective` dimensions (judged), each
    tagged `group` (`quality` | `workflow-fit`) and `weight`. See `DefaultRubric()`.

## Invariants

- Immutable after `sdlc bench freeze`. To change grading, freeze a new task.
- `base_sha` must remain reachable; the harness branches from it and never merges.
- The verbatim spec must not contain a top-level `## ` heading — `freeze` extracts
  and round-trips sections by `^## ` boundaries, so a `## ` inside the spec would
  truncate it. `freeze` warns when the source spec contains a `## ` line; use `###`
  subheadings in issues meant to be frozen.
