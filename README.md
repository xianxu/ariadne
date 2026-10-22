# Ariadne

**The collaborative AI harness — a knowledge OS for the AI era.**

*"Life takes 42 shots."*

AI runs the loops. Humans steer. AI learns. `Ariadne` forms a base of all my tinkering, it represents a paradigm of working. `weave` prepares a repository's declared layers and composes its context; use `weave link` to adopt a base and `weave compile` to prepare and refresh it.

Check `atlas/workflow/index.md` for how to use it (TODO).

When a Go type represents JSON from an independently evolving producer, put
`// wire: <why it can evolve separately>` in the type's doc comment. The
repository lint then flags direct strict JSON decoding into that type. For
example, `// wire: peer releases may add fields.` belongs directly above a
peer-message type; version-pinned local stores should remain unmarked. See the
[wire-message lint convention](atlas/workflow/sdlc-binary.md#the-go-lint-catalog-407)
for the covered decoder chain and reasoned exceptions.

For an evidence-backed retrospective of development-process friction in a
current or supplied session transcript, invoke `session-retro`; see
[`atlas/workflow/session-retro.md`](atlas/workflow/session-retro.md).

In universal review, the document's `presentation: propose` asks the edit
engine to show prose changes as proposals. Tracked Couch documents use
`couch --session-sync apply`; a hosted Pair workbench uses `pair review apply`.
An untracked document with that presentation must be tracked before editing.
Both apply paths accept one atomic batch of explained edits and derive the
round; see [universal review](atlas/workflow/universal-review.md).
New `🤖` markers carry at most one `[]` or `{}` turn. A clarification or
longer exchange moves to an indexed `→ [discussion](…)` side doc; old chains
remain readable and resolve by their final human turn. See the
[review convention](construct/local/fix/review-convention.md).

## Concurrent issue work

For a live slot check, `sdlc stage --build` switches the target to this
commit, runs `make build` plus targets declared as `# stage-build: <target>` in
Go commands' `.skip-make-build` files, and refuses missing, modified or
wrong-revision `bin/*` binaries. The target must be a simple Make target name,
not an option or variable assignment. `sdlc stage --clear` returns the target
to its resting branch, fast-forwards it to freshly fetched origin/main, then
rebuilds and checks its binaries at that exact pin. Dirty, ahead/diverged or
failed-fetch states refuse before building. Clear dry-run fetches and reports
behind/ahead counts without checkout or build changes; failed fetch says unverified.
If that post-switch build fails, fix it and repeat `sdlc stage --clear`; the
retry works from the resting branch. One stage at a time (#421): the stage is
a lease naming its holder, so another issue's `sdlc stage`, `--clear` or
`sdlc move` into that slot refuses and names the holder. When the holder is
done it clears its stage. After landing deletes its issue branch, the recorded
original slot can clear from its own resting branch with confirmed current
completion, merged PR and archive evidence. Unknown proof preserves the stage;
finish/reconcile landing or repair the evidence and retry from that slot on rest.
An abandoned one is freed with
`sdlc stage --clear :0 --abandoned <holder>`. See [stage and clear](atlas/workflow/workspace-branching.md#stage-this-branch-on-0-for-a-smoke-test).

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

Issue publication uses explicit remote/refspecs without creating issue-branch
upstream config; re-run `sdlc pr` after later commits. Main/resting tracking stays
unchanged. If protected Git config prevents cleanup of legacy upstream metadata,
landing still succeeds with a warning. Later merge/reconcile sweeps retry that
metadata automatically when config is writable. See
[workspace branching](atlas/workflow/workspace-branching.md#protected-config-during-landing-374).

For a non-issue workshop note edited while an issue branch is active, run
`sdlc trunk publish --path workshop/parley/note.md`. It publishes one unstaged,
uncommitted Markdown file under `workshop/parley/`, `pensive/`, `targets/` or
`projects/`, `workingdoc/`, or `workshop/lessons.md`, directly to main. If main changed that
path since the branch base, reconcile the edit first; the command refuses to
overwrite it. If a push response is uncertain, rerun the same command with the
local file intact. For first publication, `--expect-absent` requires the remote path to be absent or already identical on every CAS attempt; settlement supplies this guard automatically when the source has no prior pointer. Publication leaves the local file untouched. Once this
branch has merged main, run the printed `git restore -- ':(literal)<path>'`
command to discard a tracked local copy, or run it before a merge that would
otherwise conflict. For an untracked note, run the printed targeted
`git clean -f -- ':(literal)<path>'` before merging main; `git restore` cannot
remove an untracked file. If the local file changed during publication,
preserve and reconcile the newer edit instead. Issue details use
`sdlc issue publish --issue N`; plans, atlas,
code and settings travel with the reviewed issue branch. See
[trunk publication](atlas/workflow/sdlc-binary.md#issue-reservation-local-durability-and-publication-244).

Attach evidence with `sdlc issue attach --issue N --file PATH --caption TEXT`,
from the issue's canonical branch once work has started, or from the resting
branch while it is still open. Then commit both changed paths on the issue
branch or use `sdlc issue publish` for an open issue.
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
--as wontfix|punt --reason …` from that issue's canonical branch. The tip is kept under
`refs/ariadne/abandoned/NNNNNN`, the card records it, the details are archived
on main, and the remote is leased to the archived tip before deletion. Local
deletion requires that same tip and no other worktree holding the branch. A
terminal abandon interrupted after recording the card can
be retried from its recorded branch or the resting branch; another issue's
branch is refused. `sdlc issue set-status working` reopens it,
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

Migration preserves verified legacy close evidence and reports each adoption's
measurement class. For an older cutover missing a close binding, inspect
`sdlc issue migrate --repair-completions --issue N`, then apply its reviewed
digest. See [legacy completion recovery](atlas/workflow/issue-tracker-migration.md#legacy-completion-recovery).

Planning and close reviews release the local repository lock while the reviewer
runs. SDLC checks the prepared inputs again before recording a result; concurrent
edits require a rerun. A review's timeout scales with its window (30m up to 500 added lines, +15m per further 1,000, at most `2h`; plan reviews get 30m, #300); `WF_REVIEW_TIMEOUT` (allowed `1s`–`2h`) overrides it.

Boundary reviewers and their preparation commands receive `SDLC_DEP_SUBSTRATE`
and `SDLC_DEP_DATA` paths for the reviewed commit's direct `construct/deps`
dependencies. Each path names a disposable clone pinned once to that local
dependency's last-fetched `origin/main`; setup never fetches or copies dirty
source files. The prompt names each path, exact SHA, fetch-age evidence (or
`unknown`) and isolation mode. Missing sources/refs or invalid/conflicting
declarations stop setup: repair or refresh the local dependency outside the
review, then rerun the original close command.

An available OS fence makes dependency clones and source Git directories
read-only. Without the fence, the prompt identifies writable disposable clones
whose edits are discarded. Retries rebuild the same paths at the same pins;
normal cleanup and dead-owner sweeps remove them with the review checkout.
See [boundary review dependency inspection](atlas/workflow/sdlc-binary.md)
for size/time bounds and freshness limitations.

Boundary reviews can prepare generated assets inside their disposable checkout.
Commit a regular `.sdlc/review-prepare.json` file at the reviewed HEAD:

```json
{"version":1,"command":["make","runtimebundle-generate"],"timeout_minutes":10}
```

The command is an argv array, executed without an implicit shell at the checkout
root under the same fence and confinement as the reviewer. Preparation runs after
confinement on every attempt; retries reset the checkout and prepare again. The
timeout defaults to 10 minutes (explicit range 1–30), capped by the remaining
review deadline. Missing configuration means no preparation; invalid configuration,
command failure, timeout, or tracked-file changes halt the review. Ignored and
untracked generated assets are allowed and removed with the checkout. The config
is read from the pinned HEAD blob, never the mutable working tree.

Preparation output is capped at 64 KiB per attempt. Preparation and rejected-response
diagnostics have a combined 256 KiB cap per dispatch and remain permanently in the
review sidecar, archived with the issue. They cannot supply verdicts or findings.
On Darwin the write fence permits only the literal system `/bin/ps` executable to
run outside Seatbelt; invoking shells still cannot write protected source paths.

A boundary REWORK must cite an answerable blocking finding or explicitly maintain
a known blocker. A reviewer unable to verify can emit `kind: unverifiable`, naming
the failed check and evidence needed. Answer it with `sdlc finding dispute --issue N
--id BR-n --rationale '…' --evidence 'command and measured result'`; the next review
must explicitly dispose it with a note answering the evidence. Two invalid responses
halt with a protocol error and admit no findings. See the
[boundary review contract](atlas/workflow/sdlc-binary.md) for details.

## Local usage

`sdlc usage --since 6h` reads local Claude, Codex and Qoder transcripts without
contacting an agent. The first rows show each agent's last-hour usage and recent
wall-clock burn; details group measurements by slot and model. Claude and Codex
report tokens; Qoder reports exact credits, with tokens unavailable (`null`),
never inferred from credits. Mixed scopes label known subtotals separately.
`sdlc usage --issue N` separates ownership-proven worker usage, explicitly bound
reviewer rounds, and their total alongside measured active time. `--json` uses
schema version 2 with nullable measurements, coverage and unknown attribution.

Optional account calibration uses two observations of the same agent and unit:
`sdlc usage calibrate --agent codex --remaining 29276 --unit credits` (Claude
also accepts `--unit percent`; Qoder accepts native `--unit credits`). Only
calibration writes local meter state. One reading or incomplete spend evidence
yields a status and no projection; local measurements do not claim to be the
provider's quota or invoice. See `sdlc usage --help` and the
[usage map](atlas/workflow/sdlc-binary.md#local-usage-365-385) for counter
semantics, bounds, legacy reviewer gaps and calibration storage.

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
this checkout with `go build -trimpath -o bin/weave ./cmd/weave` and put `bin/` first on
PATH.

`weave link ../ariadne` also accepts an existing local base. Address links clone
into a sibling directory and record the source; local links record the checkout's
origin when available. Matching existing checkouts, including dirty ones, are
reused without pulling or resetting them.

### New repository default

For a new non-brain work repository, finish the setup procedure with tracker
initialization. After linking/compiling the base (or running `./bootstrap.sh`),
commit the initial source and setup files, configure the publication remote,
and push `main`. Then run:

```sh
sdlc issue init
```

This is the final new-repo step for operators and TLs. It publishes the empty
`issue-tracker` branch and main's `workshop/issue-tracker.json`, then safely
fast-forwards the clean main checkout. The first `sdlc issue new` publishes
without writing the checkout; `sdlc start-plan` creates its own issue branch.
Keep other legacy SDLC writers quiescent until initialization finishes.

Brains identified by `.brain/config.md` and absent/declined issue scaffolds are
skipped. Opt-out uses the ordinary adoption/compile path without invoking init.
Existing repositories with issues use the reviewed `sdlc issue migrate` flow;
`issue new` and `start-plan` give init or migrate guidance from the observed local
inventory. Unreadable inventory is never called empty. Normal link, compile,
refresh, CI and repeated bootstrap do not publish anything implicitly.

If setup lacks a commit, pushed main or publication upstream, or the checkout is
dirty/non-main, init refuses before writing. After fixing that prerequisite,
rerun `sdlc issue init`. A retry verifies an existing root/marker and finishes
partial publication without deleting the tracker, resetting work or replacing
source edits.

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
and prepares each layer's committed root `Brewfile`, foundation-first. Numbered
slots (including private dependency checkouts), Couch contexts and SDLC-invoked
weave only run `brew bundle check --no-upgrade --verbose --file=Brewfile`.
Missing packages stop preparation and print their names plus an absolute,
shell-quoted install command for the operator. Hand runs outside those contexts
retain `brew bundle install --no-upgrade --file=Brewfile`. Every brew invocation
sets `HOMEBREW_NO_AUTO_UPDATE=1`, including hand installs. Homebrew manages package state;
weave does not build tools, mount data or generate artifacts in this command.
Linux CI uses Homebrew and the same layer Brewfiles.
`weave dependencies --dry-run` makes no changes. If a missing source prevents
reading its declarations, the preview reports that it is incomplete and exits
nonzero rather than claiming a complete dependency list.

`weave compile` runs dependency preparation, then each layer's owner-local
`make tools` foundation-first, then data mounts and the leaf's generators and
artifacts. Owner `tools` targets build explicit binaries from tracked sources
into their own `bin/` and must work before composition. Ariadne builds `sdlc`,
`datatype`, `vocabulary`, `doc-review`, `tl-worker`, and `tl-lead` (the TL skill
binaries: run `tl-worker --help` or `tl-lead --help`); it does not rebuild the
distributed weave gateway as part of `tools`.

After a same-slot force switch or in a fresh hub session, follow TAKEOVER in
`tl-lead --help`. The [TL tooling map](atlas/workflow/tl-skills.md) points to
the ordered recovery procedure and its rendered-help guards.

Worker live verification uses Pair's shared disposable-Couch harness,
`cmd/internal/testcouch`; run `make test-couch-fleet` from the Pair repository.
Reuse it for isolated live tests instead of the live fleet or private fixtures.
See [TL tooling](atlas/workflow/tl-skills.md) for the worker verification rules.

For a hub with `.hub`, `tl-lead status slots FILE` shows its owned numbered
slots; `--all` shows the machine view. `tl-lead fleet resume` is an
operator-started dry run; `--all` exposes excluded slots without expanding
automatic resume authority. `--apply` waits for each terminal receipt and
stops at the first failure or uncertainty, naming untouched slots. See
[TL tooling](atlas/workflow/tl-skills.md) for scope, holds and recovery.

Before filing a new project's implementation issues, follow SHAPING in
`tl-lead --help`: users, flows, clickable mock, visual design, components and
slices, with operator review at each step. UI mocks use
`couch mock serve <project>/mocks`, numbered versions and shared fixtures.
The screen-by-screen mock-to-real switch is development-only; production
refuses mock-backed screens. The [project datatype](construct/datatype/project.md)
describes optional Users, Flows, States, Mock and Components sections.

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
weave refresh --substrate # preserve the consumer issue branch; refresh substrates and compile
```

Use `--substrate` from a clean issue branch to adopt newer declared substrates
without fetching or advancing the consumer repository. Its branch, HEAD, index
and origin/main remain unchanged; compilation can update generated composition.
It cannot be combined with `--rebase`. Numbered slots refresh only their private
substrates, with the same ownership, cleanliness and preflight safeguards.

Planning gates use `weave refresh --defer-settings`: they never write the
sandbox-protected `.claude/settings.json`. When the checkout is already current,
`weave settings` checks it read-only without installing dependencies, building
tools or generating outputs. Stale settings produce one notice: run
`weave refresh --substrate` from your shell. This publishes settings while
preserving your issue branch; the next settings check clears the notice.
The normal rebuilt-binary rerun after Git catch-up still applies.

SDLC checks the resolved suite-lock helper before acquisition. An incompatible
substrate names its checkout and HEAD and directs you to `weave refresh --substrate`,
then retry the original local-check command. If the helper belongs to the consumer
source itself, merge current main into the issue branch instead: substrate-only
refresh cannot update tracked source. With no substrates this mode only recompiles.

`--rebase` may rewrite local commits or stop with conflicts. Resolve or abort the
Git operation explicitly before retrying. Completed repository updates remain
after any later failure; there is no automatic stash, reset or rollback. Compile
runs even when refs are already current, so retry also repairs a previous build
or composition failure. A later remote advance waits for the next refresh.

Refresh updates substrate repositories, not data-mount repositories. Missing
checkouts require ordinary setup first. A target that changes parsed dependency
declarations refuses before branch updates and names the changed declarations
and fields without source URLs. In the named checkout, if its clean branch is
behind origin/main, run `git merge --ff-only origin/main`, then `./bootstrap.sh`,
then retry the original refresh command from its original checkout. An ahead or
divergent branch needs reconciliation first; run ordinary setup after adopting
the declarations. See [declaration-change recovery](atlas/workflow/setup-and-replication.md#recover-a-refresh-blocked-by-declaration-changes).
Ordinary compile preserves existing Git revisions, and claiming an issue never
invokes refresh. See [refresh behavior and implementation](atlas/workflow/weave.md#explicit-refresh-247).

## Standalone consumers and maintainer setup

A consumer owns its root `Makefile`, including its product targets and an
optional `-include Makefile.workflow`. On first compile, ariadne seeds a
generic root only when `Makefile` is absent. For an existing regular
Makefile, weave prepends the workflow include without replacing the product
rules; non-regular Makefiles are preserved with an adoption instruction.
`make weave` and the shared `make bootstrap` prerequisite delegate to
`weave compile`; consumer bootstrap extensions remain additive.

Ariadne's owner tool builds and `make test` default to `-trimpath`, so identical
packages share Go's build cache across worktrees. Existing effective `GOFLAGS`
(including compile caps) are retained; an explicit Make command-line `GOFLAGS`
assignment overrides the default. The recorded suite declares the same
`make test` command. Generated dev aliases also pass `-trimpath` when building
each peer owner's tools; reload existing shell functions from the checkout root
with `source <(construct/dev-aliases.sh)`. For direct Go commands, use
`go build -trimpath` or `go test -trimpath`. See the
[build and test details](atlas/workflow/sdlc-binary.md#build--install).
When diagnosing duplicate dev-alias owners, run
`construct/dev-aliases.sh --list --verbose`: it names same-origin checkouts
whose warning was suppressed. `--strict` fails only when a binary name has
owners from distinct Git origins; GitHub HTTPS and SSH forms identify the
same repository, while unknown origins remain distinct.

Numbered slots resolve dev-alias owners to verified fleet `:0` checkouts,
keeping private substrate clones separate from shell tools. A missing or
mismatched primary refuses generation; Couch calls run the existing central
binary without rebuilding it. Use `sdlc --version` to inspect its actual path,
embedded revision and modified bit; Couch warns on a different executable path.
Checkout refresh alone cannot replace a calling shell's existing functions:
evaluate `source <(construct/dev-aliases.sh)` in that shell, or relaunch an agent
whose tools reload a captured old shell snapshot. A PATH symlink directly to
`:0/bin/sdlc` picks up the TL's rebuild without alias regeneration. See
[dev-tool ownership](atlas/workflow/base-layer.md#dev-binaries--ownership--location-dev-aliasessh)
and [executable identity](atlas/workflow/sdlc-binary.md#executable-identity-461).

The construct shell tests run as named Go subtests, with a separate two-minute
hang bound per script in each color environment. For the same selection without
execution, run `bash scripts/test/construct-scripts.test.sh --list`; it emits
NUL-delimited paths and applies the runner's exclusions. Omit `--list` to execute
the selected batch, as the CI shell check does.

For host checks required by `make test`, provide `tests/host-preflight.sh`, or
set `WF_TEST_PREFLIGHT ?= path/to/preflight.sh` before including
`Makefile.workflow`. The workflow runs the script before acquiring the shared
full-suite lock: failure prints the script's diagnostics and stops without
queueing; success proceeds to the lock. Repositories without a preflight keep
their existing behavior. See the [test gate documentation](atlas/workflow/sdlc-binary.md#running-the-test-suite-253).

The seeded `./bootstrap.sh` runs from its own repository root, reuses a
compatible weave on PATH, or runs `brew install xianxu/ariadne/weave` before
executing `weave compile`. It requires Homebrew rather than installing it.
When creating a new work repository, continue through the
[new repository default](#new-repository-default) and finish with `sdlc issue init`.

The generic seeded CI workflow sets up Homebrew on Linux, then compiles before
running generated helpers. Consumer CI builds its gateway from the exact full
Ariadne commit in the authored `.sdlc/weave-compiler-ref`, using that source's
Go toolchain, and puts the private binary first on PATH before bootstrap. The
pin is not weave-managed: new consumers must explicitly create it from an
approved supported ref. A missing or malformed pin, fetch/build failure or
mismatched binary provenance stops CI; there is no floating compiler fallback.
Compiler metadata checks consume the full input, so a successful match cannot
turn into a producer SIGPIPE under `pipefail`.
Ariadne source CI builds the candidate separately and checks its vocabulary
exports against every supported compiler in `.sdlc/weave-compiler-refs`.
See [compiler compatibility](atlas/workflow/ci-merge-check.md#compiler-compatibility-459)
for cold-regeneration controls and pin rollout. Keep consumer packages in the
layer's root `Brewfile`. An optional
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

## Network incidents

Git resolver and connection failures report `infrastructure: network` with
the observed host and an exact diagnostic command. Run:

```sh
sdlc net check --host github.com
sdlc net check --remote origin
```

With no flags, `sdlc net check` uses the configured publication remote. It
compares native system DNS with direct DNS through the configured server,
then reports HTTPS through each path using one sampled address, verified TLS,
and no proxy or credentials. The read-only check has an eight-second budget;
missing tools and partial observations remain visible. See `sdlc net --help`
and the [SDLC map](atlas/workflow/sdlc-binary.md#network-incidents).

Workers report external dependency outages immediately to the hub. The TL
notifies the operator immediately and on verified recovery; the authoritative
rule is `tl-lead --help`. Preserve any uncertain write's existing recovery
state rather than treating a network fault as test evidence or blindly retrying.

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

Side documents for `plan.md` belong in `plan.side/`. Close and landing preflight
refuse `plan.md.side/` before publication: rename it to `plan.side/`, retain its
contents and update links, then retry. Existing parent-only archives require an
explicit owner-coordinated Git repair; a manual repair does not rewrite the old
archive's proof. See [SDLC archive behavior](atlas/workflow/sdlc-binary.md).

After review, `sdlc pr` and `sdlc merge --yes` land :0 and :N through the resting
branch's configured remote/main. They archive the PR's completed records remotely,
return to unchanged `main` or `main-slotN`, and remove only the completed local
issue branch and its configuration. The workspace and dependency clones survive;
refresh remains explicit. Resume interrupted landing with
`sdlc merge --branch <issue-branch> --yes`, including from rest after integration.
See [landing and recovery](atlas/workflow/workspace-branching.md#land-and-retain-the-workspace)
for evidence checks, legacy dependency behavior and the intentionally old baseline.

Merges in one repository land one at a time (#393). `sdlc merge` takes a turn in
the repository's landing line and prints its position while it waits, so each
admitted landing integrates current main and verifies the resulting code before
another slot can land. A separate pre-close suite verifies the issue branch; it
does not reserve a landing turn. The hub does not hold main by hand or direct
workers to chase newer main before close. `sdlc landing status` shows the landing now, the merges
waiting in order, and the ready-to-landed delay. For a fix that blocks the
fleet, the operator or the TL (from a :0 slot) runs
`sdlc landing grant --repo DIR --issue N --reason WHY [--hours H]` (at most 24
hours; `sdlc landing revoke` ends it early). That puts the issue first at the
next admission without interrupting a running landing or waiving any gate.
While any valid grant exists, normal admissions wait even before the granted
issue queues. Granted issues retain FIFO among themselves. Status and waiter
output name each grant's reason, expiry and revoke command. Expiry or revocation
releases the hold; a successful landing ends its grant. A train's queued integration
suite inherits any active grant on a car in its exact tested prefix; later grants,
revocation and expiry update its queue class without restarting it.

Landing trains (#449) batch that line. `.sdlc/landing.json` on main sets
`train_interval_minutes`, `train_max_cars` (default 8) and, since #475,
`train_max_wait_minutes` (default 0: no bound). The interval is 0 when the file
is absent, which keeps the serial line. The file is under the local-gates
guard, so the operator verifies every change to it.

With trains on:
- **Vouching:** a waiting merge checks its own review and CI gates on its pushed
  head and vouches it ready.
- **Departure:** a merge departs only when the train is due and its vouch is
  current. The train is due:
  - on the first train;
  - at once when `train_max_cars` cars are ready (#475);
  - once the oldest ready car has waited `train_max_wait_minutes` since its first ready vouch;
  - at once after a departure that left ready cars behind;
  - otherwise at the next departure: the repository's next wall-clock slot after the last train (with a machine plan), or one interval after it.

  An early departure never moves later slots. A lone or fleet-blocking landing is not a departure: it never pushes the next train back.

The machine plan (#475) spaces repositories' train departures across the
machine's suite slots. It is one file per machine beside the suite lock:
`${XDG_STATE_HOME:-~/.local/state}/ariadne/landing-plan.json`, or
`$WF_TEST_LOCK_FILE`'s directory. It belongs to the steward.
- `sdlc landing plan` prints each train-enabled repository's offset and its
  cost basis (its suite's 7-day p90 from test-stats, times its declared suite
  weight).
- `sdlc landing plan --write` recomputes the plan and writes it. It keeps
  unknown members and other repositories' entries, and refuses to replace a
  plan it can't read.
- Offsets are packed within `$WF_TEST_SUITES` slots (default 2); costs that
  can't fit are named.
- A repository without an entry, or any repository while the plan is
  unreadable, departs one interval after its last train.
- `sdlc landing status` names the grid in use.
- **Conducting:** the admitted head conducts. It boards every waiting merge
  whose remote tip and open PR are still at the vouched head. It builds one
  branch from main and each car's head, with the instance and duplicate-ID gates
  per car. It runs one suite and CI on its train PR, merges that PR once, and
  GitHub marks each car's PR merged. Each car's own `sdlc merge` then finishes
  its cleanup.
- **Membership and holds:** the first banner names candidates; the composed
  membership is printed before suites. A held car prints its reason and the
  branch-specific recovery command, then exits. Fix the problem, integrate main,
  and re-close before retrying. If the conductor’s own car is held during
  composition, it stops before publishing and releases survivors to reform under
  the next conductor. Its checkout cannot cancel that other checkout’s suite.
  An unexplained suite failure is bisected: the culprit is held, the clean prefix
  lands, and the cars behind it ride the next train at once.
- **Refusals:** a known flake gets its #394 rerun and never splits. Main failing
  the same way refuses the train, and so does red CI on the train PR.
- **Riding alone:** a fleet-blocking grant departs at once and lands alone,
  never boarding a train. So does a PR with the local-gates label.
- **Recovery:** a car that finds its own PR merged, with no word from its train,
  finishes as a resume.

A composed car can be held before merge starts with
`sdlc landing drop --issue N [--reason WHY] [--repo DIR]`, run by its owner or
the repository's :0 steward. The obsolete train stops and its survivors re-form;
the dropped runner exits with its reason and exact branch retry command. It stays
out until its owner explicitly retries. Once the Merge effect starts after green
CI, drop refuses with the current state. `sdlc landing status` names accepted
drops and any pending old-train cleanup; resume the printed conductor command
if that cleanup needs recovery. Older conductors without control support refuse
the command rather than promise a drop.

`sdlc landing status` shows the setting, the next departure, each waiting
merge's readiness, and the last trains: result, cars, train PR, split suites
and culprit.

Merge refuses red CI on the PR head. Durable workspaces wait up to 20 minutes
for pending CI, integrate a moved main locally, and recheck main before landing
(up to three code integrations); conflicts abort with the paths to resolve.
Issue-only main movement does not spend that budget. Fresh GitHub CLEAN evidence
for the exact PR/head and live base allows the tested head to land unchanged;
landing trains use the same proof to preserve their checked head and CI, with
selected landing-check evidence still valid and destination gates rechecked.
Otherwise normal integration and exact-head CI apply. Card Markdown qualifies;
tracker configuration does not. The narrow record policy
is described in [durable PR landing](atlas/workflow/sdlc-binary.md#durable-pr-landing). Legacy
worktrees require manual main integration and refuse pending CI.
A repository can declare local
checks in `.sdlc/local-checks.json` on main: `sdlc local-check run --name X`
runs one on a pushed head (under the machine-wide suite lock by default) and records it as
the commit status `sdlc-local/X`. Merge accepts that recorded pass in place of
a CI job that never got a runner (never one that ran and failed), requires it
for checks declared `required`, and runs a missing one itself, unless a saved
pass of it on that exact head can be posted instead. See
[recorded local checks](atlas/workflow/ci-merge-check.md#recorded-local-checks-345).
Commands receive `SDLC_CALLER_ROOT` and caller-resolved `SDLC_DEP_SUBSTRATE` /
`SDLC_DEP_DATA` from the tested tree's `construct/deps`, while running in their
detached checkout. Dependency paths are read-only inputs; see the environment
contract in `sdlc local-check --help`, including duplicate-row refusals.
Version2 adds `scope: "landing"`, `paths_file: "tests/conformance-paths.json"`
and `resource: "runner"` for checks whose runner owns independent capacity.
A landing entry must explicitly declare `required` and `resource`; runner checks
cannot declare suite `weight` or CI `substitutes`. The path file is
`{"version":1,"paths":["glob",...]}` with ordered includes/exclusions. Policies
are read from the pinned main tree. Existing version1 declarations keep their
per-head defaults. `local-check run --all` excludes landing checks; name one
explicitly for manual execution. Runner checks use no suite slot and retain
30s of graceful process-group cancellation even after the immediate leader exits.
Unconfirmed cleanup retains checkout, scratch and raw logs under the same Git
worktree lock and leaves queue activity unknown for operator reconciliation.
Runner execution refuses if its attempt writer cannot be created. Retention has
no automatic expiry while ownership is held or unreadable. After proving that
all owned descendants stopped, reconcile the queue and remove the named scratch
and log directories, then unlock/remove the retained worktree; preserve Git
metadata until linked artifacts are reconciled. SDLC never unlocks VM positions.
Enabled (`required:true`) landing checks run after the successful final-prefix
suite, once per landing owner, before train or single-car merge. Selection unions
each included car's paths, so a later revert cannot hide an earlier trigger.
Failures block without VM retries, suite bisection or main-baseline diagnosis.
A bare saved/manual status does not replace the owner-bound gate. Only proven
unwatched docs equivalence may reuse that owner's original exact-head evidence;
a watched `.md` change still invalidates it. `--no-ci` still owes the suite and
selected landing checks. Disabled entries run only by explicit name on pinned
main for burn-in. This capability activates no consumer or hosted workflow.

`sdlc local-check cancel --run ID` stops exactly one owned queued or running
check, using the lease ID printed when it starts. `--name N` cancels every owned
run of that check; omitting both cancels every owned check. These selectors never
cross checkouts. A same-check run on a different head prints an exact cancel
command for the existing run; both keep their FIFO places until explicitly
cancelled. Never stop one by a PID picked from `ps`. When a finished
check's status post can't reach GitHub, `sdlc local-check republish --name X`
posts the saved result later, with no rerun.

Go startup failures caused by unavailable temp space, a full cache/build output,
quota exhaustion or a missing toolchain count as could-not-run when complete
evidence proves no tests started. For an already recorded failure, its owner or
repository `:0` can run `sdlc local-check reclassify --attempt ID --reason TEXT`.
This preserves the original record in an audit and records a nonpassing v3
disposition. It changes no GitHub status and requires a genuine rerun after
repair; mixed, test-started, truncated or unprovable evidence refuses.
`sh scripts/verify-524-reader.sh` checks actual correction bytes against the pinned
historical reader in a private checkout. It prints its retained scratch path;
the invoker owns removing that directory after inspection.

`sdlc queue pause --owner TOKEN [--reason TEXT] [--cancel-running]` holds new
machine-suite and landing admission. Operator/:0 owners compose; `sdlc queue
resume --owner TOKEN` removes only that hold and never restarts interrupted work.
`sdlc queue status --json` reports the frozen v1 owner/activity schema. Incomplete
or unknown observations never prove quiescence. Isolated conformance supplies
private `WF_SDLC_QUEUE_ROOT`, `WF_TEST_LOCK_FILE` and
`WF_SDLC_QUEUE_FLEET_ROOT`; it must never pause the live fleet. See the
[queue backend and rollout](atlas/workflow/sdlc-binary.md#machine-wide-queue-drain-481).

For a retained unknown operation, the operator/:0 may run
`sdlc queue reconcile --id ID --digest SHA256 --note TEXT` after verifying that
no descendants remain. The digest names the exact raw `unresolved/ID.json`
bytes. A live/unknown original owner or lease, changed bytes or unreadable
evidence refuses. The command appends original bytes (base64), digest, actor,
note and time to permanent `reconcile.jsonl` and fsyncs it before removing only
that tombstone. If interrupted after the audit append, rerun the same ID/digest
to finish without another audit entry. It never resumes holds or restarts work.

Managed local checks retain their latest raw interruption log per check/head as
`<git common dir>/sdlc/local-checks/<head>/<check>.interrupted.log`, outside
accepted suite attempts. The normal newest-20-head log retention applies.
`tl-worker --help` describes notice-only graceful/fast drain checkpoints,
deferred work messages, continuation/restart timing around owned recorded checks,
and verification of retained checks after a restart. `tl-lead --help` owns the
steward's queue controls, restart-safety judgment and explicit operator recovery
authority, plus capability-qualified hub automation. Flags and quiet screens do
not prove restart safety. See [drain and hub automation](atlas/workflow/tl-skills.md#drain-and-hub-automation).

A repository can also declare an **affected suite** (#473) in `.sdlc/affected.json`; without the file the mode is off.
For generated Go graph inputs, that file can include
`"graph_prepare": {"command": ["make", "generate"], "timeout_minutes": 2}`.
The command is an argv array, read from pinned main and run once in the requested
head's private graph checkout before `go list`. The default timeout is two
minutes, with a maximum of 30; omission or null disables preparation. A preparation
failure or timeout selects the full suite and names its cause; cancellation
cannot record a passing affected check. Generated graph files are discarded with
that checkout: the affected test command must establish its own prerequisites.
Older sdlc versions reject the new field, so upgrade the provider before adopting it.
`sdlc local-check run --name affected` then runs only what the branch's change can break, and the result of a
pushed head's run depends on what changed:

- **Code or a declared group changed:** sdlc runs the repo's affected command with `SDLC_AFFECTED_GO_PKGS` and
  `SDLC_AFFECTED_GROUPS` set and records `sdlc-local/affected`. `SDLC_AFFECTED_GO_PKGS` holds the changed Go packages
  plus every package that imports them, test imports included (from `go list`). `SDLC_AFFECTED_GROUPS` holds the
  declared non-Go test groups the change touches. Both are sorted and space-separated.
- **A `full` path or an unclassified file changed:** sdlc runs and records the `suite` instead.
- **Only docs that no package owns or embeds changed:** sdlc records `affected` without running anything.
- **Nothing changed against main:** the run refuses.

Selection follows imports and embeds only. A Go test that scans *other* packages' source, or reads a non-Go file
outside its own `testdata/` and embeds, is invisible to it, so the repo's affected command or its `full` globs must
cover such tests (see pair#725). The schema, the classification rules and these blind spots are in [the affected
suite](atlas/workflow/ci-merge-check.md#the-affected-suite-473).

Recorded suites and their baselines replace unusable inherited `TMPDIR` and
`GOTMPDIR` with writable per-run paths, even when TestScratch falls back.
A proven Go work-directory startup failure reports an environment error naming
those variables, with a separate `.environment.log`; it posts no code-failure
status or suite attempt. Repair access and rerun the same check or merge. A
baseline that could not start leaves main's health unknown. Mixed or ambiguous
output keeps the ordinary failure rules, and earlier real failures still block.

Known flaky tests are registered on their issues (#394). A test is a known flake
while an open issue's card lists it in `flake_tests`. `sdlc flake register --test
PKG.TestX [--issue N]` writes the field only with same-head fail→pass evidence
from this machine's test-stats store (or `--evidence TEXT` for a CI run); without
`--issue` it files an unclaimed flake issue. Closing the issue unregisters its
tests.

A name may be an exact subtest (#466), such as `PKG.TestX/plain/dev-aliases.test.sh`.
Suite attribution ignores indented quotations of nested helper commands (#476).
Raw indented subtest output coarsens to the proven top-level test; JSON events
and explicit shard headers retain exact leaves. A visible note explains why a
subtest-only flake registration may miss a coarsened failure. Quoted package
trailers cannot change its package.

- **What it explains:** only its own failure and its subtests'. A sibling
  subtest's failure stays unknown and blocks.
- **Rerun:** its rerun is the whole top-level test, and it must show the
  subtest's package-bound JSON PASS event.
- **Evidence:** registering a subtest needs the top-level test's same-head
  fail→pass, and the failed attempt's saved log must name that subtest.
- **Correcting a registration:**
  - `sdlc flake narrow --issue N --test OLD --to OLD/sub` replaces a name with one of its own subtests, with no new evidence.
  - `sdlc flake unregister --issue N --test NAME` removes a name.
  - Both are for the card's owner, or anyone while the issue is unclaimed.
- **Rollout:** rebuild :0, and refresh the slots, before registering a subtest
  name. An older sdlc reads one as malformed and turns the known-flake rule off:
  every failure blocks. It also refuses a test-stats attempt whose
  `known_flakes` names a subtest, so its success guard and test-stats fail
  closed until it is rebuilt.

`sdlc flake list` is the TL's local daily view: the registry with each
issue's owner and age, same-head flakes not yet registered, and failures nothing
explains, each with its log. A same-head failure is omitted from the default
unregistered list when exactly one closed issue's card `flake_tests` covers the
failing leaf by the gate's rule (the test itself, or its nearest registered
ancestor, #479), and its fix merge landed after the failing head.
Issue Markdown never claims a test. `sdlc flake list -v` shows omitted rows
with the fixing issue and merge. A failure on a head containing that merge stays
visible as a recurrence; ambiguous or unreadable evidence also stays visible,
with a coverage reason when the uncertainty is specific to a row. CI failures
are shown as not covered until #451.
When a recorded suite (`local-check run`, or merge's) fails only on known flakes, it
reruns just those tests once on the same head, grouped into exact-package
commands under one weight-1 lease and total timeout. Registered subtests run
their whole top-level test, including siblings. Each required test and package
must have a matching structured JSON PASS event, and its
flake issue must still be open when the result is posted; then the status reads
"pass · known flake #N rerun". Any other failure blocks: an unregistered test, a
build error or timeout, or a known flake failing again. A full rerun that passes
after an unexplained failure on the same commit is recorded but never posted as
success: register the flake (that fail→pass is its evidence) and republish, or fix
it. The suite's `known_flakes`
key in `.sdlc/local-checks.json` (`rerun`, the default; `accept`, which posts the
pass with the issue link and no rerun; `off`) is the operator's policy.

An already saved suite failure can be resumed after its failures are registered:
`sdlc local-check resume --name suite --attempt ID --rev ORIGINAL_SHA` runs only
the exact package tests on that pushed commit. It preserves the original suite
record and duration. Repeating it validates and reposts completed proof without
executing tests. A pending or interrupted resume blocks later success on the
same commit. After a crash, the operator or :0 steward verifies all descendants
stopped, then runs the exact `local-check reconcile --attempt ID --link-digest
SHA256 --note TEXT` printed by the refusal. That records terminal interruption;
fix/advance the code and run a suite on a new head. Cleanup has one pass in the
original live owner. Any cleanup failure or crash reports needs-reconcile and
preserves surviving files and Git locks. Reconciliation only records the audit;
it never cleans, unlocks or retries cleanup, including on repeated calls.
Interrupted bundles and terminal/orphan receipts remain indefinitely, independent
of lock visibility; their retained checkout, scratch and logs consume disk space.
Completed-child cleanup/retention keeps its existing policy. Descendants are not
discovered automatically, and reconciliation never grants a pass or execution.

Run `sdlc local-check run --name suite` on the pushed issue head before close.
After it passes, close that tested branch without manually integrating each
newer main and repeating pre-close suites. Branch code fixes still require fresh
verification, and close's noon freshness gate still applies. On SHIP, enter
`sdlc merge --yes`: its landing turn integrates current main and runs the
authoritative suite when the integrated code needs it. Merge requires this suite's recorded pass: a declaration named `suite` overrides its
command, timeout and `weight`, but cannot make it optional. Without a declared
suite, main's `Makefile` supplies `make test` and `WF_TEST_WEIGHT` when it has a
test target. Only when main has neither a declaration nor a test target, the
exact pushed head supplies a **bootstrap** default `make test` and weight, with
ordinary managed `sdlc-local/suite` evidence. Its first declaration's custom
checks and CI substitutions take effect after landing on main; bootstrap never
stands in for server CI. See `sdlc local-check run --help` and the
[CI gate map](atlas/workflow/ci-merge-check.md).
The optional declaration `weight` (default 1) reserves that many
machine-wide suite slots. A passing record on an ancestor remains valid through
docs-only changes such as close's bookkeeping or main prose publication. Code,
tests, configuration and embedded help invalidate reuse and require another run.
When the integrated head's suite fails, merge runs an unrecorded copy on main
and labels the named failures as introduced or pre-existing in its refusal.

Reader-first compatibility verification is opt-in:
`sh scripts/verify-pre524-reader.sh` runs the committed control against pinned
pre-#524 source in a private clone. It proves the old global refusal; the new
reader scopes “needs a newer sdlc” to related proof. The invoker owns the printed
scratch directory: inspect its evidence, then remove that directory. No live
binary or production store is changed. Before #463 may activate resume links,
the TL must upgrade every slot's actual sdlc and verify paths/revisions; any
stale or unverified reader keeps activation held. See the
[reader-first rollout contract](atlas/workflow/sdlc-binary.md#reader-first-attempt-compatibility-524-prerequisite-for-463).

The private phase1 executable conformance is opt-in:
`SDLC_PRIVATE_CONFORMANCE=1 go test ./cmd/sdlc -run '^TestLocalCheckPhase1Conformance$' -count=1 -v`.
It runs the production CLI against private Git, tracker, queue, lock and fake
GitHub state, proving scoped future evidence and exact-package decoy exclusion.
Its temporary inputs and outputs are removed automatically after inspection
inside the test; it changes no installed binary or live store.

`sdlc test-stats` reads saved suite attempts across local sibling repositories
and reports wall and queue percentiles, slow packages or shards, and named
flakes. Use `--since 7d`, `--repo pair`, or `--json` to narrow or export the
view. Historical per-head logs are marked partial because overwritten retries
and queue waits cannot be reconstructed. The command does not run tests or
contact GitHub.

For a permanent weekly baseline, run `sdlc test-stats snapshot` from an SDLC
checkout. It uses the latest completed Monday 00:00 UTC boundary by default;
`sdlc test-stats snapshot --end 2026-10-18T00:00:00Z` selects another completed
seven-day UTC window. `--brain-root /path/to/brain` overrides the sibling brain
capture repo. The command writes one immutable, versioned JSON file at
`brain/memory/work/devinfra/test-economy/weekly/<end-date>.json`. An identical
rerun returns a validated existing schema-v2 artifact without recollection;
incompatible or corrupt bytes refuse instead of being overwritten. Each
repository headline puts full suites per landed issue beside covered/observed
landings, unknowns and top coverage reasons. Missing, expired or contradictory
evidence makes the ratio `null`, and the snapshot keeps physical suite statistics
separate from exact landing attribution. Publish only after reviewing the pilot
output because the dated artifact is not replaced in place.
Schema v2 adds a resource section with Pair-inventory family sizes, the largest
Couch ledgers, suite wall trends and machine gauges. Its `observed_at` records
when point resources were measured; their numeric weekly deltas require a 6–8
day gap between observations. Suite durations compare their exact seven-day
windows. The TL runs this snapshot manually after Monday 00:00 UTC; there is
no scheduled producer yet.

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

Persisted-format changes follow `ARCH-EVOLVE`: evolve additively, capture unknown
fields and preserve them unchanged through every rewrite, including nested
members. Known-field types, required fields and duplicate keys remain strict.
A version bump that older readers refuse needs a concrete reason why ignoring
an addition would be unsafe. Inspect the complete registry with
`sdlc arch-principles`; planning and review prompts receive the same definitions.

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
sdlc project close --slug example --dry-run      # every refusal, nothing written
sdlc project close --slug example --verified 'The MVP boundary holds: <evidence>'
# Tick a landed issue in every live project citing it, on each project's origin
# main (merge runs this itself; see landing_tick below):
sdlc project tick --issue ariadne#373 --pr 251 [--dry-run]
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
explicitly with `--no-retro` (`--force` waives it too). Close also requires
`--verified` evidence and re-checks that every `mvp_scope` issue landed
(`--no-mvp`, or `--force`, waives that); `status` prints the same `mvp: L/T landed` line
and the project's `done_when`.

When an issue lands, `sdlc merge` ticks its rows in every live project citing it
and appends `(PR #N)`, one commit on the project repository's origin main. A
project without `tl:` is ticked by default; a TL-owned one only with
`landing_tick: true`, and otherwise merge prints the `sdlc project tick` command
for the TL.

Issue actuals are measured from ownership intervals and the owning slots' agent
sessions. Close retains its measurement evidence so later `sdlc actual` replays
the same inputs; `sdlc actual --milestone Mx` replays a finalized milestone.
Each new close, including `--fixed-then-ship`, measures its current candidate
head and tracker authority rather than adopting a prior completion's snapshot.
Retained train waits appear as counted **unverified coverage** beside hours;
the diagnostic describes a bound waiter invocation and makes no idle subtraction.
`sdlc issue correct-actual --issue N --reason TEXT --dry-run` previews a measured
repair of a completed sample. If legacy rows are ambiguous, its diagnostic names
the ledger and eligible raw-row hashes. Have the TL/operator select the row by
its provenance and use `--target-row <full SHA256>`; ordering is not evidence.
`--exclude` invalidates a proven bad sample when numeric evidence cannot be
trusted. Pending corrections retain their target; a conflicting selector refuses.
See `sdlc actual --help` for provenance and recovery.

## License

MIT — see [LICENSE](LICENSE).


`tl-lead doc settle SIDE-DOC [--to PATH]` publishes a resolved side document to
`workshop/workingdoc/YYYY-MM-DD-topic.md` or an eligible project prose path,
then records its path and commit in the live source. Re-settling updates that
copy; Git preserves earlier editions. Resolve review questions and proposals
before settling, and use a quiet editing window. `tl-lead doc status` shows
unsettled pages and content drift from the recorded edition. The
[workingdoc prototype](construct/datatype/workingdoc.md) describes the page;
[TL skill binaries](atlas/workflow/tl-skills.md) maps publication and recovery.

`tl-lead status` uses the hub repo's gitignored `session.md` and
`session.side/`; FILE is optional on all status commands. Existing explicit
FILE arguments remain valid. Copy decisions from these local conversations
into the owning tracked issue, project Log or charter when they are made.

To migrate an older hub file, close every affected editor and run
`tl-lead status move [OLD]` to inspect the dry run, then repeat with `--apply`.
Omitting OLD requires exactly one `teams/*/tl-status-*.md`. The source must
already be registered with Couch; the refusal prints the exact registration
command. Apply copies and rebases documents, updates `.gitignore`, preserves
Couch identity/unread/owed state, verifies, then cleans up unchanged sources.
Only plain `parent: ../file.md` values are accepted; other YAML parent forms
refuse before effects and name the file. Missing ignore rules are appended,
then real Git ignore checks must pass for every destination before copying.
A partial run keeps copies for the same command to reconcile. After the root
registration switch the new tree is live: edits and locks there are allowed.
Cleanup removes unchanged old sources without comparing or overwriting live
copies. Keep old leftovers quiet between cleanup retries. Registry observations
and Git checks are batched; Couch calls are bounded to five seconds, with clear
unreachable or unconfirmed-outcome errors and no silent retry.
Reopen `session.md` after success. See `tl-lead --help` for refusals and the
quiet-window requirement.

Session migration refuses `.gitignore` files anywhere in the old or new side tree before effects. It rechecks effective Git protection before switching registrations and before each source deletion.
