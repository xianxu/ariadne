`sdlc project` authors and inspects project records. A project is the lifecycle
one level above an issue: it establishes a product baseline, coordinates a
cross-repository breakdown, and derives progress from the referenced issues.

SUBCOMMANDS

  new            Create a model-derived project scaffold
  list           List live projects
  show           Print one project's baseline and task summary
  set-status     Move through the guarded lifecycle (except →done)
  validate       Validate project frontmatter against #Project
  status         Render the issue-derived progress board
  retro          Append a retrospective checkpoint
  close          Close or drop, calibrate, and archive the project
  find           Find every project record referencing an issue, fleet-wide
  throughput     Bless a measured throughput baseline, or show the current one

PROJECT FILE

Live records reside under `workshop/projects/`; terminal records archive under
`workshop/history/projects/`. Their four model-derived sections are:

  PRD          the goal, requirements, and acceptance boundary
  Estimate     the Phase-A baseline and later calibration
  Breakdown    issue-linked work rows
  Log          decisions, evidence, forecasts, and retrospectives

A retrospective is a Log entry headed `### <ISO date> — retro`. Transitioning
to `done` is intentionally unavailable through `set-status`; use
`sdlc project close`, which owns the retro and fog-factor gates.

`find --issue <ref>` (e.g. `metis#18`, `#171`) walks every fleet sibling —
active `workshop/projects/`, archived `workshop/history/projects/`, and the
deprecated brain legacy home (flagged `(legacy)`) — and prints each matching
record's path. A project can live in a different repo than the issue it
tracks; discovery is tooling's job, not a residency rule's (ariadne#171).
Read-only, lock-free. Equivalent: `sdlc resolve --kind project <ref>`.

`throughput` bridges effort to calendar (ariadne#182). Its unit is measured
focused hours/week, summed from the calibration ledger — the operator picks a
representative span, the machinery measures the rate (trailing windows skew
under vacations/crunch, so the span is blessed deliberately):

  sdlc project throughput --bless 2026-06-22..2026-07-19   # measure + record
  sdlc project throughput                                  # show current + trailing-4wk

`--bless FROM..TO` appends `{span, hours_per_week, rows, ceiling}` to
`<brain>/data/life/42shots/velocity/throughput-baseline.tsv` (append-only —
last row is current; `--ceiling N` sets the concurrent-attention ceiling,
default 2). The bare form shows the current baseline and a trailing-4-week
comparison so staleness surfaces; it never auto-substitutes the trailing
number. The forecast at `project commit` / `show` / `status` reads this
baseline.

STATUSES

{{PROJECT_STATUS_NAMES}}

{{PROJECT_LIFECYCLE}}

Project status is changed through these verbs, never by hand-editing the
frontmatter. The status set and transition graph above derive from
`construct/vocabulary/project.cue`.

For depth:

  sdlc project <verb> --help

Phase-A method:

  brain/data/life/42shots/velocity/estimate-logic-project-v1.md
