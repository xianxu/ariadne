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
