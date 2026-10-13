`sdlc project` authors and inspects project records. A project is a TL-driven
coordination plane for cross-issue work (#329): a Charter (problem, cross-issue
requirements, scope), Trains (chains of issues, one slot each, plus the hard ⇒
edges between them) and append-only Decisions. It carries no deadline: the
forecast is derived from the trains and the calibration ledger, never typed, and
a date a team needs is set against the forecast, never by it.

SUBCOMMANDS

  new            Create a model-derived project scaffold
  list           List live projects (name, status, TL)
  show           Print one project's frontmatter, task summary and forecast
  set-status     Move through the guarded lifecycle (except →closed)
  validate       Validate frontmatter against #Project and the model's sections
  status         Render the issue-derived progress board and forecast
  retro          Append a retrospective checkpoint
  close          Close or drop the project and archive it
  find           Find every project record referencing an issue, fleet-wide
  forecast       Forecast remaining focused hours from the trains (read-only)

PROJECT FILE

Live records reside under `workshop/projects/`; terminal records archive under
`workshop/history/projects/`. Their model-derived sections, in order:

  Charter      problem, cross-issue requirements, done_when, scope
  Trains       the trains table and the cross-train edge list
  Breakdown    issue-linked checkbox rows
  Log          evidence, forecasts and retrospectives
  Decisions    cross-issue decisions only: dated, who decided, one line of why

The Trains grammar:

  | Train | Cars, in order | Slot |
  |---|---|---|
  | Main integration | ariadne#183 → #269 → #303 | ariadne:4 |

  - ariadne#183 and ariadne#320 ⇒ ariadne#269/#303 (why)

A bare `#N` inherits the repo of the car before it (the project's repo for the
first). Parenthesized text is an annotation. Every separator (→ ⇒ / ,) in a cars
cell means "next car": a train runs serially. An edge bullet holding ⇒ makes
each ref before it a prerequisite of each ref after it; prose after a colon is
the reason. `~` bullets are overlap notes the forecast ignores.

A retrospective is a Log entry headed `### <ISO date> — retro`. Transitioning
to `closed` is intentionally unavailable through `set-status`; use
`sdlc project close`, which owns the retro gate and the archive move.

`find --issue <ref>` (e.g. `metis#18`, `#171`) walks every fleet sibling —
active `workshop/projects/`, archived `workshop/history/projects/`, and the
deprecated brain legacy home (flagged `(legacy)`) — and prints each matching
record's path, with the issue's train and car position when a train carries
it. A project can live in a different repo than the issue it tracks; discovery
is tooling's job, not a residency rule's (ariadne#171). Read-only, lock-free.
Equivalent: `sdlc resolve --kind project <ref>` (whose JSON carries `train` and
`car`).

`forecast` reads each car's card (status, `estimate_hours`) and the calibration
ledger. Expected hours per unlanded car = estimate × Σ(build+converge)/Σestimate
over the ledger's split rows (the estimate ÷ build rate × (1 + converge k) of
#234, aggregated); the range is the 25th–75th percentile of the per-issue
multipliers (the point alone below 3 rows). Trains run in parallel, each
serially, and ⇒ edges hold a car until its prerequisite lands, so the result is
the critical path as a range of focused hours, plus the total remaining. An
unestimated car takes the median ledger estimate, named in a note. `show` and
`status` print the same line best-effort.

STATUSES

{{PROJECT_STATUS_NAMES}}

{{PROJECT_LIFECYCLE}}

Project status is changed through these verbs, never by hand-editing the
frontmatter. The status set and transition graph above derive from
`construct/vocabulary/project.cue`.

For depth:

  sdlc project <verb> --help
