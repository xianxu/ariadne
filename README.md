# Ariadne

**The collaborative AI harness — a knowledge OS for the AI era.**

*"Life takes 42 shots."*

AI runs the loops. Humans steer. AI learns. `Ariadne` forms a base of all my tinkering, it represents a paradigm of working. `weave` prepares a repository's declared layers and composes its context; use `weave link` to adopt a base and `weave compile` to prepare and refresh it.

Check `atlas/workflow/index.md` for how to use it (TODO).

For an evidence-backed retrospective of development-process friction in a
current or supplied session transcript, invoke `session-retro`; see
[`atlas/workflow/session-retro.md`](atlas/workflow/session-retro.md).

## Concurrent issue work

Reserve an open issue with `sdlc claim --issue N`; remote main decides whether it
is available. Checkpoint its body locally with `sdlc issue sync --issue N`.
To publish an issue and deliberately selected plan/project files together,
commit them and run `sdlc issue publish --commit SHA`. This applies that change
with three-way merging in every checkout, including primary `:0`, and preserves
unrelated local commits. See [issue publication](atlas/workflow/issue-sync.md).

Planning and close reviews release the local repository lock while the reviewer
runs. SDLC checks the prepared inputs again before recording a result; concurrent
edits require a rerun. `WF_REVIEW_TIMEOUT` defaults to `30m` (allowed `1s`–`2h`).

## Standalone weave startup

Until #241 publishes the Homebrew formula, build the CLI from this checkout:

```sh
go build -o bin/weave ./cmd/weave
# Run from the repository adopting the base:
/path/to/ariadne/bin/weave link github.com/xianxu/ariadne
/path/to/ariadne/bin/weave compile
```

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
valid while the local checkout exists; record a source to restore a missing one.
Paths/sources cannot contain whitespace or `#` in this format.

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
Formula publication is tracked separately in #241; until then, place the
source-built candidate above on PATH for bootstrap to reuse.

The generic seeded CI workflow sets up Homebrew on Linux, then compiles before
running generated helpers. Ariadne source CI provisions its root Brewfile and
builds the current candidate CLI; consumer CI uses the published gateway through
bootstrap. Keep consumer packages in the layer's root `Brewfile`. An optional
executable `scripts/ci-setup.sh` runs after compilation and before merge checks;
a missing or non-executable hook is skipped, and a failure stops the job. Checks
use the materialized local `scripts/run-merge-checks.sh`.

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

The **prepare-weave-release** workflow runs this verification and uploads the
candidate. #241 owns publishing the reviewed commit and archives and installing
the generated formula into `xianxu/homebrew-ariadne`. There are no runtime Go or
Python dependencies in the packaged weave binary. See the
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
rules and snapshot limits. For “branch :2 from :1” or an explicit resting-branch
refresh, follow [Branching and Refreshing Slots](atlas/workflow/workspace-branching.md).
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

## Project records

Projects live in `workshop/projects/` and coordinate issue-backed work across
repositories. Create and inspect them through the model-derived CLI surface:

```sh
sdlc project new --slug example --goal 'Why this exists' --done-when 'A falsifiable MVP boundary'
sdlc project list
sdlc project show --slug example
sdlc project validate --slug example
sdlc project set-status --slug example --to defined
sdlc project status --slug example
sdlc project retro --slug example --dry-run
sdlc project retro --slug example
sdlc project close --slug example
# Find every project record referencing an issue, fleet-wide (archive-inclusive):
sdlc project find --issue metis#18
sdlc resolve --kind project 'metis#18'   # same discovery via the resolver
# Bridge effort to calendar: bless a measured throughput baseline, then show it:
sdlc project throughput --bless 2026-06-22..2026-07-19
sdlc project throughput                  # current baseline + trailing-4wk comparison
# Once blessed, the forecast (throughput ÷ contention → finish vs deadline)
# informs — never blocks — at three surfaces:
sdlc project set-status --slug example --to committed   # computes + derives planned_finish
sdlc project set-status --slug example --to committed --planned-finish 2026-09-01  # override
sdlc project show --slug example         # appends the live forecast-vs-deadline line
# and project close records planned-vs-actual finish in the Calendar ledger.
# A paused or executing project may be dropped instead of completed:
sdlc project close --slug example --drop
# Explicit escape for a legacy record where neither gate applies:
sdlc project close --slug legacy-example --no-retro --no-ledger
```

`set-status` enforces the lifecycle and named guards declared in
`construct/vocabulary/project.cue`; project completion remains owned by
`sdlc project close`. When a retro or calibration ledger genuinely does not
apply, acknowledge that explicitly with `--no-retro` or `--no-ledger`
(`--force` waives both); ordinary closes should satisfy both gates.
