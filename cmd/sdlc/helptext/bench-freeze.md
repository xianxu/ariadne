Snapshot a live issue into an immutable benchmark task.

`freeze` copies the issue's `## Spec` verbatim, pins `base_sha` to the current
HEAD, scaffolds the default rubric, and writes
`workshop/benchmarks/tasks/<issue>-<slug>.md` (datatype `benchmark-task`). The
task never changes after this — to change grading, freeze a new task.

  sdlc bench freeze --issue 119

FLAGS

  --issue <n>   the live issue to freeze (required)
  --repo <name> repo name recorded on the task (default: ariadne)

NOTES

  The agents branch from `base_sha` and never merge, so the frozen task replays
  identically even after main advances. A spec containing a top-level `## `
  heading triggers a warning (it can truncate on round-trip — use `###`).
