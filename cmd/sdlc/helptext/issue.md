`sdlc issue` is the CRUD/authoring surface for the issue *record*. It
complements the flat checkpoint verbs (`close`, `claim`, `change-code`, `pr`,
`merge`, `milestone-close`) — those guard workflow *transitions*; `issue *`
edits the record itself.

SUBCOMMANDS

  new            Reserve the next ID's card on the tracker and write the details
                 locally (`--from-github N` seeds it from a GitHub issue)
  publish        Publish details to main (#284): a first publication (no owner
                 needed; makes the issue claimable) or the owner's later edits,
                 one narrow main commit for a set (`--issue 284,285`)
  move-detail    The first-publication case of `publish`, for one issue
  recovery       `list` / `reconcile --issue N` interrupted tracker operations
  migrate        One-time cutover of a legacy repository to the issue tracker
                 (dry run; `--apply --expect DIGEST`; `--reconcile` on a branch)
  set-status     Flip an issue card's status with transition guards
  set-title      Retitle an issue card (paths keep their slug)
  set-estimate   Record estimate_hours on the card (`--hours`)
  set-github     Link the card to a GitHub issue (`--number`)
  sync           (legacy repositories) commit an issue body; retired in tracker
                 repositories: commit details with git, publish with `publish`
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

CARDS AND DETAILS (#252)

An issue is two files. Its card (`workshop/issue-cards/NNNNNN-<slug>.md` on the
`issue-tracker` branch) holds the fields everyone needs current — id, status,
started, dates, estimate/actual hours, GitHub link, title — and the original
Problem. Only sdlc writes a card, by compare-and-swap commits of its own. The
details (this file, at the path below) carry Spec, Done when, Plan, Log and
branch-owned fields, plus a read-only mirror of the card fields that sdlc
refreshes; a hand edit to a mirrored field is refused with the setter to use.

Details travel with the work: `issue new` writes them in the current checkout
(a narrow commit on a feature branch; uncommitted on the resting branch), and
they land on main through the branch's PR or `issue publish` (first
publication). Only then is the issue claimable. Checkpoint design with ordinary
git commits on the issue branch.

PUBLISHING DETAILS (#284). `sdlc issue publish --issue N[,N…]` is how details
reach main without shipping a branch. Per issue:
  - first publication (details not yet on main): `move-detail`'s transfer —
    the creator's checkout publishes, no owner needed; the filing branch's
    copy is removed (narrow commit) or the rest fast-forwards, and the card
    records the handoff.
  - republish (details already on main): only the owner. The edits (details
    bodies) are judged against the copy at this checkout's merge base with
    main: main unchanged since → published; main moved → refused, bring main
    in first; already equal → nothing to publish. Every republished issue goes
    in ONE main commit (`#a,#b: issue: publish details`), ownership re-checked
    against the tracker just before the push. Afterwards a resting branch
    fast-forwards (it never carries the edits as commits), and any other
    branch commits the same bytes narrowly, so its later merge changes nothing
    there. A lost push response: rerun — main already holds the details, and
    the rerun only finishes the checkout.
A resting branch only fast-forwards to main; never commit on it. An interrupted card/main publication
keeps a receipt: `sdlc issue recovery list` shows it, and `reconcile` resumes
it from the checkout that owns it, probing before repeating anything.

A tracker repository is marked by `workshop/issue-tracker.json` on main, naming
the tracker's root commit. A checkout whose marker disagrees with the tracker
(a tracker but no marker, a marker but no tracker, another root) refuses every
card read and write with the next action; unmirrored details in such a
repository refuse the legacy close/change-code paths. `sync` is retired to a
pointer (#284); `publish --commit` and the Makefile/Python shell fallbacks refuse.

MIGRATION (#252)

`sdlc issue migrate` plans the one-time cutover from the pinned main, every
local and publication-remote branch, and this clone's worktrees, and prints the
cards (with each inferred value), duplicate IDs, refusals and a digest. It
changes nothing. Refusals name what blocks the cutover — card fields changed
on a branch but never published, issues that exist only on a branch, branches
editing archived issues, uncommitted issue edits, two active files sharing an
ID, or a codecomplete issue without exactly one provable legacy close — each
with its next action under the old workflow.

With every SDLC writer frozen, `--apply --expect DIGEST` re-plans, refuses any
other digest, bootstraps the issue-tracker branch (or adopts it only when its
root holds exactly the planned cards), then publishes one main commit that
mirrors every active details file and adds the marker. Rerun the same command
after an interruption: each phase is recognized, nothing is rolled back.
Archived details are never rewritten. Afterwards `git pull` each resting
checkout. A branch from before the cutover refuses sdlc commands until it is
brought across — now or on its next use — by `--reconcile` (or by merging
origin/main): it proves every details file the branch changed kept the
imported card's fields and has a card, then merges main's migration commit
(not the rest of main), so later merges with main start from converted
details. A conflict aborts the merge and says how to resolve it.

LEGACY CHECKPOINTING AND PUBLISHING (repositories before the #252 cutover)

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
commit is inferred. (Since #252, `change-code` publishes nothing: its design
checkpoint is a local commit on the issue branch.)

Claim only reserves fresh open issues. It does not sweep local body edits into
remote main. Ordinary reviewed PR publication or explicit `sdlc push` can still
publish a branch's commits; local sync makes no promise about those later verbs.

For depth:

  sdlc issue <verb> --help
