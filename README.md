# Ariadne

**The collaborative AI harness — a knowledge OS for the AI era.**

*"Life takes 42 shots."*

AI runs the loops. Humans steer. AI learns. `Ariadne` forms a base of all my tinkering, it represents a paradigm of working. To adapt it to a new repo (cloned as a sibling of `ariadne`), run `./bootstrap.sh` (which hands off to `make bootstrap`) — that clones the ancestor layers, builds the tooling, and invokes `weave` (the layer-composition compiler that replaced `construct/setup.sh` in #95) to compose the repo's context. Thereafter `make weave` recomposes on demand.

Check `atlas/workflow/index.md` for how to use it (TODO).

For an evidence-backed retrospective of development-process friction in a
current or supplied session transcript, invoke `session-retro`; see
[`atlas/workflow/session-retro.md`](atlas/workflow/session-retro.md).

## Standalone weave startup (in progress)

The first part of #239 adds source-aware links and dependency installation.
Until #241 publishes the Homebrew formula, build the CLI from this checkout:

```sh
go build -o bin/weave ./cmd/weave
# Run from the repository adopting the base:
/path/to/ariadne/bin/weave link github.com/xianxu/ariadne
/path/to/ariadne/bin/weave dependencies
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

The next slice integrates owner-local `make tools`, compile and bootstrap. The
legacy startup described below still applies until that integration lands.

## Standalone consumers and maintainer setup

A consumer's root `Makefile` is an upstream-owned **seed**: a real file that
weave refreshes from ariadne. Put product targets and local help in
`Makefile.local`; these work without maintainer peers. Avoid editing the seeded
root, since the next weave replaces its contents. Old root symlinks are safely
replaced without changing their ancestor's bytes or permissions.

Run `./bootstrap.sh` in the consumer to clone its peer chain and restore the
maintainer workflow. Bootstrap finds the sibling overlay even when local helper
links are missing, then orders peer setup, weave, tool builds, and installation.
Afterward `make weave` refreshes the substrate on demand.

The generic CI workflow is also upstream-seeded. Put consumer-specific tool
provisioning in an executable `scripts/ci-setup.sh` (`chmod +x scripts/ci-setup.sh`)
and commit it with the product. CI clones peers and sets up the declared Go
version first, then runs this hook before merge checks. A missing or
non-executable hook is skipped; a failing hook stops the job. The runner falls
back to bootstrapped `../ariadne/scripts/run-merge-checks.sh` when the local helper
link is absent. Repeated weave preserves the repo-owned hook.

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
