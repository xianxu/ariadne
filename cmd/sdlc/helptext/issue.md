`sdlc issue` is the CRUD/authoring surface for the issue *record*. It
complements the flat checkpoint verbs (`close`, `claim`, `change-code`, `pr`,
`merge`, `milestone-close`) — those guard workflow *transitions*; `issue *`
edits the record itself.

SUBCOMMANDS

  new            Create a new issue from the canonical template (allocates the
                 next ID; `--from-github N` seeds it from a GitHub issue)
  sync           Commit this issue's body (Spec/Plan/Log) under a message that
                 names it; `--push` to also publish
  set-status     Flip an issue's status with transition guards
  list           List issues (ID, status, title), sorted by ID; --status filters
  show           Print an issue's frontmatter + section headers (no bodies)

CANONICAL ISSUE FILE

`sdlc issue new` writes the template below; its section shape is DERIVED from
`construct/vocabulary/issue.cue` (`scaffold.sections`) via `pkg/vocab` (#145) —
that model is the single source of truth, and this doc is the human reference
(a drift-tested superset: it may list optional sections the skeleton omits).
Filename: `workshop/issues/NNNNNN-<slug>.md` (zero-padded
6-digit ID, kebab-case slug). Keep the slug to <5 words: it becomes the git
branch name verbatim, and the branch feeds the orientation slug's left segment.

  Frontmatter (in order):
    id             zero-padded 6-digit, matches the filename
    status         {{STATUS_NAMES}}
    deps           list of dependency refs, e.g. [repo#1, repo#2]
    github_issue   GitHub issue number when mirrored, else empty
    target         (optional) a workshop/targets/ slug
    created        ISO date
    updated        ISO date (bumped on status changes)
    started        ISO-8601 stamp at the open→working flip (#116); the
                   active-time window anchor — set by the verb, never
                   hand-edited
    estimate_hours derived after the plan clears plan-quality; required by change-code — not at
                   claim (#113). Optional at create.
    actual_hours   (added at close) required when status → done: number or N/A

  Body sections (in order):
    # <Title>
    ## Problem      what's wrong / what's needed
    ## Spec         desired behavior, constraints, design decisions
    ## Done when    acceptance criteria (≥1 non-empty bullet)
    ## Estimate     fenced ```estimate block deriving estimate_hours by
                    v2-lineage primitive; change-code reconciles it (#117 —
                    see `sdlc change-code --help` / helptext/estimate.md)
    ## Plan         checkable steps; `## Plan` milestones gate reviews
    ## Log          dated session notes (### YYYY-MM-DD)
    ## Side quests  (optional) unplanned work that landed

STATUS

{{STATUS_GLOSS}}
Flip status with `sdlc issue set-status` (or `sdlc claim` to start work), never
by hand-editing the frontmatter — the verbs carry the transition guards.
(`done` closes via `sdlc close`.) The status set above is derived from the model.

COMMITTING THE BODY (#206)

`sdlc claim` and `sdlc issue new` publish the *reservation* — an ID and a name.
That is the whole external contract: peers need to know an issue exists and is
taken, not what is in it. Nothing commits the body that follows, so the entire
planning phase — the longest phase, and the one that produces the design — sits
in the working tree until an unrelated verb sweeps it up, and a compaction or a
closed terminal loses it.

`sdlc issue sync --issue N` is that missing half. Reach for it whenever the
design has moved: after a brainstorm lands, after a decision you'd hate to
re-derive, before a long-running tool call, before a context checkpoint.

It does NOT push. Durability and publication are separable, and only durability
is missing mid-planning: the default is a local commit in the current worktree,
on the current branch, with no network — cheap enough to run often, and safe
because a half-written Spec cannot escape. Publication stays with the verbs that
already own an external boundary; `sdlc change-code` runs the same sync WITH the
push once plan-quality has accepted the design, and `--push` is there for other
milestone callers.

For depth:

  sdlc issue <verb> --help
