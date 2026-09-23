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
                   claim (#113) — on the full flow only. Optional at create.
    flow           the issue's SDLC flow, one line (#231):
                   {kind: full|quick, provenance: inferred|operator} — plus, on
                   quick, the quoted contract hashes `spec`/`done`. Written by
                   `sdlc change-code` (inferred: Mx milestones or a design past
                   the shell's design limit → full, neither → quick; `--flow`
                   pins it) and by `sdlc close` (a quick issue outside the shell
                   → full). Never hand-edited;
                   absent reads as full.
    actual_hours   (added at close) required when status → done: number or N/A

  Body sections (in order):
    # <Title>
    ## Problem      what's wrong / what's needed
    ## Spec         desired behavior, constraints, design decisions
    ## Done when    acceptance criteria (≥1 non-empty bullet)
    ## Estimate     fenced ```estimate block deriving estimate_hours by
                    v2-lineage primitive; change-code reconciles it (#117 —
                    see `sdlc change-code --help` / helptext/estimate.md).
                    Full flow only; the quick flow has none
    ## Plan         checkable steps; `## Plan` milestones gate reviews
    ## Log          dated session notes (### YYYY-MM-DD)
    ## Side quests  (optional) unplanned work that landed

STATUS

{{STATUS_GLOSS}}
Flip status with `sdlc issue set-status` (or `sdlc claim` to start work), never
by hand-editing the frontmatter — the verbs carry the transition guards.
(`done` closes via `sdlc close`.) The status set above is derived from the model.

CHECKPOINTING AND PUBLISHING DOCUMENTATION

`sdlc issue sync --issue N` commits only the selected issue file locally on the
current branch, with no network operation. Use it after design decisions and
before context checkpoints. Separate plans are not collected automatically.

The agent chooses a coherent same-repository documentation commit, then runs:

  sdlc issue publish --commit SHA

The complete commit must contain ordinary Markdown changes under configured
issues/plans/history roots or workshop/projects/. It may deliberately group an
issue, separate plan and project records. Mixed code/docs, root or merge commits,
symlinks and submodules refuse; split the commit explicitly. Atlas/code changes
retain the reviewed PR path.

Publication applies that commit's patch to fresh origin/main with Git three-way
merge. Compatible concurrent edits merge; conflicts refuse the entire commit.
It preserves the caller's branch/index/worktree and never publishes unselected
ancestors. The same behavior applies in :0, numbered worktrees and independent
clones. Source-Commit provenance makes confirmed retries no-ops, including after
a later revert. Unknown acknowledgment preserves the source and reports SHAs
for remote inspection rather than assuming failure.

`issue sync --push` publishes only a new issue commit created by that invocation.
When no new commit exists, use explicit `issue publish --commit SHA`; no older
commit is inferred. `change-code` similarly publishes its new narrow issue
checkpoint after gates pass, and does nothing when no new checkpoint is needed.

Claim only reserves fresh open issues. It does not sweep local body edits into
remote main. Ordinary reviewed PR publication or explicit `sdlc push` can still
publish a branch's commits; local sync makes no promise about those later verbs.

For depth:

  sdlc issue <verb> --help
