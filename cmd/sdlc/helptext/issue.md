`sdlc issue` is the CRUD/authoring surface for the issue *record*. It
complements the flat checkpoint verbs (`close`, `claim`, `change-code`, `pr`,
`merge`, `milestone-close`) — those guard workflow *transitions*; `issue *`
edits the record itself.

SUBCOMMANDS

  new            Create a new issue from the canonical template (allocates the
                 next ID; `--from-github N` seeds it from a GitHub issue)

CANONICAL ISSUE FILE

`sdlc issue new` writes — and this is the single source of truth for — the
template below. Filename: `workshop/issues/NNNNNN-<slug>.md` (zero-padded
6-digit ID, kebab-case slug).

  Frontmatter (in order):
    id             zero-padded 6-digit, matches the filename
    status         open | working | blocked | done | wontfix | punt
    deps           list of dependency refs, e.g. [repo#1, repo#2]
    github_issue   GitHub issue number when mirrored, else empty
    target         (optional) a workshop/targets/ slug
    created        ISO date
    updated        ISO date (bumped on status changes)
    estimate_hours optional at create; required when status → working
    actual_hours   (added at close) required when status → done

  Body sections (in order):
    # <Title>
    ## Problem      what's wrong / what's needed
    ## Spec         desired behavior, constraints, design decisions
    ## Done when    acceptance criteria (≥1 non-empty bullet)
    ## Plan         checkable steps; `## Plan` milestones gate reviews
    ## Log          dated session notes (### YYYY-MM-DD)
    ## Side quests  (optional) unplanned work that landed

STATUS

`open` not started · `working` in progress · `blocked` waiting · `done`
complete (close via `sdlc close`) · `wontfix` rejected · `punt` deferred.
Flip status with `sdlc set-status` (or `sdlc claim` to start work), never by
hand-editing the frontmatter — the verbs carry the transition guards.

For depth:

  sdlc issue <verb> --help
