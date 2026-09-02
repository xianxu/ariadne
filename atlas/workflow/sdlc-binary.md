# sdlc binary

`sdlc` is the SDLC checkpoint binary — one Go binary at `cmd/sdlc/`
that collects ariadne's workflow checkpoint guards into a unified verb
namespace with embedded `--help` per subcommand. Agents should invoke
`sdlc` directly; Makefile targets are compatibility wrappers for
downstream repos that have not built the binary yet.

Design rationale: `docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md`.
Build issue + plan: `workshop/history/issues/000031-sdlc-checkpoint-binary.md`
(after archive) or `workshop/issues/000031-...` (during build).

## What it owns

The **checkpoints between SDLC stages** — not the stages themselves.
Stages stay prose; the binary refuses transitions that lack required
evidence. Subcommands are added incrementally when the same drift
recurs at a stage (not by formalizing the SDLC as a state machine).

## Verb surface

| Verb              | Replaces (Make target)      | Defends |
|-------------------|-----------------------------|---------|
| `close`           | `make close-issue`          | Issue close: actual + verified + atlas + plan ticked; on full-issue close auto-dispatches the one boundary review (#69, `--no-judge` to skip) |
| `actual`          | (new #68)                   | Compute an issue's focused dev-hours (in-binary active-time-v3 engine over brain+repo transcript sources) |
| `active-time`     | (new #110; was active-time-v3.py) | Standalone CLI over the same engine — the per-segment attribution table for manual inspection; preserves the 2/3/0 loud-fail exit codes |
| `state`           | (new)                       | Workflow state inspection + drift detection |
| `fleet inventory` / `fleet policy` | (new #200) | Read-only fleet worktree evidence and prospective-path admission-policy resolution; typed JSON is the contract and human output renders the same values |
| `resolve`         | (new #144)                  | **Read-only** symbolic-ref → current path(s): the issue + its plan/review family, archive-correct + cross-repo. Locations from the `discovery:` model; grammar single-sourced as the parser. No lock (see below) |
| `open`            | (new #144)                  | Sugar over `resolve`: open the primary artifact in `$EDITOR` |
| `propagate-base`  | (new #106; precheck #109)   | Re-weave every recursive DEPENDENT of this repo (downstream counterpart to `substrateChain`): discover dependents (Makefile.workflow + substrate chain), order foundation-first, then per repo a clean-tree precheck → `make weave` + verify-complete + commit (untracking now-generated files). A dependent with a DIRTY working tree (pre-existing uncommitted work — e.g. a concurrent session) is SKIPPED untouched (never `git add -A`'d) and the run exits non-zero. `--dry-run`/`--ref`. |
| `judge`           | `make check-{dry,pure,plan,specs,lessons}` | Fresh-context LLM judge (anti-collusion) |
| `fetch`           | `make fetch N`              | **Hidden deprecated alias** for `sdlc issue new --from-github` since #56 M2 (keeps `--github-issue`) |
| `claim`           | `make issue-sync`           | Issue-file workstream-claim onto main (formerly `lock`, #39) |
| `start-plan`      | (new #75)                   | Planning-entry transition: delivers the `at-plan` architecture lens + the durable-plan pointer (`superpowers-writing-plans` → `workshop/plans/`, #72) to design against |
| `change-code`     | `make worktree` (partial)   | Planning → implementation gate, in this order (#187 B1): structural + **plan-quality (stateful, #187)** + estimate (#113) + estimate-reconciliation + estimate-quality (#117) + branching (in-place default, `--worktree=yes`/`=ask`; #39, #51) |
| `set-status`      | (new)                       | Status-transition guards. Moved under `sdlc issue set-status` (#56 M2); **hidden deprecated flat alias** kept one cycle |
| `push`            | `make push`                 | Direct-on-main ship + the #124 instance-conformance gate (`--no-validate`) + pre-flight judges (still available; not the default close path since #51) |
| `pr`              | `make pull-request`         | PR creation with Fixes-issue body |
| `merge`           | `make merge`                | Branch merge (in-place or worktree) via PR + the #124 instance-conformance gate (`--no-validate`) + cleanup + irreversible-action confirm (#51) |
| `milestone-close` | `make close-issue MILESTONE=Mx` | Milestone close + auto-dispatched boundary review (the one reviewer, per-milestone window; #69). THE milestone-close path — `close` refuses `--milestone` (#146); `--no-judge` here is the labeled skip-review escape. |
| `issue new`       | (new; xx-issues skill prose)| Allocate next ID + write canonical template (`--from-github N` seeds from GitHub) |
| `issue sync`      | (new #206)                  | Commit ONE issue's body (Spec/Plan/Log) under `#N: issue-sync: <what>`. **Does not push** — `--push` opts in. The planning-phase counterpart to `claim`, which publishes only the reservation. The mid-planning trigger is delivered by `start-plan`'s output + AGENTS.md §2/§14, not left in `--help` |
| `issue set-status`| ← flat `set-status`         | Status-transition guards (relocated #56 M2) |
| `issue list`      | (new)                       | List issues (ID/status/title), sorted by ID; `--status` filters; reuses `listIssues` |
| `issue show`      | (new)                       | Issue frontmatter + section headers, no bodies |
| `issue validate`  | (new #124)                  | Validate issue file(s) against `#Issue` — frontmatter cue-vet (via `vocabulary validate-instance`) + section presence; multi-target (#133): `<file>...` / `--issue N[,N...]` / `--all` (mutually exclusive). The on-demand surface of the instance-conformance loop |
| `project new/list/show/validate` | (new #180 M3) | Author and inspect project records. Scaffold sections/status and discovery derive from `#Project`; validation shells to the noun-generic vocabulary validator |
| `project set-status` | (new #180 M3) | Enforce the project lifecycle and its ordered named guards from `project.cue`; unknown guards fail closed, evidence lands in Log, and `done` remains owned by `project close` |
| `project status/retro` | (new #180 M4) | Derive progress, dependency frontier, remaining effort, and thread components from live issue records; append dated re-forecast checkpoints without overwriting the baseline |
| `project close` | (new #180 M4) | Require the modeled executing→done (or executing/paused→dropped) edge and a retro, roll Phase-A vs issue actuals into the brain fog ledger unless explicitly bypassed, then archive through `ArchiveSubdir(..., ArchiveProjects)` |

### Issue-file sync: durability vs publication (#206)

One dispatch — `syncIssuesToMain` in `claim.go` — serves `claim`, `issue new`,
`issue sync` and `change-code`.

**`change-code`'s invariant: the issue file ends up committed in THIS worktree,**
on the branch about to carry the work — across `resolveBranchName`'s three name
modes × {on main, in-place feature branch, feature worktree} × {untracked,
tracked-and-edited}. Two consequences that each cost a review round to find:

- The id comes from the RESOLVED issue path, not `--issue` — only one of the
  three name modes sets that flag, and in auto-detect with `--worktree=yes`
  gating on it left the new worktree holding no issue file at all, since `git
  worktree add` does not carry untracked files.
- Publishing is conditioned on **already being on main**, not on the caller's
  intent. From a branch the publish route would copy the in-progress body into
  the main worktree, commit it on main and push — a half-written Spec on
  `origin/main`, the branch's copy left dirty, two network round-trips per
  milestone re-run. `pr`/`merge`/`close` are what publish.

`TestChangeCodeSyncIssue_ModeMatrix` runs the whole table.

**Publishing is not gated on working-tree changes.** Splitting durability from
publication made "committed locally, not yet pushed" a state the code
deliberately creates, and `changedIssueFiles` is empty in exactly that state — so
both sync arms skip only the COMMIT when nothing changed, never the publish.
Otherwise `--push` is a silent no-op precisely when it is needed: finishing a
sync whose commit landed and whose push failed, which is the recovery every
warning in this area names. `issue new` follows
the same durability-before-publication rule: when its reservation broadcast
can't reach main (commonly: run from an in-place feature branch, where no
worktree is on main), it falls back to a local commit rather than leaving the
new issue untracked.

Its two arms are **not** "on main vs on a
branch"; they are *commit here* vs *publish to origin/main from elsewhere*:

| arm | what it does | reached when |
|---|---|---|
| `syncInPlace` | `add` + `commit` in THIS worktree on THIS branch, then `push origin main` unless `NoPush` | the caller isn't publishing, **or** this worktree is already on main |
| `syncViaMainWorktree` | find the worktree on main, refuse if it has uncommitted issue changes, `pull --rebase`, conflict-detect, copy, commit + push there | publishing from a feature branch |

Every step of the second arm exists to publish, so suppressing the push doesn't
just skip its last line — it selects the other arm entirely. That is what makes
a no-push sync cheap: local, offline-safe, no worktree hunt, and usable from an
in-place feature branch where no worktree is on main at all.

The publish choice is spelled `NoPush` on `claimFlags`, never `Push`: `issue
new` builds that struct as a literal, so a positive field would zero-value to
false there and silently kill the reservation broadcast (#82 M1). The commit
subject is a parameter too (`""` = each arm's historical default), which is the
only thing `issue sync` adds over the shared helper.

**The sync subject is declared bookkeeping.** `#206: issue-sync: spec/plan`
anchors `#N`, so without an entry in `gitx.bookkeepingVerbs` it would read as
shipped implementation to `IsShippedWorkSubject` — and from there to drift
detection (`state.go`), milestone review windows (`milestoneclose.go`) and
active-time attribution. The hyphen in the `issue-sync` lead-in is load-bearing:
whole-token matching keeps it from swallowing real work titled `#N: issue sync
verb: …`.

**Commit pathspecs (#206).** Every sdlc commit that follows a *narrowed* `git
add` now carries the same paths as a commit pathspec (`--`, implying `--only`).
A bare `git commit` records the whole index, so a peer agent's staged work was
swept into a commit that misdescribed it — the repo transaction lock serializes
sdlc verbs against each other, but nothing stops a peer running plain `git add`.
Seven sites: both sync arms, both `push.go` archive commits, `merge.go`'s, and
`migrate.go`'s two (source + destination). `archiveCommitArgs` derives its path
list *from* `archiveAddArgs` so the two cannot drift, and refuses an empty move
list — `git commit -m … --` with no paths is read as NO pathspec and commits the
whole index, which is the helper's own failure mode on its degenerate input.

Reachability differs by site. `syncViaMainWorktree`'s `pull --rebase` already
refuses a dirty index, so its pathspec closes a pull→commit race rather than a
live bug. `syncInPlace`, `push`'s archive and `migrate`'s **source** side have no
such guard — migrate's cleanliness check (`status --porcelain -- relPath`) is
scoped to the migrated file alone — and those three are the deterministic
regressions in `issuesync_test.go`. `migrate --no-commit` prints the pathspec'd
form in its hints too, so the operator isn't handed the defective command.

**The class is guarded at the source, not per site.** `TestGitCommitsCarryTheirPathspec`
(`commitpathspec_guard_test.go`) parses every non-test file in `cmd/sdlc` and
accepts a git-commit argv only when it carries a `--` pathspec, or carries `-a`,
or sits in a function that stages `git add -A` *and* is allowlisted with its
reason. The rule it encodes is *a commit must be as narrow as its add* —
`push.go`'s `commit -a` and `propagatebase.go`'s `add -A` pairing are the two
legitimate whole-tree cases, and requiring both halves is what stops an
allowlist entry from widening to cover a sibling commit in the same function
(the first cut keyed on the function alone and excused `push.go`'s archive
commit on the strength of an unrelated `commit -a`).

Its companion `TestVerbsWireTheirCommitHelpers` asserts the CALL SITES. Three
close-review rounds converged on the rule behind both: *a fix at a call site is
pinned only by a test entering through the production entry point* — a real-git
test that hand-builds the argv proves the helper and mocks the wiring. Where the
entry point is not in-process drivable (`runChangeCode`, #191; `runPush` /
`runMerge` / `runMigrate` die through gates a unit test cannot satisfy), the
source guard asserts the edge instead. Weaker than driving the verb, strictly
stronger than a call site that can be deleted with the suite still green.

**New failure mode:** a pathspec'd commit is a *partial* commit, and git refuses
one outright while `MERGE_HEAD` is set (`fatal: cannot do a partial commit during
a merge`). A sync or archive run mid-merge now fails loudly where the bare commit
would have folded the merge in silently. That is the better behavior, but it is a
behavior change: finish or abort the merge first.

**Flat verbs vs the `issue` group (#56).** The flat verbs guard workflow
*transitions* (close, claim, change-code, pr, merge, …). `sdlc issue *` is the
CRUD/authoring surface for the issue *record* — the noun-grouped home for
`new` (and, post-#56-M2, `set-status`/`list`/`show`). The canonical issue-file
template lives in one place: the `Render` function in `internal/issue/scaffold.go`,
documented in prose by `sdlc issue --help`.

## Fleet inventory and policy (`sdlc fleet`, #200)

The read-only `sdlc fleet inventory` command owns assembly of canonical
repository/worktree rows across the shared filtered fleet walk. Each row places
measured Git evidence (HEAD SHA, commit timestamp, base divergence, and dirty
count) beside branch-associated issue metadata whose declared status is
explicitly attributed with `branch-prefix` provenance. It reports observations
only: neither the typed contract nor the human renderer derives `cold`, `drift`,
`liveness`, `staleness`, or a similar judgment.

Concurrency declarations live at `.sdlc/fleet.json`, but that spelling is owned
by `construct/vocabulary/fleet-policy.cue` and reaches Go through
`pkg/vocab.FleetPolicy().DeclarationPath`; inventory and the CLI share
`fleet.PolicyDeclarationPath` rather than restating it. `fleet.LoadPolicyFile`
is the strict declaration boundary, and `fleet.ResolvePolicy` is the pure core
over a validated capability plus canonical paths. Inventory therefore exposes
only the repository's policy **capability** (or its declaration diagnostic); it
does not invent a target-specific admission key. `sdlc fleet policy --path P`
returns the distinct resolved **policy result** for prospective `P`.

Before policy resolution, `fleet.CanonicalProspectivePath` resolves existing
components in filesystem order (including symlinks before later `..`), retains a
safe nonexistent suffix, and returns the canonical request plus its deepest
existing directory. The command passes that directory to `NormalizeVantage`,
which selects the containing worktree and its shared repository/fleet identities.
Dangling links, ambiguous parent traversal after a missing component, and
non-directory traversal fail closed. Missing/invalid declarations and
outside-scope requests remain structured diagnostics: inventory keeps other rows
visible, while the policy command writes the typed diagnostic to stdout and then
returns a nonzero refusal without usage text.

Implementation pointers: `cmd/sdlc/fleet.go` is the thin command adapter;
`cmd/sdlc/internal/fleet/{inventory,load,policy,gitpaths,types,render}.go` owns the
typed core and boundaries. Contract coverage lives beside it in
`{inventory,load,gitpaths,json,render}_test.go`, with stateful fake/portable Git
agreement in `git_conformance_test.go`; command routing and refusal semantics are
pinned in `cmd/sdlc/fleet_test.go`.

## Repo transaction lock (#132)

Mutating `sdlc` verbs are serialized by an SDLC-owned local transaction lock at
`.git/sdlc.lock`. Most mutating verbs hold the lock for the whole command
transaction, not just individual Git calls: issue ID allocation, issue/status
file writes, commits, branch changes, local archive moves, and pushes all run
under the same holder.
The lock directory is created atomically with `mkdir`; holder metadata lives in
`meta.json` inside the directory and records pid, hostname, cwd, command, argv,
and start time.

**Test hermeticity (#149/#165).** Because the lock path (and other repo state)
resolves from cwd via `git rev-parse`, a `cmd/sdlc` test that drives a mutating verb
through `buildRoot().Execute()` without chdir-ing into a temp git repo would grab
the developer's REAL `.git/sdlc.lock` (hanging `go test` under a live holder) and
could mutate the real tree (a stray test sequence corrupted `main` in the #148
session). Two guards: command-tree tests chdir into an isolated repo via
`hermeticRepo(t)` (so the lock resolves to the temp `.git`); and a package `TestMain`
snapshots the real repo (HEAD/branch/porcelain/`.git/sdlc.lock`) before+after the
run and FAILS a passing run that left durable damage (`snapshotDiff`, pure) — the
backstop that catches any test that still leaks.

The lock path is resolved from `git rev-parse --git-common-dir`, so linked
worktrees for one repo share the same lock. That is intentional: worktrees share
the issue namespace, object store, and remote refs that the motivating races
touched. The lock does not serialize another clone or machine, so remote
push/ref races still surface through the existing push/merge retry guidance.

`close` and `milestone-close` are narrower: they lock the compute phase, release
the lock while the external boundary review runs, then reacquire before
finalization and refuse to write if the issue file or any prepared project-file
edit changed while the lock was released, if the canonical durable plan changed
presence or contents, or if the commits that landed during the review carry code
surface (#194) — a doc-only delta finalizes, since the reviewed code is unchanged.
`change-code`, `merge`, and `push` may still hold the lock while
synchronous judges run. Their wait/timeout messages call this out as a
long-running review/ship transaction; quick commands should wait or retry
instead of deleting a live lock. Recovery is conservative but not wedging:
`die()` drains the active lock cleanup registry before `os.Exit`, missing
`meta.json` during the tiny mkdir-before-write window is treated as holder
initialization and polled through, and a confirmed-dead same-host holder is
reclaimed by atomically renaming the stale lock directory before removal.
Cross-host or over-age uncertainty still produces operator-facing recovery
guidance rather than silent deletion; a live same-host pid overrides the age
ceiling.

## Artifact-reference resolution (`sdlc resolve` / `open`, #144)

ariadne artifacts cross-reference each other with **symbolic** refs
(`ariadne#11`, `#15 M4`, `pair#84`). The id is stable but the path is not — slugs
get renamed, and `sdlc merge`/`push` move an issue and its whole plan/review
family `issues|plans/ → history/` (#160). So refs stay symbolic and resolve at
**read time**; nothing is stored, nothing rots. `sdlc resolve <ref>` is that
resolver (the parley#160 editor UX shells to it).

- **Grammar single-sourced as the parser.** `parseRef` (`cmd/sdlc/resolve.go`) is
  the *only* implementation of the ref grammar. parley#160 and agents shell to
  `sdlc resolve` rather than re-encoding it in Lua — so the grammar can't diverge.
  `helptext/resolve.md` documents it for humans; a test (`TestResolveDocExamplesParse`)
  binds every documented example back to `parseRef` so the doc can't drift.
- **Locations derive from the model, not hardcoded.** The issue's `discovery:`
  block (`construct/vocabulary/issue.cue`, read via `pkg/vocab` `Discovery()`) now
  carries `home` + `glob` + `archive` (`workshop/history` — the ROOT; writes
  land in per-kind subdirs `history/issues` + `history/plans` derived by the
  Go-owned `vocab.ArchiveSubdirs`, #181; reads tolerate the flat legacy
  layout, so downstream repos migrate lazily) + `plans`
  (`workshop/plans`). `familyFiles` globs those three dirs for `NNNNNN-*.md` and
  `classifyFamily` sorts issue → plan → reviews. A 6-digit id resolves the whole
  family; `#id Mx` narrows to the `-mX-review.md` sidecar; `gh#id` labels a GitHub
  ref without resolving a local file (read-only + offline).
- **Cross-repo** by scanning the current repo's parent for a sibling: exact
  basename wins, else a unique case-insensitive prefix (`parley` → `parley.nvim`).
  The sibling enumeration is factored into `project.SiblingRepoDirs` (ARCH-DRY),
  shared with the cross-repo **project discovery** below.

### Cross-repo project discovery (`project.DiscoverByIssueRef`, #171)

The close gate no longer looks project files up in a single `--brain-dir`
(`FindByIssueRef`, removed). `project.DiscoverByIssueRef(parentDir, repo, id,
scope)` walks the fleet (`SiblingRepoDirs`) and globs each peer's
`workshop/projects` — plus, under `ActiveAndArchive`, `workshop/history/projects`
(derived via `vocab.ArchiveSubdir`) — for the marker `[repo#id`, returning
**every** match (multiple projects referencing one issue is legitimate
membership, not ambiguity). The deprecated `brain/data/project` legacy home is
still scanned (a loud deprecation warning nudges migration) so the still-active
`metis-v2` keeps ticking until it moves. Scope is load-bearing: the close gate
uses `ActiveOnly` (and drops terminal-status legacy matches so an archived
`done` project is never re-ticked); `find`/`resolve`/parley use
`ActiveAndArchive`. A fleet-glob skip-list excludes non-fleet siblings
(`*.bak`, `worktree`, dot-dirs) that the exact-match resolver was immune to.
`sdlc close` loops the existing tick/upsert helpers over every match
(`closeResult.projectEdits []projectEdit`), each carrying its `repoDir` for the
safe peer-write commit decision below.

### Safe peer-write commit mechanics (`peerwrite.go`, #171 M3)

A close that edits a project file in a *peer* repo must not strand the edit
uncommitted in that peer's working tree — but blindly committing into someone
else's checkout is worse. `planPeerWrites(edits, states, curRepoDir,
closingRef)` is the genuinely-pure decision core: per peer repo it authorizes
a **scoped auto-commit** only when git state makes the commit unambiguous —
on `main`, clean index, clean *target files*, and not a brain capture repo
(`RepoGitState.IsBrain` via `gitx.IsBrainRepo`, #176). Anything else —
off-main, pre-existing staged changes, pre-existing uncommitted edits to the
very files being committed (`TargetFilesDirty`, snapshotted BEFORE the close's
file writes so another session's work is never absorbed), undeterminable
branch, unknown state, brain — is **report-only**: the file stays written and
the operator gets the reason plus the exact `cd … && git add … && git commit`
next action (for a brain: leave it — nous sweeps it). The current repo is always omitted (its edit rides the normal
close commit), and a report-only outcome never fails the close.
`readRepoGitState` + `applyPeerWrites` are the thin shell over the shared
`gitRunner` seam (`runner.go`); the close path owns a package-level
`closeRunner` and `applyClose(stdout, stderr, gitRunner, …)` runs the
plan→apply pair after its file writes. The peer commit is scoped by pathspec
(`git commit -m … -- <files>`) and its message cites the closing ref
(`project: close-time update (<repo>#<id>)`), so peer history says which close
produced it. Pinned end-to-end by real multi-repo git fixtures
(`peerwrite_apply_test.go`), including the cross-repo close case where the
matched project lives in a different repo than the closing one.
- **Read-only ⟹ lock-free by construction.** `resolve`/`open` are never tagged
  `markMutatingCommand`, so `wrapRepoLockCommands` skips them and they never touch
  `.git/sdlc.lock` (proven structurally + under a held lock). That's what makes it
  cheap enough (~process spawn) for parley to shell to on a keypress.

### Fleet project navigation (`projectfind.go`, #171 M4)

`sdlc project find --issue <ref>` and `sdlc resolve --kind project <ref>` both
answer "which project records reference this issue, anywhere in the fleet?"
via one shared seam: `discoverProjectsForRef` (parse ref → `resolveRepoDir`
sibling matching, exact-then-unique-prefix → `DiscoverByIssueRef` under
**`ActiveAndArchive`**). Navigation is archive-inclusive by design — active
`workshop/projects/`, archived `workshop/history/projects/`, and the
deprecated brain legacy home (flagged ` (legacy)` in text mode; JSON rows
carry kind `"project"`). Default `resolve` (kind issue) is pinned unchanged.
Read-only, lock-free. parley.nvim binds `gP` (`ResolveRefProject` →
`sdlc resolve --json --kind project`) as the always-cross-repo project jump,
separate from `gf`'s issue-family flow. Caveat (shared with issue resolution,
`resolve.go` sibling-model note): cross-repo discovery applies THIS repo's
discovery model to siblings — fine while all peers share the ariadne layout;
a peer customizing `discovery:` would need its own model loaded here.

Pure core (`parseRef`, `classifyFamily`) is unit-tested with no IO; the IO seams
(`resolveRepoDir`, `familyFiles`) test against temp repos (ARCH-PURE). **Follow-up
(#163):** the existing `workshop/plans`/`workshop/history` hardcoders in
`push`/`merge`/`state` archive logic should migrate onto the same `Discovery()`
accessor — a DRY consolidation, separate from this resolver.

### Project calendar forecast (`throughput.go` + `forecast.go` + `projectforecast.go`, #182)

Bridges effort (hours) to calendar (a date) — a project's load-bearing
`deadline:` is a date, but both estimators produce hours. The forecast
**informs, never blocks** (estimation is often wrong and slippage is
recoverable by means the math can't see), so it's a computed statement
recorded at the right surfaces, not a gate.

- **Measured throughput (`internal/estimate/throughput.go`, M1).** `SpanThroughput`
  sums the calibration ledger's `actual` hours — **deduped to one row per issue**, since the
  ledger is written per CLOSE and a re-close carries a cumulative partial sum (#192) — over an
  operator-**blessed** representative span ÷ span-weeks; the operator picks the span (trailing
  windows skew under vacations/crunch), the machinery measures the rate.
  `sdlc project throughput --bless FROM..TO` appends a provenance row
  (`{span, hours_per_week, rows, ceiling}` — `rows` counts distinct ISSUES since #192;
  earlier rows counted raw ledger lines and are not comparable) to
  `brain/data/life/42shots/velocity/throughput-baseline.tsv` (append-only,
  last = current); the bare form shows the current baseline + a trailing-4wk
  comparison, never auto-substituted. Reuses the existing `estimate.LedgerRow`
  parser (one ledger parser, ARCH-DRY).
- **Pure forecast core (`internal/project/forecast.go`, M2).** `ComputeForecast`
  divides this project's remaining issue-hours by its share of throughput
  (baseline h/wk ÷ n active projects) → projected finish. **Unit identity:**
  numerator (remaining issue-estimates) and denominator (ledger actuals/wk)
  are both ship wall-clock engineer+AI hours (#118) — no human-hour
  conversion, and parallelism is already priced into the measured rate, so
  the attention ceiling is a **warning threshold, not arithmetic**. Paused
  projects weigh 0 as named risks; a project with neither resolvable board
  hours nor a Phase-A estimate is `unknown` (weight 0 + warning, never
  silently dropped). `RenderForecast` is the single one-line statement every
  consumer prints. No IO in the core; zero remaining / zero baseline → error
  so callers fall back.
- **Fleet load assembly (`projectforecast.go`, M2 — the IO seam).**
  `ListFleetProjects` builds each active project's contention load via the
  same `computeBoard` the status board uses, reusing
  `project.ListActiveProjectFiles` — the `DiscoverByIssueRef` fleet walk
  factored into a shared `walkFleetProjects` (one fleet-walk source,
  behavior-identical for discovery). A project whose breakdown resolved into
  issues reads board hours even at 0 remaining (a complete project reads ~0,
  not the stale PRD number); Phase-A is a fallback only when nothing resolved;
  a project with 0 remaining or `unknown` load doesn't contend.
  `loadThroughputBaseline` maps absence AND parse failure to one
  `errNoBaseline`; `forecastForProject` is the shared assembly the three
  consumers call.
- **The three consumers — inform, never block (M3).**
  1. **`set-status →committed`** computes the forecast, prints the statement,
     injects it as the `reality-check` evidence when the operator gave no
     `--reality` (so the existing `evidenceGuard` passes on *having computed* —
     zero guard/model change), and derives `planned_finish:`. Precedence
     (`plannedFinishDecision`): explicit `--planned-finish` > a pre-existing
     value > the derived date, each provenance-noted in the Log. The derived
     `planned_finish` is applied to the doc **before** the guard loop, because
     the `baseline-set` guard requires it. No blessed baseline → the legacy
     `--reality` prose fallback (with a bless hint); it never refuses on the
     forecast's *answer*, only on the absence of *any* evidence.
  2. **`project show` / `status`** append the live forecast line
     (`forecastLine`, best-effort — a read verb never fails on a forecast
     error; no baseline → a quiet bless hint).
  3. **`project close`** appends a planned-vs-actual `## Calendar ledger` row
     (`slip_days = actual − planned`, `n/a` when unset) beside the fog row in
     the same `estimate-logic-project-v1.md` — the project-level calendar
     calibration loop, the analogue of the fog factor.
  `appendProjectLedgerRow` is heading-parameterized (`Fog ledger` |
  `Calendar ledger`), each error naming its own heading.

## Artifact migration (`sdlc migrate`, #179)

The write-side companion to resolve: `sdlc migrate <file> <dest-repo-dir>`
moves a markdown artifact to a peer repo with repo-relative refs rewritten
(bare `#N` → `<source>#N`; `<dest>#M` → `#M`; everything else passes through),
because under #171's peer-repo addressing an artifact's home repo is a soft
center-of-gravity default and moves are normal. Same two-layer shape:
`rewriteRefs` (pure — fence-aware via `issue.SplitFences`, inline spans
rewritten only when the whole span parses as one ref, every candidate filtered
through `parseRef`) + `runMigrate` (guards, verification, scoped two-repo
commits, inbound-ref report). Every rewrite is verified from the
**destination's vantage** (`resolveArtifacts` against the dest root) before
any write — fail-closed. Guard inversions worth knowing: migrate deliberately
is NOT in the spine guard set (moving an artifact OUT of brain is its #171 use
case); instead it refuses a brain **destination** (SDLC process artifacts
don't live in brain, #171 amendment). Id-keyed issue-family files refuse (ids
are per-repo sequences; renumbering migration is v2). Round-trips
canonicalize (self-qualified → bare on return) and are then idempotent.

## Progressive disclosure

  - `sdlc --help` — the workflow contract (start-of-work runbook, conventions,
    cobra-generated verb list)
  - `sdlc <verb> --help` — per-checkpoint contract + flags + examples
  - `sdlc state` — runtime "where am I" surface for compaction recovery

`sdlc --help` is the single source of truth for the workflow contract.
`construct/local/sdlc/SKILL.md` (the `xx-sdlc` skill) is a **static pointer**
to it — it carries no copy of the contract, so it can't drift. The old
`sdlc --index` regenerator was retired once the skill stopped duplicating the
help text.

## Architecture

```
cmd/sdlc/
  main.go              cobra root + verb registration
  term.go              cinfo / cok / cwarn / die + env helpers (shared)
  runner.go            gitRunner interface + execGitRunner impl (shared)
  repolock.go          root-level Cobra wrapper for mutating commands:
                       metadata annotation, command-context re-entrancy, and
                       Git-common-dir lock acquisition
  ghclient.go          ghCaller interface + realGH impl (shared)
  preflight.go         runPreflightJudges (push + merge pre-flight)
  close.go             ← scripts/close-issue.py
  actual.go            new (#68): computeActual → internal/activetime → measured --actual (adopted at close when omitted, #178)
  activetime.go        new (#110): `sdlc active-time` CLI (runActiveTime + table renderer)
  internal/activetime/ new (#110): native v3 engine ported from active-time-v3.py
                       (event/commit/segment loaders + Compute; pure core + thin IO seam)
  internal/transcripts/ new (#134): transcript-source harness registry — a Harness
                       per agent CLI (claude.go / codex.go), pure Select aggregator
                       feeding actual.go; adding a harness = one entry
  state.go             new (read-only inspection + drift detection; see "Drift checks")
  resolve.go           new (#144): `sdlc resolve`/`open` — pure parseRef + classifyFamily
                       behind the resolveRepoDir/familyFiles IO seams; read-only, lock-free
  judge.go             ← scripts/pre-merge-checks.sh
  fetch.go             thin hidden alias → runIssueNew --from-github (#56 M2)
  issue.go             new (#56): `sdlc issue` group — new / set-status / list / show / validate (#124)
  project.go           new (#180 M3): thin project new/list/show/validate IO shell
  projectsetstatus.go  project lifecycle legality + named-guard runner; →done
                       delegates to the M4 close verb
  validategate.go      deterministic instance-conformance gate (#124, generalized
                       by #180 M2): noun table enrolls issue + project; push/merge
                       validate frontmatter on every changed instance, with
                       added-only section checks for issues; shells
                       `vocabulary validate-instance --type <noun>`; `--no-validate`
                       remains the loud escape hatch
  start.go             migration stub (REMOVED in #39 — errors with
                       "use claim + change-code")
  claim.go             branch-aware issue synchronization + claim (#39)
  changecode.go        new (#39): planning → implementation gate
  branchcreate.go      new (#39): branch-creation helpers shared by
                       changecode.go (worktree + in-place paths) + the
                       name-resolution previously in start.go. #156: both
                       paths are IDEMPOTENT for milestone re-runs — pure
                       deciders `decideInPlaceBranch` (onTarget/switch/create)
                       + `decideWorktreeBranch` (reuse/addExisting/addNew)
                       over `currentBranch`/`branchExists`/`worktreeForBranch`
                       probes, so re-running change-code on an existing branch
                       switches/reuses instead of dying `checkout -b: already
                       exists`. `worktreeForBranch` filters the single-source
                       `parseWorktrees` (extracted from state.go's listWorktrees;
                       findMainWorktree refolded onto it — one porcelain grammar).
  setstatus.go         new
  push.go              ← Makefile push:
  pr.go                ← Makefile pull-request:
  merge.go             ← Makefile merge:
  milestoneclose.go    composition over close + judge milestone-review
  estimatesource.go    new (#134): `sdlc estimate-source` pull — names the shared
                       method + repo-local calibration doc (estimateSourceStatus
                       seam over estimate.SourceGuidance); start-plan/change-code push it
  helptext/            //go:embed *.md — one .md per verb + root
  internal/
    repolock/          local repo transaction lock: pure holder metadata /
                       stale-observation decisions + thin mkdir/meta.json
                       acquire/release IO shell
    gitx/              git invocation seam (`run` shim, Capture, DiffBase,
                       MainRef, CommitWindow, WorkingTransitionISO (#113 claim
                       anchor), DiscoverWindowIssues, RunGit,
                       IsShippedWorkSubject/ShippedWorkOnMain — #76 ship probe)
    issue/             frontmatter parse/edit + plan-section regexes +
                       scaffold.go (NextID/Slugify/Render — #56)
    judge/             Category enum, prompt builder, classify, dispatch
    project/           project-file core: line-preserving typed Doc/Task parser +
                       checkbox/frontmatter/section mutations (#180 M2); model-derived
                       scaffold, pure summaries, and pure named guards (#180 M3), alongside
                       the legacy brain-residency lookup/detail-block helpers (#171
                       will lift residency)
```

## Drift checks (`sdlc state`)

`state` gates nothing — it reports. `detectDrift` (state.go) surfaces warn-only
inconsistencies so an agent recovering after compaction sees what's off:
missing-frontmatter, `done`/`wontfix`/`punt` still in `workshop/issues/` (should
be archived), and `working` with **zero** plan items ticked (the no-progress
end). #76 added the inverse — the **close-off candidate**: an `open`/`working`
issue whose plan is all (or all-but-one) ticked *and* whose work has shipped to
main, i.e. done work that never got formally closed. The "shipped" signal is
gh-free by design: a subject-anchored `git log <MainRef>` scan
(`gitx.ShippedWorkOnMain` → pure `IsShippedWorkSubject`) distinguishes a real
`#N Mx:` work commit from bookkeeping (`file issue`/`ticket`/`claim`/`close`),
so a merged PR's work commits *are* the signal — no network, degrades to
"not-shipped" when there's no main ref. Warn-only on purpose: it points at `sdlc
close --issue N` for a human glance, never auto-closes (closing carries
actual/verified judgment a heuristic can't supply). The probe is injected into
`detectDrift` as a `shipProbe` func so the drift logic stays testable without
git.

## Friction audit (`process-manual --friction-report`, #172)

The per-gate `--no-<gate>` bypass design (below) was built to make bypasses
*explicit and measurable*; `process-manual --friction-report` closes that loop. It
measures, across the **whole Claude corpus** (all repos), where the spine creates
friction — per-gate bypass rates, and (M2+) refusal→retry loops + firing-order
anomalies — as a clean, command-anchored measure that replaces a contaminated grep.

The core problem is **discrimination, not capture**: this repo *develops* sdlc, so
`close.go`'s source and cat-n log reads spray every `--no-<gate>` string into tool
output. So the instrument (1) **anchors to `Bash(sdlc <verb>)` invocations** (drops
the source/edit/log-read noise — the dominant contamination), joined to their
`tool_use_id`-linked tool_result content-block; (2) classifies each output line
against a **per-gate signature catalog** (`internal/processmanual/gatesig.go` —
12 gates / 16 sigs / 3 ACK grammars: G1 close·mclose `--no-X (or --force): …`,
cinfo no-judge, G2 change-code `X gate bypassed (--force: …)` (silent alone), G3
merge/push `--no-X: …`), requiring the runtime `\x1b[0m` reset for a bypass ACK and
grammar+digit-anchored patterns for refusals (so the `printSemanticWarmup` success
line and source restatements are rejected); and (3) states **observability limits**
honestly (change-code silent-alone bypasses countable only via `--force`; merge/push
refusals that don't name the flag). A **cross-command drift guard** (`gates_test.go`)
asserts the catalog matches each command's registered `--no-*` flags, so a new gate
can't be added without the audit noticing (ARCH-DRY). Codex coverage (M3) derives a
net-new Go parser from `atlas/workflow/introspect.md` → "Codex transcript format" —
the two can't share code (Python vs Go), only the spec.

M1 headline over the real corpus: `--no-judge` dominant, `--no-verified` = 0 (the
gate design works), and bypasses concentrate in **peer repos, not ariadne** — the
substrate repo follows its own gates; lighter repos route around them. The raw grep
over-counted ~4× (unlinked echoes: process-manual outputs, transcript reads).

The M2 detectors ride the same invocation stream (`buildFrictionReport` is the
composition seam; gate events are deduped per invocation — one no-validate refusal
prints two matching lines):

- **Refusal→retry** (`detectRefusalRetries`) pairs each gate refusal with the next
  same-verb+same-issue invocation in the same transcript; `resolved` distinguishes
  satisfying the gate from routing around it (`via bypass`). merge/push refusals
  never name their flag → paired by verb+context, labeled `flag-omitted`.
- **Firing-order** (`detectFiringOrder`) walks each (repo, issue)'s invocations
  across transcripts against the AGENTS.md §2 ladder (claim ≺ start-plan ≺
  change-code ≺ milestone-close ≺ close ≺ merge/push). Only an observed **order
  inversion** flags (change-code after a clean close/merge); legal loops —
  mclose→change-code, start-plan re-runs, close→change-code after a REWORK verdict
  (recovered via `judge.ParseVerdict`) — stay silent, and absent early stages are
  partial observation, not anomalies (precision over recall). merge/push carry no
  `--issue` → attributed from segment context or counted unattributed. The
  `skill-late` arm flags a plan/TDD Skill load after a non-`.md` file edit in the
  same segment+issue (Edit/Write/MultiEdit → `KindFileEdit`, excluded from
  `--session` reports).

M2 headline: refusals mostly resolve by **satisfying** the gate (no-estimate-recon
19/19 satisfied; via-bypass is rare — no-verdict 2, no-atlas 1), i.e. refusals do
their job; firing-order found change-code-after-close 16 / skill-late 2.

**M3 — codex coverage (both agents).** The walk also enumerates
`~/.codex/sessions/**/rollout-*.jsonl` via `codex.go`, whose format knowledge
derives entirely from `atlas/workflow/introspect.md` → "Codex transcript format"
(the spec is the DRY point with Python introspect; a cross-language golden at
`internal/processmanual/testdata/codex-golden/` pins the shared keep/skip +
classification decisions). Fork-replay rollouts (`forked_from_id` on the FIRST
`session_meta`) are skipped — the real corpus skips exactly the spec's 40 —
while sub-agent threads are processed; repo labels come from `session_meta.cwd`
(worktree-normalized). sdlc's ANSI survives codex's exec_command wrapper, so the
SAME `classifyOutputLine` serves both agents. `SdlcInvocation.Failed` (Claude's
`is_error` flag / codex's non-zero `Process exited with code N` — deliberately
NOT the spec's hint-gated taste-friction `is_error`) keeps failed invocations
from raising the firing-order ladder. Report adds the per-agent bypass split +
forks-skipped counter; skill-late is Claude-only (codex has no Skill tool).

M3 headline: codex 43 vs claude 37 bypasses; codex's signature move is the
**re-close** (no-reclose-guard: 25 bypasses, and its 3 refusals all resolved
**via bypass** — the one gate agents route around after refusal); no-actual
refusals jump to 38 corpus-wide (35 satisfied). Firing-order anomalies unchanged
by codex (16 + 2).

## Anti-collusion + form-vs-essence

Checkpoint guards defend against **omission** (claiming done without
doing) via deterministic checks (`close` refuses without `--actual` +
`--verified`). The judge subcommand defends against **theater** (form
without substance) via fresh-context LLM review — every Dispatch call
spawns a new subprocess; the agent has no doer-session state.

**Judge agent defaults (#129).** Every judge dispatch (`sdlc judge`,
`change-code` plan/estimate quality, close and milestone boundary reviews, plus
push/merge preflight) resolves through `internal/judge.ResolveAgentCLI`. The
precedence is: explicit `--agent`, then `AGENT_CMD` (operator/script override,
including invalid values so dispatch validation still fails loudly), then
`PAIR_AGENT`, then conservative process/session signals such as `CODEX_CI` /
`CODEX_THREAD_ID`, then `claude`. Command flags stay empty by default so
environment defaults live in one place; call sites pass an explicit-source bit
from Cobra's `Flags().Changed("agent")`. Close dry-run uses the same boundary
review prompt/options builder as real dispatch and prints the would-be command
line, so `PAIR_AGENT=codex` is inspectable as `codex exec` before any subprocess
runs.

**Dispatch progress heartbeat (#140).** A boundary review can run silently for
minutes; `internal/judge.Dispatch` now emits a heartbeat to `opts.Stderr` every
`heartbeatInterval` (30s) while the agent subprocess runs — elapsed + agent +
child PID (`… still working — 1m0s elapsed via claude (pid N; inspect: ps -p N)`).
It is harness-agnostic by construction: all three fields come from `sdlc` wrapping
the child (`Run` now Start→`onStart(pid)`→Wait into separate stdout/stderr
buffers, exposing the PID), not from child output, so it reads identically for
claude/codex/gemini.
Gated on `opts.Stderr != nil`, so the fast path (unit tests, quick dispatches)
stays synchronous and silent. `classifyRunResult` is the shared process-to-review
boundary: it forwards captured stderr to the diagnostic sink and returns stdout
only, so `Classify`/`ParseVerdict`/the sidecar never consume harness diagnostics.
Deliberately no byte-count/log-tail signal — `claude -p` buffers to the
end (a live counter would read "0 B" and look stalled) and no agent exposes a
reliably-locatable per-invocation log; the PID is the automated form of the
operator's manual `ps` inspection.

**The estimate shell (#117)** applies this same form/essence split to
`estimate_hours`, plus a feedback arm — the one forecast in the system with a
deterministic ground-truth measurement (`sdlc actual`). *Form:* the change-code
**estimate-reconciliation** gate parses the issue's `## Estimate` fenced block
(`internal/estimate`, pure) and refuses unless `estimate_hours` reconciles with
an itemized v2-lineage primitive derivation (`Σdesign×(1+buffer)+Σimpl×familiarity`). No
unitemized estimate — a fabricated number can't pass. *Essence:* the
**estimate-quality** judge checks the derivation was applied, not back-fit.
*Feedback:* on a full-issue close `sdlc close` appends every numeric estimate↔actual pair
to `brain/.../velocity/calibration-ledger.tsv` (`$WF_CALIB_LEDGER` override) and
flags >2× same-direction drift over the last N unique **window-trusted** rows for
the latest recognized model revision — pre-#116 rows are `window-trusted=no` and
excluded (a truncated actual isn't a clean point; `actual_hours: N/A` rows are
excluded; duplicate append rows for the same issue/model count once;
skip-with-warning when no brain/ ledger dir exists downstream). This
closes the loop the hand-kept validation log never did. Grammar + closed
vocabulary: `helptext/estimate.md` (canonical slugs in
`internal/estimate/vocab.go`, a bidirectional drift test guards the mirror).

**Estimator-source discovery (#134).** The shared method is single-sourced in
`sdlc` (the grammar + `Models()`/`Primitives()`), but the *repo-local
calibration* — the actual per-primitive hour ranges, which drift as closes accrue
(#127) — lives in a brain artifact. So an agent could satisfy the block grammar
while picking hours from memory. `sdlc estimate-source` (`estimatesource.go`, the
pull; mirrors `arch-principles`) names BOTH in one output: the shared-method
pointer + the resolved calibration doc for `estimate.CurrentModel()` by default
(`estimate.SourcePath` →
`$WF_ESTIMATOR_SRC` override, else `estimate.VelocityPath(brainDir,
<model>.md)` — the same `data/life/42shots/velocity/` builder `close.go`'s ledger
path now derives from, ARCH-DRY), tagged `ok | stale | MISSING` (stale = sibling
ledger newer than the doc). The pure renderers (`SourceGuidance` full / `SourceLine`
one-line) live in `internal/estimate/source.go`; `estimateSourceStatus` is the
thin stat/mtime seam. Fail-loud is asymmetric on purpose: the PULL exits non-zero
when the source is missing, while the PUSHes (`start-plan` emits `SourceLine` after
the estimate nudge; `change-code`'s missing-block error points at the command)
warn-and-continue so a brain-less downstream repo never breaks the gates.

**Per-gate bypass (#67).** `close` has 8 gates (actual, verified, atlas,
milestone-verdict, plan-unchecked, project, re-close, and the #69 boundary
review), each with its own `--no-<gate>` flag (`--no-actual`, `--no-verified`,
`--no-atlas`, `--no-verdict`, `--no-plan-check`, `--no-project`,
`--no-reclose-guard`, `--no-judge`);
`closeFlags.skip(gate)` is the single arbiter (`Force || the field`). A
per-gate flag is an *acknowledgment* that one guard doesn't apply (e.g. a
pure bugfix → `--no-atlas`); it logs an audit `[!]` line and only fires
when the gate would actually have refused. `--force` waives all at once.
The atlas gate additionally auto-satisfies (info line, no flag needed) when
the window contains no code surface — `hasCodePath`, the single docs
classifier: `*.md` / `workshop/` / `atlas/` / `docs/` are documentation,
everything else is surface (#177; docs-only closes had an incoherent demand).

The milestone-verdict gate demands per-milestone `Review-Verdict:` evidence,
EXCEPT for **trailing** unclosed milestones (#175): Mx rows after the last
verdict-carrying one — or all rows, when none carries a verdict (the
single-pass over-split shape) — are accepted with a loud info line, because
the issue-close boundary review's window (branch-point→HEAD) covers their
work and the close *is* their first boundary. `partitionMissingVerdicts`
(pure, `close.go`) does the split; **midstream** misses (a later milestone
closed WITH review) still refuse — that boundary was genuinely crossed
unreviewed — and `--no-judge` turns trailing acceptance back into a refusal,
since it skips the very review the acceptance leans on. The plan-quality
judge flags over-split Mx plans at design time (the forward fix).

**Off-workflow guards (#176, `repoguard.go`).** The 7 lifecycle verbs (=
`processmanual.WorkflowVerbs()`, enforced by a drift test) refuse up front in
a brain repo (`.brain/config.md` — the #172 audit found brain concentrating
bypasses because its own merged constitution invited sdlc) and in repos
without `workshop/issues/`; reads (estimate-source, actual, state,
process-manual, issue …) are unguarded by construction since sdlc legitimately
reads brain. `WF_SPINE_GUARD=off` is the single emergency hatch and cwarn-ACKs — greppable
in transcripts (GateCatalog wiring for this env-gate family is deliberate
follow-up work; the instrument today derives only from catalog rows). `guardIssueNotDone` closes the front
door on terminal issues: start-plan/change-code refuse on `status: done`
(re-close was already guarded at the back).
`milestone-close` forwards the same flags into its delegated `computeClose`
(the #139 compute→review→finalize; `runClose` is now test-only, #146).
The convention generalizes `merge`'s pre-existing `--no-judge`.

**Measured actuals (#68, #110).** `--actual` is computed, not hand-typed. `sdlc
actual --issue N` (`actual.go`'s `computeActual`, shared with close's
missing-`--actual` explainer) runs the native **`internal/activetime`** engine
(`activetime.Compute`, in-process — no python3) over the issue's `CommitWindow` +
`DiscoverWindowIssues` peers, feeding it **brain + the issue's repo** transcript
sources. Source selection is a **harness abstraction** (`internal/transcripts`,
#134), not a Claude-only path convention: each agent CLI implements a `Harness`
(`Name()` + `Sources(cwds) → Sources{Dirs,Files}`); `DefaultHarnesses()` is the
registry; a pure `Select` merges their contributions for the brain+repo cwds.
Two harnesses ship — **Claude** (one `~/.claude/projects/<cwd-encoded>` dir per
cwd → `Dirs`, engine globs them) and **Codex** (date-sharded
`~/.codex/sessions/YYYY/MM/DD/*.jsonl` filtered by `session_meta.cwd` → `Files`,
since Codex stores cwd inside the file, not the path). A third agent CLI is one
new `<harness>.go` + one registry entry — `actual.go` and the engine never change
(ARCH-PURPOSE). Pure encoders/parsers (`cwdToClaudeDir`, `codexCWDFromBytes`) are
unit-tested directly; each `Sources` method is the thin IO seam, tolerant of
malformed / empty / no-`session_meta` Codex files. This is the validated heuristic (events come only from transcripts;
the wrong/missing sources were why actuals read 0 and got faked). The engine
returns a structured `Result.Status`: `TelemetryGap`
(commits-but-0-events → labeled judgment), `EmptyWindow` (nothing to measure), or
`Measured` (`PerIssue[N]` → hours). Attribution is global-boundary based (#92):
source-scoped activity runs are claimed by nearby issue-referenced commits;
overlaps collapse within one transcript source but not across parallel sources;
all commit-subject issue refs become claimants, while no-ref commits are neutral
time boundaries. Suspicious attribution is surfaced as `Result.Warnings` and
rendered by `actual` / `active-time`. Dir-selection is deliberately narrow (NOT all
folders/sessions) — an unrelated concurrently-edited repo inflates the count.
`WindowCapDays` is 61 (was 31) so month-long issues keep their window. The
window-**start** is the *earlier* of `CommitWindow`'s parent-of-first-`#N`-commit
and the **engagement anchor** (`resolveWindowStart`), anchoring at the cheap early
`claim` so DESIGN attention (brainstorm / spec / plan / reviews) before the first
code commit is in-window instead of cut off; gap-truncation keeps a dormant
claim→work gap from inflating the actual. The anchor is resolved in robustness
order (#116): the explicit `started:` frontmatter stamp (written once at the
open→working flip in `applyStatus`, local-offset RFC3339 to match `%aI`) →
`gitx.WorkingTransitionISO` (the #113 git-log heuristic, now the legacy fallback)
→ commit-parent. The explicit stamp survives rebases/moves where the heuristic's
"best-effort" history scan could silently miss and drop design time.

`sdlc active-time` (#110) is the standalone CLI over the same engine — the
manual-inspection sibling that prints the full attribution table and warnings. It preserves
the #68 loud-fail exit-code contract: **2** = misinvoke (no `--dir`/`--issue`/
`--git-repo`), **3** = telemetry gap (commits but 0 events), **0** =
measured-or-empty. (Before #110 the engine was `active-time-v3.py`, a python3
subprocess whose human-formatted stdout `actual.go` regex-scraped; #110 ported it
into the binary and deleted the script.)

**Passed-`--actual` backstop (#87).** When `--actual` *is* given, close still
runs the engine and compares (`actualDeviation`, the pure comparator in
`close.go`): ratio ≥3× → warn, ≥10× → **refuse** (bypass with `--force`), with a
0.5h absolute floor so small gaps don't trip. Skips silently when the engine
can't measure. Closes the hole where a hand-typed value (the failure #86's docs
prime against) was trusted blindly — the doc fix removes the priming, this
removes the blind trust. `milestone-close` inherits it (computes via `computeClose`).

`push` and `merge` auto-dispatch `judge plan|specs|lessons` as pre-
flight so the checks run consistently rather than as a remembered
manual step. `milestone-close` auto-dispatches `judge milestone-review`
as a post-action.

**Judges are read-only (#62).** A judge is a reviewer, not a doer — all
categories run with a read-only tool allowlist (`Read,Grep,Glob,Bash`); they
report findings and the main agent (full context) applies fixes. The `specs`
judge used to auto-edit stale docs (`Edit,Write`), which let a *passing* gate
leave the tree dirty and strand the subsequent merge. `merge` now also (a)
re-asserts no **tracked** dirt immediately before the irreversible `gh pr merge`
(refuse, don't strand), and (b) resumes an interrupted merge — a re-run detects
an already-merged PR and finishes the local cleanup instead of erroring. The
resume path is guarded (#148): before cleanup it fetches `origin/main` and counts
`origin/main..origin/<branch>` (`countUnmerged`, fakeable seam); a nonzero count
means a **reused branch name** (its old work shipped via that PR, new work piled on)
→ `decideMergeAction` returns `actionResumeBlocked` and merge refuses *before* any
switch/delete/archive, rather than silently stranding the new commits.
The clean-tree guards (step 2 and the 9b re-assert) refuse only on tracked
**code** changes via one pure `assessDirty(...).Refuse()` decision (#78);
`assessDirty` buckets each porcelain line into Blocking / Untracked / Tracker.
**Untracked** files survive `git switch main`, so they're surfaced as a warning,
not a blocker — unrelated local WIP no longer forces a stash-around-the-merge
detour. **Tracker** files (`workshop/issues|history/NNNNNN-*.md`) are likewise
never blocking, tracked-modified *or* untracked (#82 M2): they're append-only
shared state synced to main out-of-band (#82 M1), not code contention, so a dirty
issue file never gates a merge. (Path matching reuses push.go's
`isIssuePath`/`isHistoryPath`; the path is pulled by field-split, not column
slice, since `worktreeDirty` whole-trims and strips the first line's leading
status space.)
Both behaviors have e2e regression coverage (#63, `merge_e2e_test.go`): a
`tempRepo(t)` harness runs `runMerge` against a real throwaway repo + local bare
origin (in-place topology), so switch/pull/archive/branch-delete execute for
real. The unlock is a trio of `func`→`var` test seams — `die` (term.go,
swapped for panic+recover via `expectDie`), `detectRepo` (fetch.go), and
`runPreflightJudgesFn` (merge.go step 5, used to inject a tree-dirtying
"judge"). `expectDie` is the reusable pattern for testing any `run*` verb's
refusal path.

**Judge → classifier contract (#70).** One contract, both sides reference it:
the human mirror is `construct/judge-output-contract.md`; the Go source of truth
is `internal/judge/contract.go` (`ContractTokens`, `Blocking()`, and
`ContractPreamble` — the format snippet every agent prompt embeds). Every
agent-emitting judge leads its response with `VERDICT: <TOKEN> (confidence: …)`;
the classifier (`ParseVerdictToken`) scans for that token **anywhere** (tolerating
a preamble — the `VERDICT:` prefix + a trailing precision guard make prose
false-positives near-impossible) and gates on the token's **blocking-ness**, never
on prose presence. Tokens: `CLEAN INFO SHIP FIX-THEN-SHIP` pass; `FAILURE REWORK
BLOCK` block. **Boundary review (#147): the handoff is now block-first.** The
reviewer emits a fenced ```` ```verdict ```` block validated against `verdict.cue`
(`ParseVerdictBlock`, the authoritative structured handoff — see [Vocabulary](vocabulary.md)
+ the `agent-binary-handoff-schema` target); the prose `VERDICT:` line is the
logged fallback. So a prose-narrated verdict no longer degrades to `unknown`. Lessons is the exception — a fixed `REMINDER:` line, no agent. This
killed the bug where a `VERDICT: CLEAN` behind a title scored `FAILURE` and blocked
the merge (and the milestone `unknown`). A thin legacy sentinel-grep remains for
un-migrated/foreign outputs; a `judge_test.go` drift test keeps the doc + Go
tokens in sync.

**Architecture principles (#75).** Injected architectural taste the model lacks
(payoff shows months downstream → no training signal). `internal/judge/
architecture.md` is the single source — markered `ARCH-*` entries, each with a
`principle` / `at-plan` / `at-review` lens — `//go:embed`'d as
`ArchitectureRegistry` and delivered verbatim into the prompts that need it (one
file, embedded per fresh context). Today: the **plan-quality** judge renders the
`at-plan` lens (highest leverage — the design is still changeable), the
**milestone-review** judge renders `at-review` (backstop), and the standalone
**dry/pure** judges render their principle from the registry (authored once).
The **estimate-quality** judge (#117) is a change-code-time-only sibling of
plan-quality — deliberately NOT in `AllCategories()` so it never enters push/merge
bulk dispatch; it checks the `## Estimate` derivation was *applied*, not back-fit.
Cite the marker (`ARCH-DRY`) in plans/Logs/findings. Adding an `ARCH-*` entry
flows into every consumer with no other edit. **`sdlc start-plan`** (#75 M2)
delivers the `at-plan` lens to the main thread at design time — the forward
counterpart to `change-code`'s plan-quality review (`claim → start-plan →
change-code`). #71 adds `ARCH-SHIM`. #126 landed `ARCH-PURPOSE` (serve the issue's
actual purpose; single-source ⇒ every consumer *derives* — the registry's 3rd
marker, disambiguated from Simplicity-First/YAGNI). **#205** adds
`ARCH-CONSTRAINTS`: classify the workload and interaction path, declare material
runtime budgets/ranges with their evidence basis and exceeded behavior, then
review representative measurements against that operating envelope. It supplies
domain prompts rather than universal performance defaults. **#128** added
**`sdlc arch-principles`** — a standalone command that prints the registry (the
same `ArchitectureBlock` primitive start-plan calls), the *pull* path for non-gate
work (autonomous fixes, quick edits, Q&A) that never hits start-plan. AGENTS.md's
narrative collapsed from a per-marker restatement to a *route* to that command
(definitions single-sourced; the drift test — `TestArchitecture_NarrativeRoutesToArchPrinciples`
— now guards the route + marker awareness, not an enumeration). The start-plan
*push* stays: a gate-time push beats a model-dependent pull (the same asymmetry
that keeps the judges' inline embed). **#82 M3 / #83** also have `start-plan` print non-blocking
**dependency-path contention** — one line per repo this one reads live. The
symlink model means a repo reads ALL its transitive upstreams' working trees, so
the "moving ground" is the whole dependency chain, not a single base.
`substrateChain(root)` walks `construct/deps` transitively (resolving each
`substrate <path>` against its *declaring* root; `parseSubstrateTargets` is the
Go port of `lib-deps.sh deps_substrate_targets`, kept in lockstep); per upstream,
`gatherBaseContention(root, …)` reads branch + dirty *code* count (tracker files
excluded, reusing M2's `assessDirty.Blocking`) + other `status: working` issues,
and the pure `baseContentionSummary` renders the line (clean → green "clear to
plan"). The **root** (ariadne — no upstream) reports its own concurrent work; a
**derivative** reports its upstream(s) (`base (ariadne): …` from a nous session),
which is the Spec's primary case. (#83 replaced #82 M3's broken `isBaseRepo`
heuristic — `construct/` is a real dir in *every* repo, so it fired everywhere.)
It never refuses. **#72** adds a third payload between the architecture block and
the contention line: the pure `planPointer(issue)` durable-plan reminder — author
the plan via the `superpowers-writing-plans` skill into
`workshop/plans/NNNNNN-slug-plan.md`, never the harness builtin's ephemeral
`~/.claude/plans/`. So start-plan emits, in order, *what to design against*
(architecture) → *how/where to capture it* (the skill pointer) → *the moving
ground* (contention).

**One boundary review, binary-owned (#69).** The *procedure* and the *principles*
are separate embedded sources: `internal/judge/code-review.md` (`//go:embed`'d as
`codeReviewTemplate`, rendered by `CodeReviewBody`) is the **one reviewer prompt**
— the superpowers quality/testing/readiness checklist reconciled with ariadne's
Core-concepts cross-check, Atlas gate, severity buckets, and the
`SHIP|FIX-THEN-SHIP|REWORK` verdict. It *refers* to the ARCH-* markers (the
`{{ARCH_STAR}}` token expands to the live marker list via `ArchitectureMarkers()`,
the single extraction site shared with the AGENTS.md drift test); the principle
*definitions* arrive co-present from `ArchitectureBlock("at-review")` at dispatch.
The procedure must not inline principle bodies (a guardrail test pins this). Both
`milestone-close` (per-milestone window) and `close` (whole-issue / end-of-issue
window) dispatch this same review — so the agent does **not** run a separate
`superpowers-requesting-code-review` pass at a boundary (AGENTS.md §3); that skill
remains for ad-hoc/in-session reviews. The double-review #69 removed was the
agent's superpowers pass *plus* the binary's auto-dispatch on the same diff.

**Pinned review manifest (#162).** Boundary prompts no longer embed the unified
patch. `ReviewWindowManifest` (`internal/judge/reviewwindow.go`) purely validates
and renders the repository root, immutable base/head commits, issue/optional plan
paths, exclusions, and stat/name-status/full/targeted Git argv. The thin
`resolveBoundaryReviewManifest` IO seam (`cmd/sdlc/reviewwindow.go`) pins refs and
validates paths through a stateful Git-runner fake plus live-repository
conformance. Automatic close requires already-concrete anchors captured under
the repo lock and keeps the historical `workshop/history` exclusion; manual
`judge milestone-review` resolves supplied refs, respects its issue/history/plan
directory flags, and keeps issue-less ad-hoc review valid. Omitting manual
`--head` means base-vs-working-tree: committed-after-base plus staged/unstaged
tracked changes are in scope, untracked files are not. Reviewers must run the
manifest's read-only inspection recipes and return REWORK if the repository,
objects, or commands are unavailable. This keeps agent argv bounded even for a
multi-megabyte legitimate window while Git remains the patch source (ARCH-DRY,
ARCH-PURE, ARCH-MOCK).

**Repo orientation (#137).** The review prompt orients the fresh reviewer to the
**actual repo under review**, derived from the live git context — not a hardcoded
`ariadne`. `boundaryOrientation` (`cmd/sdlc/orientation.go`) computes the repo name
(git-root basename), root path, the `<repo>#N` issue ref (so a `pair` review reads
`pair#72`, never `ariadne#72`), issue file, boundary kind, and a base-vs-downstream
note (base detected via `construct/base.manifest`); these are passed as plain
strings into the pure `internal/judge` layer (`PromptInput` → `code-review.md`
header), keeping git IO at the cmd boundary (ARCH-PURE). Computed once in the
shared `boundaryReviewDispatchOptions` (ARCH-DRY); the same derivation feeds the
sidecar H1.

**Two-phase close, finalize-after-verdict (#139).** `sdlc close` and
`sdlc milestone-close` no longer write before the review. `runClose` splits into a
read-only `computeClose` (all gates + composes the new issue/project text → a
`closeResult`, writing nothing) and `applyClose` (the writes + calibration ledger).
Full-issue close and milestone-close both **compute → review → finalize**: the
boundary review runs against the *un-mutated* working tree (the reviewer reads the
honest `status: working` issue), and `applyClose` fires only on a **finalizing**
verdict via the shared finalization helper. The command path releases
`.git/sdlc.lock` while the external review subprocess runs, then reacquires and
checks that HEAD, the issue file, and any prepared project-file edit still match
the reviewed snapshot before writing. The canonical durable plan candidate is in
that same artifact snapshot even when absent, so creation, deletion, modification,
or replacement during the unlocked review also refuses finalization.
`closeVerdictOutcome` derives from
`vocab.Verdict()` (#147): finalizing (SHIP/FIX-THEN-SHIP) → finalize; blocking
(REWORK) → **not finalized**, issue left `working`, non-zero exit, "fix + re-run"
(no `--no-reclose-guard` needed on the rerun since it never went `done`);
unknown / dispatch-error → **halt**: don't finalize an ambiguous gate — stop and
consult a human. `--no-judge` finalizes (explicit operator skip, handled before
dispatch). The success messages ("flipped → done") print only from `applyClose`, so
a REWORK never claims a write that didn't happen.

**Subprocess PATH (#138).** The agent subprocess `sdlc` spawns for a review (and
for `sdlc judge`) gets the owner `bin/` prepended to its `PATH`, so a fresh
reviewer can resolve `sdlc` (and sibling tools) even when the spawning shell's
startup files never put `ariadne/bin` on `PATH`. The dir is `dir(os.Executable())`
(single-sourced via `ownerBinDir`, in `internal/judge/dispatch.go`), injected at
the one launch seam (`Run`) via the pure `binAugmentedEnv` — so it works from
downstream repos (the binary is `…/ariadne/bin/sdlc` regardless of cwd) with no
dependence on the user's `~/.zshenv`/`~/.bash_profile`. Launch-failure errors name
the attempted agent + that bin dir.

**Review sidecar (#136/#201).** The boundary review is no longer a transient terminal
artifact: every actually-dispatched review writes its semantic final response to a durable
sidecar under `workshop/plans/` — `NNNNNN-slug-close-review.md` for a whole-issue
close, `NNNNNN-slug-m<x>-review.md` for milestone `Mx`. The write lives in the
single shared `dispatchBoundaryReview` (`reviewsidecar.go`: pure `sidecarMeta` +
`renderReviewEntry` + `sidecarPath` behind a thin atomic-write seam — ARCH-PURE),
so both close paths inherit it for free (ARCH-DRY). Each file carries a metadata
header (issue id/title, repo, issue file, boundary kind, milestone, base..head
window, command, reviewer, timestamp, verdict) plus that response. Harness
diagnostics, progress, prompt echo, and tool transcript remain terminal stderr
and never enter verdict/finding parsing or either durable sidecar. A re-run of the same
boundary **appends** a timestamped `## Re-review` section rather than overwriting
(the §1 revision convention). The terminal still prints the semantic body + the
`Review-Verdict:` trailer; the sidecar adds a durable surface an agent can reopen
after scrollback loss or compaction (the path is echoed as `review sidecar: …`).
`--no-judge`/`--dry-run`/not-run boundaries write nothing — no body to persist.

**Window base — prior review boundary (#58).** `boundaryWindowBase`
(`milestoneclose.go`) is the single source for *both* the atlas-coverage gate
(`computeClose`) and the boundary review's window, so they provably cover the same
commits (ARCH-DRY). A milestone window bases on the **previous review boundary**
— the most recent prior commit touching the issue file that carries a
`Review-Verdict:` trailer (the prior milestone close), found by
`previousReviewBoundary` — not on the first `#N Mx` commit. This closes a gap
where an inter-milestone `#N`-but-not-`Mx` commit (a `side-quest:`, a fix) landed
between M(x-1)'s close and Mx's first commit would slip *both* windows and escape
review. The first milestone (no prior boundary) uses the feature branch point,
so an issue filed early cannot pull unrelated main history into M1. If a prior
close's trailer was never pasted, the lookup likewise uses that branch point —
over-covering prior branch work rather than under-covering. Only direct-on-main /
no-divergence work falls back to the parent of the first `#N` implementation
commit.

The **whole-issue** close (the end-of-issue integration review) bases on the
**branch point** — `gitx.MergeBaseWithMain()` = `merge-base(main, HEAD)` — so it
windows exactly this branch's commits, not unrelated history merged onto main
before the issue's first commit (#77; an issue filed early but implemented late
otherwise over-captured everything since it was *filed*). On `main` (the direct
`sdlc push` flow, no divergence) merge-base == HEAD, so it falls back to the
issue's branch start. `MergeBaseWithMain` is deliberately separate from
`DiffBase` (the `sdlc judge` window): same merge-base core, but it returns `""`
on no-divergence so the caller can pick the issue-specific fallback.

## Build + install

```
make ensure-go         guarantee the Go toolchain (#61): no-op if present,
                       brew-installs on macOS, else fails fast with go.dev/dl
make sdlc-build        builds bin/sdlc (build-in-owner since #60 — see gotcha);
                       depends on ensure-go
make sdlc-install      build + append the repo's bin/ to the shell PATH
                       (`sdlc-bootstrap` is a back-compat alias)
```

`make build` also picks `sdlc` up via the cmd/*/main.go scanner.

**Go is a base-layer build dependency (#61).** ariadne ships `cmd/sdlc` and
compiles it in `tools`, so `bootstrap` provisions Go up front (`ensure-go` is its
first prerequisite) — before the peer-clone cascade and the recursive ariadne
bootstrap's tool build. Pre-sdlc, ariadne needed only shell + python (always
present), so bootstrap never provisioned a toolchain; #61 closed that gap. nous
owns its richer toolchain (Homebrew/GPG/gh/…) separately.

### Downstream staleness gotcha

Downstream repos ship a *prebuilt* `bin/sdlc` — they have no `cmd/sdlc` source
of their own. As of #60 `make sdlc-build` resolves sdlc's **owner** (ariadne) by
location via `construct/dev-aliases.sh --list`, then builds in the owner's own
module into `<repo>/bin/sdlc` (build-in-owner) — no `construct/go.mod`
`replace` needed. That binary does **not** auto-rebuild when the base-layer
tool changes; it goes stale until the operator reruns `make sdlc-build`
(or `make sdlc-install`) in the downstream repo. A stale binary silently
lacks new behavior and can fail in confusing ways — e.g. a pre-#51 binary
hits the in-place branch flow's `git rev-parse --git-dir` → `.git` path but
still routes to the worktree topology, dying with `find main worktree: could
not find a worktree on branch 'main'`. Surfaced live by the #51 dogfood
(ariadne #53 Phase B): you-decide's binary was a month stale. **Rule:** after
any base-layer `cmd/sdlc` change reaches downstream, rerun `make sdlc-build`
there before relying on the new verb behavior.

### `sdlc push` archive recovery

`sdlc push` has a recovery preflight before its normal untracked-file guard.
If a previous push already moved terminal-status issue files from
`workshop/issues/` to `workshop/history/` but failed before the archive commit,
the working tree contains deleted issue files plus untracked or staged history
files. That state is otherwise indistinguishable from forbidden untracked files
to the generic guard, so `push` recognizes exact prepared archive pairs,
verifies the history copy still has a terminal status, stages **only those exact
moved paths** (`git add -- <src> <dst>` via the shared `archiveAddArgs` helper —
never a directory-wide `git add issues/ history/`, which would also sweep
unrelated untracked WIP onto main, #80), commits "archive completed issues to
history", pushes, and exits without rerunning judges against an archive-only
retry. The same precise-staging helper backs the non-recovery archive commit in
both `push` and `merge`. The archive sweep also moves the issue's
`workshop/plans/NNNNNN-*` artifacts (durable plan + review sidecars) into history
alongside it (`archivePlanArtifacts`, #143); recovery reconstructs those plan
moves too, but — since plan artifacts carry no terminal frontmatter — gates them
on their id-prefixed plans-dir source rather than the terminal-status check used
for issue files. A review sidecar is frequently **untracked** at ship time
(`sdlc close` writes it after the implementer's last commit, and a FIX-THEN-SHIP
fixup stages explicit paths, not the sidecar), so `archivePlanArtifacts` probes
each source with `git ls-files` (the injected `gitSrcUntracked` seam) and marks
the move `SourceUntracked`; `archiveAddArgs` then stages **only the history dest**
for an untracked source — adding its vanished pre-rename path would abort the
whole archive commit with `pathspec did not match` (#154). Tracked sources (issue
files, committed durable plans, recovery moves) still stage source-deletion +
history-addition exactly as before. Any unrelated dirty file keeps the refusal
path and tells the operator to clear that unrelated work before rerunning `sdlc
push --yes`.

## Makefile wrappers (transition state)

Each Make target delegates to `bin/sdlc` when built, falling back to
the original shell logic when absent:

  `make close-issue` → `sdlc close`
  `make fetch <N>`   → `sdlc fetch --github-issue N` (deprecated alias →
                       `sdlc issue new --from-github N`, #56 M2)
  `make worktree`    → `sdlc change-code --worktree=yes --no-judge --no-structural
                       --no-estimate` (post-#39, #113; preserves the make target's
                       pre-existing quick-and-dirty gate-free semantics)
  `make issue-sync`  → `sdlc claim` (renamed from `sdlc lock` in #39)
  `make push`        → `sdlc push`
  `make pull-request` → `sdlc pr`
  `make merge`       → `sdlc merge`
  `make check-<cat>` → `sdlc judge <cat>`

The fallback exists so downstream repos that vendor `Makefile.workflow`
but haven't yet run `make sdlc-build` keep working. M8 (not yet started)
deprecates the shell fallbacks and removes the scripts.

## When to add a new verb

The rule from the pensive: **when the same drift gets caught at review
twice**, promote it from a `workshop/lessons.md` entry to an `sdlc <verb>`
check. The first time → prose. The second time → code.

Do not formalize the workflow into a state machine. Add checkpoint
guards for known commit moments where drift recurs; everything between
checkpoints stays prose-driven.
