# Ariadne

**The collaborative AI harness — a knowledge OS for the AI era.**

*"Life takes 42 shots."*

AI runs the loops. Humans steer. AI learns. `Ariadne` forms a base of all my tinkering, it represents a paradigm of working. `weave` prepares a repository's declared layers and composes its context; use `weave link` to adopt a base and `weave compile` to prepare and refresh it.

Check `atlas/workflow/index.md` for how to use it (TODO).

For an evidence-backed retrospective of development-process friction in a
current or supplied session transcript, invoke `session-retro`; see
[`atlas/workflow/session-retro.md`](atlas/workflow/session-retro.md).

## Concurrent issue work

In an issue tracker repository (#252; `workshop/issue-tracker.json` on main),
each issue's card — id, status, dates, hours, title — lives on the
`issue-tracker` branch and changes only through sdlc's verbs. `sdlc claim --issue
N[,N…]` takes the lock on one issue or a set (#283, #284), and `sdlc unclaim`
releases it. `sdlc start-plan` starts the issue on its own branch, where the
design is committed with git. `sdlc issue new` publishes the card and the
details to origin at creation (#308). Later edits reach main with the issue
branch, or directly with `sdlc issue publish --issue N` by the owner (anyone,
while unclaimed: #274) (#284). A resting branch only fast-forwards to main. See
[issue tracker](atlas/workflow/issue-tracker.md).

Attach evidence with `sdlc issue attach --issue N --file PATH --caption TEXT`,
then commit both changed paths on the issue branch or use `sdlc issue publish`.
The issue body indexes a bounded plain-Git bundle that follows publish,
archive, restore and slot moves. Repository policy is operator-owned;
public or unknown host visibility defaults to text-only, private to allowed.
Source material must also be eligible for sharing. See
[issue assets](atlas/workflow/issue-assets.md) for policy, limits and recovery.

A claim also records its owner on the card (#277): the operator, a fingerprint
of the machine, and the worktree. Only that workspace continues the issue
through start-plan, change-code and close. `sdlc unclaim` releases it, handing
started work off with its branch pushed, and a plain `sdlc claim` elsewhere
takes it over at exactly that tip (#284); `sdlc move` carries ownership with
the branch within one machine. Moving responsibility between owners, after the
operator coordinates out of band, is `sdlc reclaim --issue N`. It inspects
first; then `--expect REV --reason '…'` transfers the issue to the workspace
running it, refusing if the card changed since you looked (#278). `sdlc issue
show N --json` answers who owns an issue, where its work is, which reviews
passed and whether it landed. It is read-only, from any checkout, with each
answer marked fresh, stale or unknown (#279). What each verb does when it is
repeated, interrupted or answered by a lost response, and which tests prove it,
is in every workflow verb's `--help` and in `sdlc help recovery` (#280).

The issue branch is pushed at every boundary (start-plan, milestone-close,
close, unclaim), leased on the copy last fetched, so started work survives a
lost machine; `sdlc merge` deletes it from the remote once landed (#286).
Started work ends as `wontfix` or `punt` only through `sdlc abandon --issue N
--as wontfix|punt --reason …`: the tip is kept under
`refs/ariadne/abandoned/NNNNNN`, the card records it, the details are archived
on main and the branch is deleted. `sdlc issue set-status working` reopens it,
restoring the branch at that tip with the details moved back out of the
archive.

Older binaries refuse claimed
cards, so after this lands, update ariadne:0 and run `weave refresh` in every
:1+ slot.

A legacy repository cuts over
once with `sdlc issue migrate` — see
[issue tracker migration](atlas/workflow/issue-tracker-migration.md); until then
it keeps [issue publication](atlas/workflow/issue-sync.md): claim on main,
`issue sync` locally, `sdlc issue publish --commit SHA` for selected doc commits
(legacy repositories only).

Planning and close reviews release the local repository lock while the reviewer
runs. SDLC checks the prepared inputs again before recording a result; concurrent
edits require a rerun. A review's timeout scales with its window (30m up to 500 added lines, +15m per further 1,000, at most `2h`; plan reviews get 30m, #300); `WF_REVIEW_TIMEOUT` (allowed `1s`–`2h`) overrides it.

## Classifying issues

An issue can carry a **severity** (`critical | high | medium | low`, shared by
every repository) and **areas**: `/`-paths such as `couch/broadcast/web` naming
where it belongs. Each repository declares its own allowed areas in
`.sdlc/vocabulary.json` on main:

```json
{"version": 1, "enums": {"area": ["couch/broadcast/web", "sdlc/merge"]}}
```

Declaring `couch/broadcast/web` also allows `couch` and `couch/broadcast`. The
file sits under the local-gates guard, so changing the taxonomy needs the
operator's `local-gates-verified` label; an absent file declares no areas. Set
them with `sdlc issue new --area couch/broadcast --severity high`, or later
with `sdlc issue set-area --issue N <area>...` and
`sdlc issue set-severity --issue N <level>` (each takes `--clear`). Both are
card fields. Query with `sdlc issue list --area couch/broadcast`, which matches
whole segments (`couch/broadcast/web`, not `couch/broadcaster`), and
`--severity high,critical`. An undeclared area is refused by these writers, by
`issue publish`, by `sdlc issue validate` and by the merge/push instance gate.
See [repo-defined enums](atlas/workflow/vocabulary.md#repo-defined-enums-348).

## Standalone weave startup

Install the published CLI from the Homebrew tap, then link and compile from the
repository adopting the base:

```sh
brew install xianxu/ariadne/weave
weave link github.com/xianxu/ariadne
weave compile
```

Naming the formula in full lets Homebrew load it from the untrusted third-party
tap. To have a bare `brew upgrade` pick up new releases, trust the tap once with
`brew trust xianxu/ariadne`. To work on weave itself, build the candidate from
this checkout with `go build -o bin/weave ./cmd/weave` and put `bin/` first on
PATH.

`weave link ../ariadne` also accepts an existing local base. Address links clone
into a sibling directory and record the source; local links record the checkout's
origin when available. Matching existing checkouts, including dirty ones, are
reused without pulling or resetting them.

Dependencies are declared in `construct/deps`:

```text
substrate ../ariadne https://github.com/xianxu/ariadne.git
data https://github.com/example/content.git data/content
```

A substrate row takes a path and optional source; a data row takes a source and
mount. Blank lines and `#` comments are supported. Other row kinds and malformed
column counts fail with a line number. Existing two-column substrate rows remain
valid while the local checkout exists. Paths/sources cannot contain whitespace or
`#` in this format.

A two-column row converges to the explicit form on its own. In a primary
checkout, `weave compile` and `weave dependencies` write the source of each such
row into `construct/deps` from its checkout's remote origin, leaving an ordinary
edit to commit; a row whose checkout has no remote origin gets a warning
instead. In a Couch slot (`worktree/<repo>-slotN/`), where the sibling checkout
is absent, weave recovers the source from the same-named checkout next to the
primary, prints a notice naming the row to add, and never edits
`construct/deps`. Without either, record a source to restore a missing checkout.

In a linked Git worktree outside a Couch slot (for example a feature worktree at
`worktree/<repo>/<branch>`), a relative row whose target is absent beside the
worktree resolves beside the repository's primary checkout instead, and
`weave compile` prints `resolved via primary checkout` for it. A target present
beside the worktree still wins. A dependency absent in both places is refused,
not cloned beside the worktree: check it out beside the primary (a
`weave compile` there restores it) and retry. See
[setup-and-replication](atlas/workflow/setup-and-replication.md#linked-worktrees).

With Homebrew on PATH (macOS or Linux), `weave dependencies` restores transitive sources
and runs each layer's committed root `Brewfile` through `brew bundle install
--no-upgrade --file=Brewfile`, foundation-first. Homebrew manages package state;
weave does not build tools, mount data or generate artifacts in this command.
Linux CI uses Homebrew and the same layer Brewfiles.
`weave dependencies --dry-run` makes no changes. If a missing source prevents
reading its declarations, the preview reports that it is incomplete and exits
nonzero rather than claiming a complete dependency list.

`weave compile` runs dependency preparation, then each layer's owner-local
`make tools` foundation-first, then data mounts and the leaf's generators and
artifacts. Owner `tools` targets build explicit binaries from tracked sources
into their own `bin/` and must work before composition. Ariadne builds `sdlc`,
`datatype`, `vocabulary`, and `doc-review`; it does not rebuild the distributed
weave gateway as part of `tools`.

Child processes receive the layer tool directories on PATH. After compilation,
weave prints the directories to add to your own shell PATH; startup never edits
shell configuration. Managed ignore entries and an output identity inventory
allow later compiles to retire unchanged owned outputs while preserving edited
or unrecognized files.

Dynamic-skill generators declare `# weave-output: argv1` and receive an isolated
absolute output directory as their first argument. Their working directory stays
the leaf for graph reads. Weave validates the staged output and publishes it
through ownership checks; generator failure leaves published outputs intact.
Markers and children must stay in the assigned process group and preserve
inherited descriptors rather than daemonizing. Stage cleanup must wait for all
producers to stop; a dead parent alone is insufficient. Marker authors must
migrate legacy scripts before they can run; see the
[generator contract](atlas/workflow/weave.md#dynamic-skill-output-contract).

## Explicit source refresh

Run `weave refresh` from the repository root when you want to adopt newer source
revisions. It discovers the host and its transitive substrate checkouts, requires
all of them clean with no active Git operation, fetches each `origin/main`, and
checks every checkout can fast-forward before advancing any working branch.
Each update uses its captured target SHA and rechecks the starting branch/HEAD.
A numbered host can stay on `main-slot1`; dependency clones stay on their current
branches. On a primary checkout this explicitly includes its declared shared peers.

```sh
weave refresh          # all-repository preflight, fast-forward only, then compile
weave refresh --rebase # explicitly permit replaying local commits, then compile
```

`--rebase` may rewrite local commits or stop with conflicts. Resolve or abort the
Git operation explicitly before retrying. Completed repository updates remain
after any later failure; there is no automatic stash, reset or rollback. Compile
runs even when refs are already current, so retry also repairs a previous build
or composition failure. A later remote advance waits for the next refresh.

Refresh updates substrate repositories, not data-mount repositories. Missing
checkouts require ordinary setup first. A target that changes parsed dependency
declarations refuses before branch updates: reconcile that declaration change
separately so every update stays inside the preflighted graph. Ordinary compile
preserves existing Git revisions, and claiming an issue never invokes refresh.
See [refresh behavior and implementation](atlas/workflow/weave.md#explicit-refresh-247).

## Standalone consumers and maintainer setup

A consumer owns its root `Makefile`, including its product targets and an
optional `-include Makefile.workflow`. On first compile, ariadne seeds a
generic root only when `Makefile` is absent. For an existing regular
Makefile, weave prepends the workflow include without replacing the product
rules; non-regular Makefiles are preserved with an adoption instruction.
`make weave` and the shared `make bootstrap` prerequisite delegate to
`weave compile`; consumer bootstrap extensions remain additive.

The seeded `./bootstrap.sh` runs from its own repository root, reuses a
compatible weave on PATH, or runs `brew install xianxu/ariadne/weave` before
executing `weave compile`. It requires Homebrew rather than installing it.

The generic seeded CI workflow sets up Homebrew on Linux, then compiles before
running generated helpers. Ariadne source CI provisions its root Brewfile and
builds the current candidate CLI; consumer CI uses the published gateway through
bootstrap. Keep consumer packages in the layer's root `Brewfile`. An optional
executable `scripts/ci-setup.sh` runs after compilation and before merge checks;
a missing or non-executable hook is skipped, and a failure stops the job. Checks
use the materialized local `scripts/run-merge-checks.sh`. It runs on PRs to
`main` and on pushes to `main`, so main carries a CI status of its own.

A second seeded workflow, `local-gates-guard.yml`, fails a PR that touches the
local gates' configuration (merge checks, workflows, hooks, sandbox and review
config) unless the PR carries the `local-gates-verified` label. Only the
operator applies that label, after verifying the change; agents never do. Create
the label once per repository (`gh label create local-gates-verified`). `sdlc
merge` refuses to land a PR whose checks are red or still pending. See
`atlas/workflow/ci-merge-check.md`.

## Preparing a weave release

Maintainers with Go, Python 3.9+, and Homebrew can build and test a candidate:

```sh
bash scripts/test/release-weave.test.sh /tmp/weave-release weave-v0.1.0
```

Use a new output directory. The test prepares standalone archives for macOS and
Linux on arm64 and amd64, `SHA256SUMS`, and a generated `weave.rb` formula; it
checks the native binary and runs the formula's local composition fixture.
`scripts/release-weave.sh` prepares the same artifacts without running the test
suite. The version comes from the supplied tag name; these commands build the
current checkout and do not create a Git tag or publish anything.

Each archive holds the `weave` binary and `LICENSE` (MIT). There are no runtime
Go or Python dependencies in the packaged binary.

To publish a reviewed commit (first done for `weave-v0.1.0` in #241):

1. Tag the reviewed commit `weave-vX.Y.Z` and push the tag. Never rebase or
   squash the branch afterwards; merges keep the tagged SHA reachable from main.
2. Build the release on a clean runner at the tag:
   `gh workflow run weave-release.yml --ref weave-vX.Y.Z -f tag=weave-vX.Y.Z`.
   That workflow runs the full release test and uploads the
   `weave-release-candidate` artifact.
3. Download the artifact, check it with `shasum -a 256 -c SHA256SUMS`, and create
   the GitHub release from those exact files (four archives plus
   `SHA256SUMS`): `gh release create weave-vX.Y.Z --verify-tag …`.
4. Commit the artifact's `weave.rb` as `Formula/weave.rb` in
   [`xianxu/homebrew-ariadne`](https://github.com/xianxu/homebrew-ariadne).
5. Verify with an empty trust store (`HOMEBREW_USER_CONFIG_HOME` set to a new
   temp dir): `brew install xianxu/ariadne/weave`, `weave --version`,
   `brew test xianxu/ariadne/weave`, and
   `brew audit --strict --online --formula xianxu/ariadne/weave`. See the
[consumer migration notes](atlas/workflow/setup-and-replication.md#consumer-cutover).

## Workspace identity

Resolve the current checkout or a numbered workspace from Git topology:

```sh
sdlc workspace --json
sdlc workspace :0 --json
sdlc workspace pair:1 --json
sdlc state --json
```

`workspace` is read-only. Its JSON v2 contract separates repository identity,
primary checkout, and current worktree, and reports the address, active branch,
and resting branch. Ordinary feature worktrees have no numbered address or
resting branch. `state` includes the same identity for its current checkout.
Numbered checkouts live at `<fleet>/worktree/<repo>-slotN/<repo>`. Run
`weave compile` there to initialize missing sibling dependencies as independent
clones of recorded remote `origin/main`; later runs preserve selected revisions
and local work. Dependencies have no separate Couch slot. Publish dependency
changes through their own normal SDLC flow before dependent changes.

See the [workspace contract](atlas/workflow/workspace-identity.md) for validation
rules and snapshot limits. To move the current issue branch into another slot
for testing, run `sdlc move` (`:0`) or `sdlc move :N`. For “branch :2 from :1”
or an explicit resting-branch refresh, follow
[Branching, Moving and Refreshing Slots](atlas/workflow/workspace-branching.md).
These ordinary Git procedures preserve the chosen baseline and sibling dependencies;
prepare the issue branch before SDLC planning checkpoints.

After review, `sdlc pr` and `sdlc merge --yes` land :0 and :N through the resting
branch's configured remote/main. They archive the PR's completed records remotely,
return to unchanged `main` or `main-slotN`, and remove only the completed local
issue branch and its configuration. The workspace and dependency clones survive;
refresh remains explicit. Resume interrupted landing with
`sdlc merge --branch <issue-branch> --yes`, including from rest after integration.
See [landing and recovery](atlas/workflow/workspace-branching.md#land-and-retain-the-workspace)
for evidence checks, legacy dependency behavior and the intentionally old baseline.

Merge refuses red CI on the PR head. Durable workspaces wait up to 20 minutes
for pending CI, integrate a moved main locally, and recheck main before landing
(up to three integrations); conflicts abort with the paths to resolve. Legacy
worktrees require manual main integration and refuse pending CI.
A repository can declare local
checks in `.sdlc/local-checks.json` on main: `sdlc local-check run --name X`
runs one on a pushed head under the machine-wide suite lock and records it as
the commit status `sdlc-local/X`. Merge accepts that recorded pass in place of
a CI job that never got a runner (never one that ran and failed), requires it
for checks declared `required`, and runs a missing one itself. See
[recorded local checks](atlas/workflow/ci-merge-check.md#recorded-local-checks-345).

Run `sdlc local-check run --name suite` on the pushed head before close. Merge
requires this suite's recorded pass: a declaration named `suite` overrides its
command, timeout and `weight`, but cannot make it optional. Without a declared
suite, main's `Makefile` supplies `make test` and `WF_TEST_WEIGHT` when it has a
test target. The optional declaration `weight` (default 1) reserves that many
machine-wide suite slots. A passing record on an ancestor remains valid through
docs-only changes such as close's bookkeeping; code changes require another run.

## Fleet queries

Inspect every Git worktree in the sibling-repository fleet from a caller path,
or resolve the concurrency policy for a prospective actor path:

```sh
sdlc fleet inventory --path .
sdlc fleet inventory --path /path/to/nested/checkout --json
sdlc fleet policy --path /path/to/prospective/actor
sdlc fleet policy --path /path/to/prospective/actor --json
```

The policy query prints a typed diagnostic to stdout and exits nonzero when the
repo declaration is missing or invalid, or the requested path is outside the
policy's declared roots. It never infers policy from a repository name. Both
commands are read-only; omit `--json` for their human rendering.

## Manual boundary review

For an ad-hoc fresh-context boundary review, use a pinned committed range or
omit `--head` to include committed-after-base plus staged and unstaged tracked
changes:

```sh
sdlc judge milestone-review --base <ref> --head <ref> --issue <n>
sdlc judge milestone-review --base <ref> --issue <n> --plans-dir <path>
```

`--plans-dir` selects the optional canonical durable plan named to the reviewer
when `--issue` is present; it defaults to `$WF_PLANS_DIR` or
`workshop/plans`. The review prompt carries pinned, read-only Git inspection
recipes rather than embedding the patch.

## Disputing a gate finding

A plan-quality (`PQ-n`) or boundary-review (`BR-n`) finding you can show is wrong,
or not worth its fix, is disputed rather than patched or left to the round cap:

```sh
sdlc finding dispute --issue <n> --id BR-4 \
  --rationale "<why it is wrong>" --evidence "<the command you re-ran and what it showed>"
```

The dispute is committed in the gate ledger; the next review round answers it
against the evidence, and a Critical keeps blocking until it does. See
`sdlc finding dispute --help` and `atlas/workflow/gate-state.md`.

## Project records

Projects live in `workshop/projects/` and coordinate issue-backed work across
repositories. Create and inspect them through the model-derived CLI surface:

```sh
sdlc project new --slug example --goal 'Why this exists' --done-when 'A falsifiable MVP boundary'
sdlc project list
sdlc project show --slug example
sdlc project validate --slug example
sdlc project status --slug example
sdlc project retro --slug example --dry-run
sdlc project retro --slug example
sdlc project close --slug example
# Find every project record referencing an issue, fleet-wide (archive-inclusive):
sdlc project find --issue metis#18
sdlc resolve --kind project 'metis#18'   # same discovery via the resolver
# Forecast remaining focused hours from the trains (critical path as a range):
sdlc project forecast --slug example
# Lifecycle: defined → executing (needs a train) → closing (every mvp_scope issue landed) → close
sdlc project set-status --slug example --to executing
sdlc project set-status --slug example --to closing
# A live (executing or closing) project may be dropped instead of closed:
sdlc project close --slug example --drop
# Explicit escape for a record with no retro:
sdlc project close --slug legacy-example --no-retro
```

`set-status` enforces the lifecycle and named guards declared in
`construct/vocabulary/project.cue`; project completion remains owned by
`sdlc project close`. When a retro genuinely does not apply, acknowledge that
explicitly with `--no-retro` (`--force` waives it too).

## License

MIT — see [LICENSE](LICENSE).
