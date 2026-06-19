Multi-agent benchmark harness (#119): run different coding agents (claude,
codex, …) against the SAME frozen task in isolated worktrees, then grade each by
measured objective signals plus blind head-to-head LLM/operator judging.

The harness is base-layer; the task library + results are repo-local under
`workshop/benchmarks/` (see its README and the `benchmark-task` datatype).

PIPELINE

  freeze   snapshot a live issue → an immutable task (spec + base SHA + rubric)
  run      fan out N agents into isolated worktrees, per mode, never merging
  grade    Stage A measure (build/tests/artifacts/metrics) + Stage B blind judge
  review   fold in the operator's rankings, de-anonymize, record both verdicts
  report / leaderboard   aggregate run records

A frozen task pins `base_sha`; agents branch from it and never merge, so a task
replays identically forever — even after main advances. Effort is compared by
wall-clock/turns (not `sdlc actual`, which is Claude-transcript-only).

  sdlc bench <subcommand> --help    for detail
