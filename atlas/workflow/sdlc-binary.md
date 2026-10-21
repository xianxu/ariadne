# sdlc binary

`sdlc` is the SDLC checkpoint binary — one Go binary at `cmd/sdlc/`
that collects ariadne's workflow checkpoint guards into a unified verb
namespace with embedded `--help` per subcommand. Agents should invoke
`sdlc` directly; Makefile targets are compatibility wrappers for
downstream repos that have not built the binary yet.

Design rationale: `docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md`.
Build issue + plan: `workshop/history/issues/000031-sdlc-checkpoint-binary.md`
(after archive) or `workshop/issues/000031-...` (during build).

## Executable identity (#461)

`sdlc --version` reports the actual executable path, embedded VCS revision and
modified bit; missing build metadata is shown as unknown rather than inferred
from a checkout. In Couch, startup warns if the executable differs from the
fleet's `:0` build. The warning is read-only and does not rebuild or switch it.
Shell aliases use that verified central owner; slot-private substrates remain
composition inputs. Refresh/relaunch captured aliases after updating the script.

Start-plan's refresh only requests a new running build when its executable's own
source checkout advanced beyond its embedded revision. A private sibling moving
alone no longer causes the generic built-before refusal. Partial refresh failures
still refuse with their actual error, and existing freshness checks remain.

## What it owns

The **checkpoints between SDLC stages** — not the stages themselves.
Stages stay prose; the binary refuses transitions that lack required
evidence. Subcommands are added incrementally when the same drift
recurs at a stage (not by formalizing the SDLC as a state machine).

## Verb surface

| Verb              | Replaces (Make target)      | Defends |
|-------------------|-----------------------------|---------|
| `close`           | `make close-issue`          | Issue close: actual + verified + atlas + plan ticked; on full-issue close auto-dispatches the one boundary review (#69, `--no-judge` to skip) |
| `actual`          | (new #68)                   | Compute an issue's focused dev-hours (in-binary active-time-v3 engine over ownership-proven sessions with retained replay inputs), split into build vs converge at the boundary reviews; `--milestone Mx` measures one milestone's increment (#234) |
| `agent-id`        | (new #367)                  | Print this session's `Agent: <agent>/<model>/<effort>` commit trailer (`cmd/sdlc/internal/agentid`): native session signals and validated transcript metadata; Qoder uses its launch binding and recorded workspace (#385), `SDLC_AGENT_ID` overrides; with no agent session the value is `human` (`agentid.Human`; legacy `sdlc` parses the same, #387). Every sdlc commit site stamps it through `agentid.StampCurrent` (`TestEveryCommitSiteStamps` refuses a new unstamped site); `close` warns about first-parent window commits without it |
| `converge-report` | (new #234)                  | **Read-only** converge factor k = converge/build per model, flow and review recipe, over ledger rows that carry a split |
| `active-time`     | (new #110; was active-time-v3.py) | Standalone CLI over the same engine — the per-segment attribution table for manual inspection; preserves the 2/3/0 loud-fail exit codes |
| `state`           | (new)                       | Workflow state inspection + drift detection; per-issue owner and claim age, claims by slot and by operator (#284, claim times from tracker history via `tracker.Repository.ClaimTimes`) |
| `fleet inventory` / `fleet policy` | (new #200) | Read-only fleet worktree evidence and prospective-path admission-policy resolution; typed JSON is the contract and human output renders the same values |
| `resolve`         | (new #144)                  | **Read-only** symbolic-ref → current path(s): the issue + its plan/review family, archive-correct + cross-repo. Locations from the `discovery:` model; grammar single-sourced as the parser. No lock (see below) |
| `open`            | (new #144)                  | Sugar over `resolve`: open the primary artifact in `$EDITOR` |
| `propagate-base`  | (new #106; precheck #109)   | Re-weave every recursive DEPENDENT of this repo (downstream counterpart to `substrateChain`): discover dependents (Makefile.workflow + substrate chain), order foundation-first, then per repo a clean-tree precheck → `make weave` + verify-complete + commit (untracking now-generated files). A dependent with a DIRTY working tree (pre-existing uncommitted work — e.g. a concurrent session) is SKIPPED untouched (never `git add -A`'d) and the run exits non-zero. `--dry-run`/`--ref`. |
| `judge`           | `make check-{dry,pure,plan,specs,lessons}` | Fresh-context LLM judge (anti-collusion) |
| `fetch`           | `make fetch N`              | **Hidden deprecated alias** for `sdlc issue new --from-github` since #56 M2 (keeps `--github-issue`) |
| `claim`           | (formerly `lock`, #39)      | Take the lock: record this workspace as owner on `issue-tracker` (#252, #283), status unchanged; a set `--issue a,b` is one all-or-nothing tracker commit (#284); refuses until the details are on main; never publishes to main |
| `unclaim`         | (new #284)                  | Release the lock: clear the owner, record the release; an open claim publishes its edits first; status unchanged |
| `start-plan`      | (new #75)                   | Planning-entry transition (with `--issue`, #252: requires a claimed card and moves design onto the issue branch at fresh main): delivers the `at-plan` architecture lens + the durable-plan pointer (`superpowers-writing-plans` → `workshop/plans/`, #72), sized against the quick-flow shell (#231) to design against |
| `change-code`     | `make worktree` (partial)   | Planning → implementation gate. First it infers the issue's flow (#231: Mx rows or a design past the shell's design limit → full; neither → quick, which runs none of the gates that follow; `--flow` pins it); the flow is recorded after the gates pass, re-derived from the issue as it is then. Then, in this order (#187 B1): structural + **plan-quality (stateful, #187)** + estimate (#113) + estimate-reconciliation + estimate-quality (#117) + branching (in-place default, `--worktree=yes`/`=ask`; #39, #51) |
| `set-status`      | (new)                       | Status-transition guards. Moved under `sdlc issue set-status` (#56 M2); **hidden deprecated flat alias** kept one cycle |
| `push`            | `make push`                 | Direct-on-main ship + the #124 instance-conformance gate (`--no-validate`) + pre-flight judges (still available; not the default close path since #51) |
| `pr`              | `make pull-request`         | PR creation with Fixes-issue body |
| `merge`           | `make merge`                | PR landing with instance/publish gates and confirmation; :0/:N archive remotely and return to unchanged rest, ordinary worktrees/dependencies retain legacy cleanup |
| `landing`         | (new #393)                  | The repository's landing line: `status` (the landing now, waiters in admission order, live grants, delay p50/p90); `grant` / `revoke` fleet-blocking priority, refused from a worker slot. `merge` waits in this line ([landing line](#landing-line-393)) |
| `local-check run` | (new #345)                  | Run checks declared in main's `.sdlc/local-checks.json` on a pushed commit, in a detached worktree under the machine-wide suite lock, and record each as the commit status `sdlc-local/<name>`; merge's CI gate accepts it for a CI job that never got a runner and requires it for `required` checks ([ci-merge-check](ci-merge-check.md#recorded-local-checks-345)) |
| `local-check cancel` | (new #400)                  | Select one owned run with `--run ID` (printed at run start), all owned runs of `--name N`, or all this checkout's checks; ID/name are exclusive and foreign/invalid/inactive IDs refuse. Same-check/different-head overlap prints a notice and exact existing-run cancellation command, preserving both FIFO places until explicit cancellation. Each run holds a flock'd lease in `<git common dir>/sdlc/local-check-runs/` naming its checkout, published already locked; cancel SIGTERMs the run (after matching the pid's kernel start time), waits for it to unwind, and sweeps leases of dead runs. The run also names its checkout in the suite lock's held-by record (`WF_TEST_LOCK_OWNER`) |
| `local-check republish` | (new #431)                  | Post a completed check's saved attempt (test-stats store) as its commit status without rerunning: refuses another commit/check/repo, a baseline, an incomplete record or a log whose sha256 no longer matches. A run's status post retries (15/45/90 s) and its final failure names this verb; merge's local-check runner posts a saved verified pass on the exact head before it would rerun |
| `test-stats`      | (new #376, #392, #402)      | Local-fleet suite wall, queue, slow-unit and flake summary (`--since`, `--repo`, `--json`). `snapshot` joins exact attempts to confirmed landing cohorts, separates physical weekly attempts, and atomically publishes a versioned weekly JSON artifact in brain. Schema v2 adds inventory-backed resource sizes and machine gauges with actual `observed_at`; point-resource deltas require observations 6–8 days apart, while suite deltas use exact weekly windows. Device `observed` validity is separate from comparison state, so a valid reading remains usable after an unknown delta. Unknown evidence leaves the ratio null; legacy logs are partial. |
| `flake`           | (new #394)                  | Known-flake registry verbs. A test is a known flake while an open issue's **card** lists it in `flake_tests` (kind `test-list`, mirrored into details). `flake register --test PKG.TestX [--issue N] [--evidence TEXT]` writes it only with same-head fail→pass evidence from this machine's test-stats store (`internal/flakes.SameHeadFailPass`, the rule test-stats' "confirmed" flakes also use), or `--evidence` for a CI run; without `--issue` it files an unclaimed flake issue. It is a card setter, with no owner check. `flake list` is the TL's local daily view: the registry, unregistered same-head flakes, and unexplained failures older than a day. #458 filters already-fixed same-head rows using one terminal issue's card `flake_tests`, landing merge and Git ancestry (never issue Markdown); since #479 each row is an uncovered failing leaf (`flakes.Uncovered`), matched to a terminal issue by the same `Covers` rule, nearest name first (`fixedCandidates`); `flake list -v` reveals them, while post-fix recurrences stay visible. Ambiguous or unreadable evidence cannot hide a row. CI is shown as `not covered` until #451. The registry is built from a Fresh tracker read (`loadFlakeRegistry`), so a stale or unreadable card makes the rule unavailable, never an empty registry. One owner per test is enforced where the registry is read (`BuildRegistry` refuses a test two open issues name), not only at registration. **The gate (M2):** `runLocalCheck` classifies a failed suite attempt (`knownFlakeStep`, `internal/flakes.Classify`). When only registered tests failed, it reruns them once through the shared `runUnderSlotBatch` (weight 1, its own lease, the #383 slot release), saved as a v2 attempt of check `suite-rerun` (`rerun_of` the failed one; never read as a suite run), and posts "pass · known flake #N rerun". `known_flakes` (`rerun`\|`accept`\|`off`) on the declared suite is the operator's policy. merge's `diagnoseSuiteFailure` labels a shared failure as a persistent main failure. **Schema note:** suite attempts are written as version 2; an sdlc built before #394 refuses them as an unsupported version, so the fleet runs ariadne:0's rebuilt binaries once #394 lands. **Subtests (#466):** a name may carry a subtest path (`flakes.ParseName`: the leftmost dot that yields an import path and a Test function, then validated segments). A failing top-level test is judged by its failing leaf subtests (`Leaves`), each covered (`Covers`, `Registry.Cover`) by the leaf or an ancestor, never a sibling; the gate, the success guard (`Unexplained`) and `flake list` share that one rule (`flakes.Explained`), with attempts carrying the subtests read from their saved logs (`withLogSubtests`; an unreadable log fails closed). `SuiteLogFacts.Subtests` is never persisted, and records keep top-level failures. The rerun is the owning top-level tests (`RerunTests`), each entry's package-bound JSON PASS event required; overlapping names on two issues are ambiguous. `flake narrow` and `flake unregister` are owner-only compare-and-swap edits under the details-publish rule (`requireOwnedToEdit`). Rollout: an sdlc built before #466 refuses a subtest name in `flake_tests` (its known-flake rule is then off) and an attempt whose `known_flakes` names one (its success guard fails closed), so rebuild :0 before registering one. |
| `milestone-close` | `make close-issue MILESTONE=Mx` | Milestone close + auto-dispatched boundary review (the one reviewer, per-milestone window; #69). THE milestone-close path — `close` refuses `--milestone` (#146); `--no-judge` here is the labeled skip-review escape. |
| `issue new`       | (new; xx-issues skill prose)| File on origin: the next ID's card on `issue-tracker` and the details on main, writing no checkout; a numbered slot claims what it files unless `--no-claim` (#252, #308; `--from-github N` seeds from GitHub; `--details FILE\|-` gives the body as `## <Section>` blocks, parsed by `issue.ParseDetails` before any ID is reserved, #371) |
| `issue move-detail` | (new #252; hidden #308)   | Deprecated alias of `issue publish --issue N`: finishes a pre-#308 creation by publishing its local details to main; net-zero removal on the source |
| `issue recovery`  | (new #252)                  | `list` / `reconcile --issue N` interrupted tracker operations, from the checkout that owns them |
| `issue set-title` / `set-estimate` / `set-github` | (new #252) | Card-field setters (compare-and-swap on the card) |
| `issue set-area` / `set-severity` | (new #348) | Classification setters: areas must be declared (or ancestors of declared values) in main's `.sdlc/vocabulary.json`; severity is the shared enum ([vocabulary](vocabulary.md#repo-defined-enums-348)) |
| `issue sync`      | `make issue-sync ISSUE=N`   | Legacy repositories: commit one issue locally under `#N: issue-sync: <what>`. Retired in tracker repositories (#284): a pointer at `git commit` and `issue publish` |
| `issue publish`   | (new #244; tracker mode #284) | Tracker repositories: `--issue a,b` publishes details to main (later edits, or finishing a pre-#308 creation by the owner (anyone, while unclaimed, #274) in one narrow commit, never over a moved main). Legacy repositories: `--commit SHA` applies one selected documentation commit to fresh remote main |
| `trunk publish --path P` | (new #240) | Publish one dirty non-issue workshop Markdown file directly to main from an issue branch, with a CAS and path-level conflict check; the local file remains untouched |
| `issue set-status`| ← flat `set-status`         | Status-transition guards (relocated #56 M2) |
| `issue list`      | (new)                       | List issues (ID/status/title), sorted by ID; `--status` filters; reuses `listIssues` |
| `issue show`      | (new)                       | Issue frontmatter + section headers, no bodies |
| `issue lint-ids`  | (new #213)                  | Refuse issue ids reused across a range, and report ids already duplicated within one tree. Read-only; the logic CI's `40-duplicate-issue-id.sh` shells to |
| `issue validate`  | (new #124)                  | Validate issue file(s) against `#Issue` — frontmatter cue-vet (via `vocabulary validate-instance`) + section presence; multi-target (#133): `<file>...` / `--issue N[,N...]` / `--all` (mutually exclusive). The on-demand surface of the instance-conformance loop |
| `project new/list/show/validate` | (new #180 M3) | Author and inspect project records. Scaffold sections/status and discovery derive from `#Project`; validation shells to the noun-generic vocabulary validator for frontmatter and checks the scaffold sections itself (#329; the changed-instance gate stays frontmatter-only) |
| `project set-status` | (new #180 M3; #329) | Enforce the project lifecycle and its ordered named guards from `project.cue` (`trains-present`, `mvp-landed` through the shared card lookup); unknown guards fail closed, and `closed` remains owned by `project close` |
| `project status/retro` | (new #180 M4; #373) | Derive progress, dependency frontier, remaining effort, thread components, MVP landed/open and the done_when from live issue records; append dated re-forecast checkpoints without overwriting the baseline |
| `project close` | (new #180 M4; #329; #373) | Require the modeled closing→closed (or executing/closing→dropped) edge, a retro, `--verified` evidence and a landed MVP (`project.MVPProgress`, shared with the `mvp-landed` guard), then archive through `ArchiveSubdir(..., ArchiveProjects)`; a linked record (ops) is closed in place and its link moves. `--dry-run` reports every gate |
| `project tick` | (new #373) | Tick a landed issue's rows (+ `(PR #N)`, `project.TickLandedIssue`) in every live project citing it, each by one CAS commit on its repo's origin main (`publishProjectEdit`, shared with close's peer publish). `sdlc merge` runs it after landing (`tickOnLanding`, warn-only); TL-owned projects only with `landing_tick: true` (`TLLandingTickDefault`), else merge prints the command |
| `project forecast` | (new #329) | Read-only: the trains' critical path as a range of focused hours, from each car's card and the calibration ledger (below) |

### Issue reservation, local durability and publication (#244)

These are separate operations in every checkout, including primary `:0`, numbered
slots and independent dependency clones. See [issue publication](issue-sync.md)
for the operator workflow and recovery contract.

| Operation | Publication unit |
|---|---|
| `claim --issue N` | (#252, #283) The card's owner (the lock) on `issue-tracker`, status unchanged; nothing on main |
| `issue new` | (#252, #308) A new card on `issue-tracker` at `max(id)+1`, then the details on main (own commit); from `repo:N`, a claim commit |
| `change-code` | (#252) Nothing: the design is a local commit on the issue branch |
| `unclaim --issue N` | (#284) The cleared owner and its release record on `issue-tracker`; an open claim's edits first go to main as `issue publish` |
| `issue publish --issue N` | (#284) Finishing a pre-#308 creation: the transfer (`move-detail`). Republish: one main commit with the issues' edited details (never another workspace's, #274), `#a,#b: issue: publish details` |
| `issue move-detail` | (#252) One main-native commit adding the initial details, plus the card's handoff record |
| `issue sync --issue N` | Legacy repositories: one local issue-file commit; no network operation by default |
| `issue sync --push` | Legacy repositories: only the narrow issue commit created by that invocation |
| `issue publish --commit SHA` | Legacy repositories: the selected non-merge commit's entire eligible Markdown change set |
| `trunk publish --path P` | One uncommitted prose file under `workshop/parley/`, `pensive/`, `targets/`, `projects/`, or `workshop/lessons.md` to main; never an issue-branch commit |

For product/discussion work in a tracker repository, `issue publish --issue N`
owns issue details (and issue plans remain on their issue branch). `trunk
publish --path P` is the explicit home for a non-issue workshop note edited
mid-issue. It reads one regular, unstaged Markdown file up to 1 MiB; the path
must be unchanged in the issue branch since its merge base with main. Main's
copy must still equal that base, unless it already has the selected bytes. The
CAS rechecks main on every attempt; a path-level conflict refuses without
touching local work. Publication retains the selected local file. For a
tracked path, it prints an exact literal-path `git restore` command to run
after merging main into the branch, or before a merge that would otherwise
conflict. An untracked path instead gets a targeted `git clean -f` command to
run before the merge; Git restore cannot remove untracked files. If the local
file changed during publication, it warns to preserve and reconcile that
newer edit. A retry after an uncertain push uses the surviving local bytes to
recognize an already-applied publication.
Atlas changes travel with the code they map, and code/base-layer settings
travel through the normal reviewed issue branch; they are outside this command's
allowlist.

`runClaim` owns remote status reservation. `syncIssuesToMain` dispatches local
checkpointing, creation reservation, or publication of a newly created issue
commit. `syncInPlace` only commits the selected local issue; `syncViaTrunk` is
reserved for (pre-#252) creation. Since #252, `change-code` commits the
accepted design on the issue branch and publishes nothing; `issue sync --push`
asks for an explicit source SHA when there is no new commit.

Documentation publication applies source-parent → source to fresh `origin/main`
with Git three-way merge semantics. It preserves unrelated remote edits and all
unselected local commits. Mixed code/document commits and nonregular files
refuse as a whole. A conflict names its paths; missing prerequisites must be
published explicitly in order. There is no local-main whole-branch shortcut.

Publication preserves the caller's checked-out branch, index and working files;
fetch may update remote-tracking refs. The resulting commit retains the selected
message and adds `Source-Commit:` provenance. Repeating an already-published
source does not replay it after a later edit or revert. A byte-identical merge
reports no change without creating a remote commit.

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
Local issue sync, archive and migration commits follow this rule.
`archiveCommitArgs` derives its path
list *from* `archiveAddArgs` so the two cannot drift, and refuses an empty move
list — `git commit -m … --` with no paths is read as NO pathspec and commits the
whole index, which is the helper's own failure mode on its degenerate input.

Remote publication constructs Git objects independently of the caller's index.
`syncInPlace`, `push`'s archive and `migrate`'s source side still commit locally;
`issuesync_test.go` protects these paths against unrelated staged work. `migrate --no-commit` prints the pathspec'd
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
entry point is not in-process drivable (`runPush` / `runMerge` / `runMigrate`
die through gates a unit test cannot satisfy), the source guard asserts the edge
instead. `runChangeCode` was the first such verb until #191 made `exitWithCode` a
swappable var like `die` (`expectExit` in `die_test.go`); its gate loop is now
driven in-process, and its source guard stays as a cheap net. Weaker than driving the verb, strictly
stronger than a call site that can be deleted with the suite still green.

**Behavior changes worth knowing about**, both introduced here:

- A pathspec'd commit is a *partial* commit, and git refuses one outright while
  `MERGE_HEAD` is set (`fatal: cannot do a partial commit during a merge`). A
  sync or archive run mid-merge now fails loudly where the bare commit would
  have folded the merge in silently. Better behavior, but finish or abort the
  merge first.
- `issue sync`'s no-push default is a property of the **verb**, not of the
  repository. A later explicit documentation publication or normal PR/ship flow
  can publish that commit. `claim` and `issue new` do not sweep it into their
  reservation transactions.

### Issue-id allocation reads the trunk, not the checkout (#213)

`issue.NextID` used to scan local directories only, so a branch cut before an
issue landed on `main` had a `workshop/issues/` that never contained it — and
`sdlc issue new` there reallocated a **published** id. Not a race: it reproduces
days later, because the branch simply lacks the newer files. The repo transaction
lock does not help and never could; the colliding allocations shared one
`.git/sdlc.lock` and were serialized. A lock orders access to one view, it cannot
reconcile two disjoint views.

Nothing downstream caught it, because the two files get different slugs and are
therefore different **paths** — git merges both cleanly. Eight collisions across
two repos existed when this was written, the oldest from May 2026.

    resolveIDDirs(...)         WHERE ids live — resolved once, Rel + Abs
    issue.PathsByID(listing)   THE parser — id → every path claiming it
    refIDSpace(ref, dirs, r)   THE reader — one ref's id space
    LocalPathsByID(dirs.Abs)   this checkout, same shape
    issue.NextID(idSets...)    PURE — the decision, no IO
                               → local ∪ published, so unpushed issues count

One parser and one reader, deliberately (#213 BR-23). There were three of each —
one yielding ids for allocation, one yielding paths for the merge gate, one
yielding duplicates for the lint verb — and each carried its own idea of what a
failed read meant. That is why the *same* silent-degradation defect had to be
found four separate times across four review rounds.

**Every degraded read is announced**, which is the rule the whole defect family
reduces to: a read feeding the id space must be verified **fresh, complete and
on-target**, or said out loud. A non-answer is not an empty answer, and unioning
zero ids in from a blind read is how each instance silently handed out a
published id. The four:

| | the blind read | announced as |
| --- | --- | --- |
| BR-7 | a per-directory `ls-tree` failed → half the trunk | hard error |
| BR-15 | dirs outside the repo → a stranger's id space | refuse + local fallback |
| BR-23 | fetch failed but the ref still resolved → stale trunk | `POSSIBLY STALE` warning |
| BR-25 | dirs joined onto the **cwd** → a directory that does not exist | fixed at the root; empty-from-every-source still warns |

The same rule reaches the sites that only *report*: `sdlc merge`'s gate warns
when it could not refresh the trunk (it used to pass with a confident `[ok]`
over exactly the window it exists to cover), `claim`'s publish path
errors instead of reporting success over a read it could not make, and
— the sharpest —
`sdlc issue lint-ids` **exits 2 on a degraded read**. That verb is what CI shells
to, and every read failure used to warn and exit 0: a GREEN required status
check on a check that never looked. Green is the one answer a check that did not
run must never give, because it is indistinguishable from "looked and found
nothing" — which is this issue's entire subject.

    0  clean      1  collisions introduced      2  COULD NOT RUN

BR-25 is the sharpest: `workshop/issues` names a place in the **repository**, not
a place relative to wherever the operator is standing. Joined onto the cwd it
yielded `docs/sub/workshop/issues` from a subdirectory — still inside the repo,
so the containment guard passed — and both `ls-tree` and `os.ReadDir` truthfully
answered "nothing". `sdlc issue new` then allocated `000001`, **wrote the file
into the subdirectory**, and pushed it: a live issue where no gate looks, whose
natural repair (`git mv` into `workshop/issues/`) manufactures exactly the
collision this issue exists to prevent. Reading and writing now share the one
resolution, so they cannot disagree about where ids live.

Offline falls back to the local scan but **loudly**: a silent fallback recreates
the bug. `resolveIDDirs` refuses issue dirs outside the current repo — without it
a `--issues-dir` elsewhere would read *this* repo's trunk as its own id space.

**Three layers, because the first two are not enforcement:**

| layer | where | strength |
| --- | --- | --- |
| allocation | `issueids.go` | prevents new collisions on a fetched checkout |
| local gate | `sdlc merge` step 4.6 | operator feedback — bypassed by a GitHub-UI merge, bare `gh pr merge`, `--no-validate`, or an unpulled actor |
| CI check | `merge-checks.d/40-duplicate-issue-id.sh` | server-side, can be a required status check; **symlinked** so it reaches derivatives, and resolves sdlc via `construct/dev-aliases.sh` so a repo without `cmd/sdlc` still runs it |

The CI check resolves the **trunk tip itself** rather than using the base the
runner hands it. That is not a detail: the runner contract supplies
`merge-base(base, head)`, and this collision's defining shape is *branch cut
first, colliding id published after* — so the merge-base predates the published
file and the id looks new on both sides. Measured on the exact reproduction:
merge-base reported "no reused ids"; the trunk tip refused.

**A branch-vs-trunk comparison structurally cannot see a collision already
merged** — both files are on the trunk, the trees agree, nothing is reported.
`issue.DuplicatesIn` asks the other question ("does one tree contradict
itself") and is the only way the existing damage is visible. It REPORTS rather
than refuses: those collisions predate the check, renumbering is operator work,
and blocking every merge until it is done would be worse than the bug.

**Flat verbs vs the `issue` group (#56).** The flat verbs guard workflow
*transitions* (close, claim, change-code, pr, merge, …). `sdlc issue *` is the
CRUD/authoring surface for the issue *record* — the noun-grouped home for
`new` (and, post-#56-M2, `set-status`/`list`/`show`). The canonical issue-file
template lives in one place: the `Render` function in `internal/issue/scaffold.go`,
documented in prose by `sdlc issue --help`.


## Durable PR landing

`pr.go` and `merge.go` resolve workspace identity before legacy primary/origin
lookups. For :0 and :N, `landing.go` uses the resting branch's single named
remote tracking main; effective fetch/push destinations must identify the same
supported GitHub repository. `sdlc pr` builds its window from fetched configured
main and publishes the issue branch there. `ghlanding.go` binds a same-repository
PR to exact head/base identity, submits an expected-head server merge, and checks
MERGED plus reachable integration. Queue admission is not integration; already
merged squash/rebase PRs use the original PR head and integration evidence.

Issue-only main movement (#472) uses one path policy in `mainmovement.go`:
Markdown under `workshop/issues/`, `workshop/issue-cards/`, `workshop/plans/` and
`workshop/history/`. Tracker configuration (`workshop/issue-tracker.json`) does not qualify.
Every novel commit and merge-parent difference is classified, including rename
sources; code added then reverted, mixed changes and unreadable history remain
ordinary. The existing integration state machine charges only ordinary moves to
its three-attempt budget, including interrupted replay and reset.

After exact-head gates, `ghreadiness.go` reads GitHub's effective merge status and
live `baseRef.target.oid` (the listing's `baseRefOid` can be historical). Fresh
CLEAN evidence bound to this PR/head/branches and re-fetched main permits landing
unchanged despite issue-only movement. Destination instance, duplicate-ID and
publish gates are rechecked. Strict, unknown, unavailable or stale-base evidence
uses ordinary integration and exact-head CI without charging an issue-only move.
No workflows or statuses are rewritten. Existing docs-only suite reuse is broader
than this conservative direct-merge policy. Recovery and uncertain merge outcomes
retain the existing branch-preserving rules.

PR candidate selection (`selectLandingCandidate`, #418) retains OPEN PRs even
when their heads are stale, so two OPEN records refuse and name both. Completed
older heads are history once the branch head is known. The same rule governs
strict selection, interrupted integration and PR-head registration polling;
closed PRs remain ineligible. A sole stale OPEN PR retains the named push remedy.
Identity validation still applies before history is filtered.

`landingarchive.go` owns the scoped remote-main archive through
`gitx.TrunkFile.UpdateMany`: PR-owned close anchors select completed records
(in a tracked repository, card completion bindings whose evidence commit is in
the PR, #252 — their details archive with the done card mirrored in, #275), and one conditional commit
moves issue/plan/review artifacts and their side folders (#424). Retry verifies
reachable provenance and the complete archived generation before cleanup.
A done-card sweep can archive a car without per-PR trailers (#494, #512).
Every archive check, including fresh already-merged recovery and the publisher's
refreshed main view, can locate that archive through bounded destination-path
history. The shared transition verifier must confirm its exact tracked completion,
all artifact moves and unchanged current archive. This proof uses the same pinned
main snapshot and selected completion as the ordinary check; it needs no transient
train notification. Missing, ambiguous, changed or mismatched proof still refuses
cleanup. Legacy untracked issues continue to require per-PR archive provenance.
Landing returns the selected checkout to unchanged `main` or `main-slotN`, then
compare-and-deletes only the proven local issue ref and removes its configuration.
Tracked dirt (including tracker edits), active Git operations, conflicting
occupancy or uncertain evidence refuse. Noncolliding untracked/ignored files,
sibling workspaces and the enclosing dependency environment remain intact.

`sdlc merge --branch NAME --yes` resumes from the issue branch or its resting
checkout. At rest only already merged work can resume; an absent ref still needs
integration/archive proof before leftover branch configuration is cleaned up.
Recovery uses Git/GitHub evidence, with no journal. Remote branch deletion is not
added to this path. An intentionally old local resting tracker may still show a
record archived remotely; refresh is explicit, never part of durable landing.

Ordinary feature worktrees and private dependency clones retain legacy cleanup,
including refresh of their own main and removal of ordinary feature worktrees.
Ship a dependency such as Ariadne first through its normal SDLC flow, verify Pair
against it, then land Pair separately. There is no recursive publication.
See [workspace branching and landing](workspace-branching.md) for the procedure;
`landing_test.go` uses real Git plus a stateful GitHub fake for preservation and
interruption/retry coverage.

### Landing line (#393)

The pre-close suite verifies the pushed issue branch; it does not reserve a
landing turn. After that pass, close the tested branch and enter `sdlc merge`
on SHIP. Do not manually chase each newer main with another pre-close suite.
The existing noon freshness gate and verification after branch code fixes
still apply. Merge integrates current main after admission and runs the
required suite on that integrated code when the earlier record cannot cover
it. An unchanged branch record is not proof of newly integrated code.

The hub lets this turn order integration instead of holding main by hand or
asking workers to integrate newer main before close (#425). Ordinary docs-only
main publication can continue: the existing first-parent record search reuses
a pass only when the cumulative difference has no code surface, including
embedded help. Code, tests and configuration still require a new suite. The
line coordinates SDLC merges; independent Git pushes remain subject to main
revalidation and bounded integration retries.

Concurrent merges used to integrate against the same main. Each landing then forced
the others to integrate again and rerun a full suite. The landing line orders a
repository's landings, so one merge at a time integrates and lands.
`cmd/sdlc/internal/landingline` is a pure core (`order.go`) inside a thin IO shell
(`line.go`, `log.go`). The line directory is `<git-common-dir>/sdlc/landing/`,
which every slot of a repository shares, beside #370's `sdlc.lock`. The repo lock
is mkdir-based; the line uses flock(2) through `internal/flock`, which it shares
with reviewbox.

- **Tickets and the head.** A merge holds a flock on its own ticket file while it
  waits. It holds `head.lock` once admitted, and leaves the ticket set at that
  point. Every flock dies with its process: the next scan prunes a dead waiter's
  ticket, and a dead head frees the line. Nothing is left that a crash can make
  inconsistent, and there is no daemon. Scans and admission run under the
  `line.lock` mutex. A ticket appears already locked (it is written under a
  temporary name and renamed into place), so a scan never takes a new ticket for
  a dead one.
- **Order.** `Order` and `NextHead` put fleet-blocking tickets first, then normal
  ones, FIFO by ticket number within each class. Waiters poll every 2 s. A
  running head is never preempted: priority decides only who is admitted next.
  `GrantHold` holds normal admissions while any valid grant exists, even before
  its issue queues or between two granted issues in one checkout (#470).
  Granted tickets keep FIFO among themselves. Status and waiter output share
  the hold explanation: issue, reason, expiry and revoke command in this repo.
  Cancellation does not revoke a grant; expiry/revoke releases its hold.
- **Grants.** `grants.jsonl` is the only source of priority. `ValidGrants` keeps
  the latest entry per issue when it is unrevoked and unexpired (at most 24 h),
  has a reason, and records a grantor that is the operator's shell or a :0 slot.
  That check runs on read as well as on write. Trust is cooperative: it guards
  against agents' mistakes on one machine, not against a process that forges the
  grantor.
- **Events and retention.** `events.jsonl` records enqueue, admit (with the
  tickets it displaced), landed and refused. `Delays` gives the
  enqueue-to-landed samples for the pilot's p90, through test-stats'
  nearest-rank helper. Both logs are appended under the mutex. A daily sweep
  applies the pure rules in `retention.go` and counts the malformed lines it
  drops.
  - Events stay while they are dated within the last 30 days and not in the
    future.
  - A grant entry stays while it and its issue's latest entry are fresh by
    their grant time.
  - Invariant (tested over random logs): a sweep never changes which grants are
    valid.
  - A revoke outlives its grant, because it is dated when it is made.
  - A future-dated grant confers nothing.
- **Relation to the suite line (#452).** `scripts/test-lock.py` reads the same
  grants through the hidden, read-only `sdlc landing suite-priority` v1 protocol.
  Go's `ValidGrants` remains the authority. Python retains its own machine-wide
  slot locks and FIFO tickets; no grant or landing turn is created by a suite.

**In merge** (`landingqueue.go`): `runDurableMerge` confirms first, so a prompt
never holds the line. It then takes a turn before it observes the branch head,
the PRs or the integration state, and holds the turn until it returns. An
admitted merge therefore never acts on an observation from before a possibly
long wait. A duplicate invocation that waited behind the original resumes as
already merged (logged `resumed`, which the delays skip), and a fix pushed
while waiting is what lands. Rule: a merge without a turn acts only on an
already-merged PR. It skips the line only when the first listing shows a merged
PR and no open one; an empty listing queues, since it may lag a new PR. The PR
it selects is checked against the rule, so a lagging listing that hid an open
PR refuses, having changed nothing. The turn logs enqueue, admit, and landed, refused or resumed; landing
ends the issue's grant. A log write failure warns once on stderr. The landing's own recovery is unchanged, and the line
adds ordering only. This lock-shaped line replaced #390's eight-state durable
queue (#404 lock-first, TL-approved).

**`sdlc landing`** (`landingcmd.go`, help in `helptext/landing.md`): `status`
shows the landing now, the waiters in admission order, live grants and the delay
p50/p90. `grant` and `revoke` take `--repo DIR` and are refused from a worker
slot.

**Landing train (#449).** `.sdlc/landing.json` on main
(`train_interval_minutes`, `train_max_cars`; local-gates guarded) turns the line
into trains. Absent, or an interval of 0, keeps the serial line above. With
trains on:
- A waiting merge vouches its readiness in `ready/<seq>` beside its ticket
  (`landingcar.go`): its own publish and CI gates, read-only, for its exact
  pushed head. Each `carReadiness` pass gets a nested issue-records scope
  (#499): it observes external close/status updates on the next pass while
  sharing one coherent tracker generation across that pass's gates, without
  invalidating the parent command or another pass. A carried completion binding
  remains a publish obligation after its card is reopened (#502). The shared
  publish-only guard (`refuseReopenedCompletions`, `trackercompletion.go`) checks
  every carried binding before review-anchor selection, in ordinary publish and
  immutable landing. It refuses with the issue, current status and exact
  `sdlc close --issue N --verified '<evidence>'` action, followed by push/retry.
  Binding repository/evidence/current-token ownership survives rebase and an
  independently published details copy; unrelated working cards and branches
  carrying no close stay valid. Completion, settlement and archive retain their
  existing status eligibility. Transferred-details ownership
  checks and completion selection use the parsed `Records.CardSnapshot`;
  alternate details directories reuse those cards through `Records.WithDetails`.
  The publication checkout selects that card authority, even if a configured
  details directory points into another repository. Where main declares
  `.sdlc/affected.json`, `carCIGateFacts`
  makes a recorded `affected` run stand in for `suite` (`ciInputs.StandIn`,
  #473 D3): the train's prefix suite is the one full run. Merge without a
  train sets no stand-in. A passing full-suite record already satisfies the
  car directly; the affected stand-in is an alternative. The missing-record
  message names affected or its full-suite fallback and explains conditionally
  to wait if affected fell back (#513), without asserting a queued run exists.
- Admission waits for the departure (the last `train` event's `next_due`)
  and a current vouch. A grant departs at once, alone.
- Candidate and composed membership have distinct banners (#521); the terminal
  summary counts retained cars. Held outcomes print their reason and exact branch
  recovery in the owning runner (`reportTrainHeld`). If the conductor’s own car
  is held, the Publish boundary stops it after all composition holds are delivered;
  existing Aborted cleanup releases survivors to reform. No sibling suite runs
  under the excluded checkout’s lifetime lease. A reformed suite belongs to its
  new conductor’s checkout, so foreign local-check cancellation leaves it alive.
- The pure core is in `internal/landingline`: `train.go` (schedule, readiness
  codec, boarding), `trainstate.go` (the running train's state machine) and
  `bisect.go` (splitting a failing train).

`sdlc landing status` shows the train setting and each waiter's vouch.

**Departure slots, N-or-Y, no clock reset (#475).**
- **Due check:** `DepartureDue(now, schedule, events, Waiters)` (`internal/landingline/train.go`) decides when a train leaves. It leaves early when `train_max_cars` cars are ready, or when the oldest has waited `train_max_wait_minutes`, measured from its first ready vouch (`Readiness.Since`, kept across re-vouches at one head). Otherwise it leaves when the next departure comes.
- **The next departure** (`NextDue`) is the grid's next slot after the last train when the repository has a machine-plan offset (`Slots`, `internal/landingline/slots.go`: epoch + offset + k·interval, never shifted by an early departure), and the interval after it otherwise.
- **Lone and granted landings** are not departures: their `alone` record carries no `NextDue`, unless they left cars ready, which can only bring the next train forward.
- **Admission and status:** each admission pass gets a `Waiters` summary (`ReadyWaiters`). Train events record `departed`. `sdlc landing status` prints the same decision with its reason.
- **Machine plan (M2):** one file per machine, `landing-plan.json` beside the suite lock, gives each repository its grid offset.
  - It is decoded and packed in `internal/landingline/plan.go`. `MachinePlan` keeps unknown members; `PackOffsets` places the heaviest p90 × weight reservation first, within the machine's suite slots.
  - It is read by `landingScheduleSource`, under the schedule's 30 s cache (`landingconfig.go`, `machinePlanPath` / `readMachinePlan`).
  - It is written by `sdlc landing plan [--write]` (`landingplan.go`). Its costs come from `buildTestStatsReport` and `fleetRepoPaths` (split out of test-stats) and each repository's declared suite weight on main.

**The conductor** (`landingtrain.go`) is the admitted head when other cars board:
- **Boarding:** cars are boarded only after their remote tip and open PR are re-read at the vouched head.
- **Build:** it builds `landing-train/<id>` in a detached worktree from fetched main, then each car's exact head. The instance and duplicate-ID gates run per prefix, and a conflict holds that car.
- **One PR, one suite:** it opens one train PR and runs one recorded suite (#394's flake step inside it) and CI on it.
- **Suite ownership:** the detached train is the content tree; its original conductor checkout remains the caller through `localRunDeps.WorkflowAnchor` and `callerRoot()`. Saved attempts (including baselines and prefix retries), queue operations, cancellation leases and capacity-holder diagnostics use that caller. Relative Git common directories resolve there too. Overlapping issues' actuals can then exclude another slot's train wait without inventing provenance (#492).
- **Merge:** it merges once, pinned to the head. GitHub marks each car's PR merged by reachability.
- **Outcomes:** every car is told through `outcomes/<seq>` (boarded, landed, held, refused, released). A landed car finishes today's resume cleanup.
- **Landed commit:** each car proves its landing with `landedCommit` (its mergeCommit or its own head on main; `LandedOID` feeds the archive and the card).
- **Locking:** ref writes take the repo transaction lock.
- **Main moves (#511):** issue-record-only movement uses the shared strict history classifier and fresh same-PR/head/base CLEAN readiness. The train retains its exact checked head and CI only while policy/car/log-bound landing-check proof remains valid and the final destination passes instance/duplicate-ID gates under the merge transaction. Other movement or unavailable/stale readiness reintegrates main; generic docs-only reintegration can retain the suite but still needs CI on the new head. No statuses are copied.
- **Lost reply:** a merge whose reply was lost is read back from GitHub.
- **Branch cleanup:** a landed train's branch is deleted. Every train-branch delete goes through one helper, under the repo lock and leased on the tip last pushed or observed.
- **Leftovers:** the next conductor closes an open `landing-train/*` PR left by a dead conductor, and deletes train branches with no open PR.
- **A car with no word:** a waiting car that finds its own PR merged at its head (its conductor died after the merge, or the outcome write failed) ends its wait and finishes as a resume.
- **Uncertain merge:** a conductor that stops while its merge's outcome is unknown records `merge-unconfirmed`, never landed or abandoned. Its cars go back to wait, and each one reads its own PR.
- **Splits:** an unexplained suite failure is bisected over the prefixes (at most ⌈log₂ n⌉ more suites). The culprit is held, the clean prefix lands, and the cars behind it are released, so the next train departs at once. A known flake's #394 rerun is a pass. Main's baseline failing the same way refuses the train.
- **Refusals:** red CI on the train PR refuses the whole train (no CI bisect). A fleet-blocking grant, or a PR with the local-gates label, lands alone.
- **Report:** `sdlc landing status` lists the last trains from the `train` events: result, cars, train PR, split suites and culprit.
- **The train's state machine as built** (`internal/landingline/trainstate.go`; `TestPlanNamesEveryTrainEventAndEffect` keeps the events and effects in step with the code):
  - Events: `Boarded`, `Built`, `SuiteDone`, `CIDone`, `MainMoved`, `Reintegrated`, `Merged`, `MergeUnknown`, `Probed`, `Aborted`.
  - Effects: `Build`, `Publish`, `RunSuite`, `WaitCI`, `Merge`, `Probe`, `Reintegrate`, `Hold`, `Release`, `Refuse`, `LandCars`, `ClosePR`.
  - Phases: forming, integrating, testing, bisecting, reintegrating, landing, merging, and the terminal landed, abandoned, refused and merge-unconfirmed.

With no other car, the head lands alone, as before, and logs the departure. Plan: `workshop/history/plans/000449-landing-train-plan.md`.

## The quick flow at close (#231)

`internal/flow` is the pure core: the one-line `flow:` record codec, `Decide`
(change-code's inference), the hard-shell limits (`MaxAddedLines` and
`MaxDesignLines`, rendered with the Mx rule by `ShellSummary` into help via
`{{QUICK_SHELL}}`), `Measure`/`Crossings`/`DesignLines`, the contract hashes and
the Done-when checks. change-code and close size the shell through the same
`Measure`/`Crossings` — change-code before any diff exists, on the design
(`## Spec` + `## Plan` + the durable plan, in lines) and the Mx rows. On a quick
issue, `computeClose` (`closeflow.go`) measures the window the atlas gate and the
review already use — added lines in code files (`churn.IsCodeFile`) via numstat,
the design as it stands, and the Plan's Mx rows. Outside the shell it composes an upgrade to
full/inferred (written at finalize); inside, the dispatch runs the
`small-diff-review` recipe. The shell measures size only: neither how many
files a change spreads across nor where it lands counts (the file-count limit
and the shared-surface declaration were removed after the first trial reading,
2026-09-18 — for a diff this small, tests and the one close review are the
guard). Fixes made after the verdict ride into the close commit unmeasured, so
the publish gate (`publishgate.go`, `quickGrewPastReview`) re-measures a quick
issue's final diff over close's own window (`boundaryWindowBase` → HEAD):
past `flow.MaxAddedLinesAfterReview` (twice the line limit) it refuses and
sends the issue back to close, which upgrades it and runs the full review. The
gate itself stays LLM-free.

**Done-when freshness (F3, #183).** `flow.DoneWhenFresh` flags a contract
(`## Spec`/`## Revisions`) that moved since change-code while `## Done when` did
not. That is either a reframe or a review-fix Revisions entry, and only a reader of
the diff can tell which. So `doneWhenDecision` (`closefixed.go`) defers to the
close review whenever one runs. The prompt carries the note through
`ReviewNotes`, the reviewer states `Done-when: current|stale`, and the close log
line records `done-when: … judged by the close review` before its verdict annotation.
Under `--no-judge` the deterministic refusal stands (`--no-done-when-fresh` waives
it). A review that did not run never finalizes, so it cannot pass the check.

## Fleet inventory and policy (`sdlc fleet`, #200)

The read-only `sdlc fleet inventory` command owns assembly of canonical
repository/worktree rows across the shared filtered fleet walk. Each row places
measured Git evidence (HEAD SHA, commit timestamp, base divergence, and dirty
count) beside branch-associated issue metadata whose declared status is
explicitly attributed with `branch-prefix` provenance. It reports observations
only: neither the typed contract nor the human renderer derives `cold`, `drift`,
`liveness`, `staleness`, or a similar judgment.

Rows also carry this machine's tracker **claims** (#288): `fleet.PlaceClaims`
(pure, `internal/fleet/claims.go`) joins each repository's claims — one
`LookupRepoClaims` read per repository, sharing the branch-prefix lookup's
cached records load — to the row whose tree is the claimant's worktree, judged
against `machine` (the `localMachine` derivation `claim` records). Claims whose
worktree matches no row are `dangling_claims`. Each row's `claims_state`
(present/stale/partial/unknown/absent) is read quality, so an empty list is
"none" only when the read carried value. This is the bulk, cross-slot
counterpart of `issue show --json`'s per-issue assignment.

**Slots** (#289): `discoverSlots` finds each workspace from row paths
(`repo:0` = a fleet primary, `repo:N` = `workspace.SlotPath`); a numbered slot's
members come from `DeclaredMembers` (`internal/fleet/membership.go`), the
transitive walk of the host's `construct/deps` substrate rows via
`pkg/layergraph.ParseRows`, confined to the environment root. Declared clones are
collected as rows (independent clones the fleet walk never sees) and read their
tracker through the fleet primary with the same publication identity
(`gitx.PublicationRepository`), so they add no network read. `JudgeCheckout`
(pure, `slots.go`) judges each member from its facts — including the operation
in progress, read from the worktree's git directory with no git process
(`workspace.ActiveOperation`) — and its recorded issue-lookup outcome and
claims; `AssembleSlots` folds the worst. A detached member needs recovery
unless it is staged (#344): `TreeRow.Staged`, set when HEAD equals the
worktree's `sdlc stage` record (`fleet.StagedCommit`), is the hold `staged`. The inventory JSON is versioned
(`schema_version` 1). Measured here: 6.3s with 39 slots (+0.4s over #290).

Tracker reads are the inventory's network cost (#290). `repoRecords` is a
per-key-once cache, warmed for every tracked repository (`tracker.CutOver`,
local) with at most 8 concurrent loads before the row walk. Each load waits at
most 15s on its remote (#360): the budget (`TrunkFile.BoundRemote`) covers only
remote contact (ls-remote, fetch, push), shared across them, and local git work
is not counted, so a loaded machine cannot spend it. A hanging remote then reads
as stale, from the last fetch, like any unreachable remote (unknown if it was
never fetched); a load that ends on the deadline is kept, so the walk never waits
twice. A read is
one `ls-remote` when the tracker is unchanged: `TrunkFile.RemoteExists` hands
the tip it saw to the next `Snapshot`, which reads it directly when the local
tracking ref already points there, and fetches otherwise. Tracker fetches pass
`--no-auto-maintenance`. Measured here: 17.5s → 5.9s, identical output.

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
backstop that catches any test that still leaks. A lock present after the run
counts only when this test process holds it or its holder is dead
(`lockMayBeOurs`, #253): every slot shares the lock, so another slot's live verb
holding it is not a leak.

The lock path is resolved from `git rev-parse --git-common-dir`, so linked
worktrees for one repo share the same lock. That is intentional: worktrees share
the issue namespace, object store, and remote refs that the motivating races
touched. The lock does not serialize another clone or machine, so remote
push/ref races still surface through the existing push/merge retry guidance.

`change-code`, `close` and `milestone-close` prepare under the lock, release it
while external reviewers run, then reacquire and validate before recording any
response (#244). `reviewstate.go` shares the capture/compare policy: canonical
repository and private worktree Git directory, branch identity, issue/plan bytes
(including absent plans), and the relevant gate ledger. Close also protects every
prepared project edit and its first-round plan-ledger seed. `change-code` requires
exact HEAD; close retains the existing docs-only descendant allowance when its
captured inputs are unchanged. Validation precedes every verdict's sidecar and
ledger writes, including failed or malformed reviews. Two responses prepared
against one ledger cannot both advance it.

`planningreview.go` owns change-code's prepare/unlock/relock phases while keeping
plan-quality before estimate checks. Cancellation, dispatch failure, stale inputs
and failed reacquisition stop without new authority writes; `--force` cannot waive
these safety failures. Review timeout: boundary reviews scale with the window's added lines
(`judge.ReviewTimeout`: 30m up to 500, +15m per further 1,000, at most 2h; #300),
plan reviews get 30m; `WF_REVIEW_TIMEOUT` (1s–2h) overrides both.
The process runner bounds graceful shutdown and pipe draining to five seconds;
Unix reviewers have owned process groups so cancellation also kills descendants,
and the direct reviewer is reaped before returning. CLI interruption is carried
through the review context; no asynchronous writer survives command completion.

`merge` holds the lock around short integration Git operations, the leased push,
the final expected-head merge, and archive/cleanup. Publish replay, GitHub PR
registration and CI polling, and the local suite run happen unlocked. Every
reacquisition checks the exact branch HEAD and worktree and fetches main: main
drift is reclassified for integration or exact-bound issue-only readiness, while
HEAD drift refuses. The legacy
dependency-clone path also releases the lock for local checks and refuses drift
before server merge. `push` retains its whole-transaction lock. Wait messages
name the live holder; wait or retry rather than deleting its lock. Recovery is
conservative but not wedging:
`die()` drains the active lock cleanup registry before `os.Exit`, missing
`meta.json` during the tiny mkdir-before-write window is treated as holder
initialization and polled through, and a confirmed-dead same-host holder is
reclaimed under a nonblocking flock on the opened candidate directory (#413).
Reclaimers verify its inode and holder generation, recheck liveness under the
guard, then revalidate the inode, cancellation and deadline before atomically
renaming it for removal. Busy, changed or initializing candidates return to the
existing bounded observation loop. Closing the descriptor releases the guard;
no additional durable lock file is created. Unsupported flock or permission
errors refuse rather than falling back to unguarded reclamation.
"Dead" means `kill(pid, 0)` answers ESRCH; EPERM is a live process this one may
not signal, which is every peer slot's process when sdlc runs in an agent
sandbox (#330: counting it dead let a sandboxed verb take a live peer's lock).
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
(`closeResult.projectEdits []projectEdit`), each carrying its `repoDir` and,
for a peer repository, the `peer` mark and the edit's inputs for the origin
publish below. A **TL-owned** match (frontmatter
`tl:`, read by `project.TLOwners`, #326) is skipped before it reaches
`projectEdits`, so the tick, the detail-block gate, the peer publish and the
review snapshot all pass it by; the close prints an info line naming the TL,
and the worker reports the landing there. `find`/`resolve` still list it.

### Peer projects publish to their origin (`peerproject.go`, #308)

A close never edits a *peer* repository's checkout: that tree is someone
else's (the TL's live ops checkout, kink 38), and a local commit there
diverges it from origin. `projectCloseEdit` (`projecttick.go`) is the one pure
edit — tick the rows, upsert a milestone's detail block, or refuse a missing
block with the authoring skeleton. `computeClose` previews it on the local copy
(messages, the missing-block refusal, the dry run), and `applyClose` writes
this repository's projects locally (they ride the close commit). Last, after
the local writes, `publishPeerProject` re-derives the same edit on the peer's
fetched main and publishes it with one `TrunkFile.Update` compare-and-swap
commit (`project: close-time update (<repo>#<id>)`) to the upstream of the
peer's resting branch. The path is symlink-resolved (`peerProjectPath`), so
the commit changes the real file. A moved origin is re-read and the edit
recomputed on its bytes, within `TrunkFile`'s bounded attempts. A copy that
became TL-owned on origin, or whose `tl:` cannot be read, refuses; so does a
brain, or a peer with no upstream. A refusal is a warning naming the file and
never fails the close. A brain's legacy `data/project` home is the exception:
it is not a peer, and its tick is written in place, uncommitted, for nous to
commit on its own rhythm, as before #308. A peer is always published, because
origin's bytes decide whether there is a change; the local copy only previews. Peer projects are not in the review snapshot, so a TL's
mid-review edit cannot void a SHIP (kink 55). `--no-project` skips the whole
project update; `--force` only waives a missing detail block. Pinned by real
multi-repo fixtures with bare origins (`peerproject_test.go`).
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

### Project trains and forecast (`pkg/trains` + `trainforecast.go` + `projectforecast.go`, #329)

A project is a TL's coordination plane with no deadline (#329 retired the #182
calendar forecast, the throughput baseline and the Phase-A fog/calendar
ledgers along with `deadline:`/`planned_finish:`). The forecast is derived on
read and never typed in.

- **Trains parser (`pkg/trains/trains.go`, pure; shared with `tl-lead`, #350; `internal/project` forwards the names).** `ParseTrains` reads
  `## Trains`: a `| Train | Cars… | Slot |` table (a bare `#N` inherits the
  previous car's repo; parentheses are annotations; every separator means "next
  car") and `⇒` edge bullets (prose after a colon is the reason; `~` bullets are
  ignored). A malformed row is skipped with a warning, never a panic (fuzzed).
  `TrainPlan.Locate` answers "which train, which car" for find/resolve.
- **Calibration (`internal/estimate/forecastcal.go`, pure).** `CalibrateForecast`
  takes the newest split row per issue with an estimate (the `ConvergeFactors`
  eligibility plus `ExcludedFromCalibration`): point multiplier
  Σ(build+converge)/Σestimate (= estimate ÷ build rate × (1 + k)), range the
  25th–75th percentile of per-issue multipliers (collapsed below 3 rows), class
  prior the median estimate.
- **Scheduler (`internal/project/trainforecast.go`, pure).** `ForecastTrains`
  builds the car DAG (train order plus in-plan ⇒ edges), computes finish times
  at low/point/high with Kahn's algorithm, and reports the critical path, the
  train it ends in and the total remaining. A cycle is an error naming a car.
  Landed cars count 0; unestimated or unreadable cards take the prior and are
  named. `RenderTrainForecast` is the one statement every surface prints.
- **IO seam (`projectforecast.go`).** `forecastForProject` resolves the
  project's checkout (its repo name for bare refs, its root as the lookup
  vantage), reads each car through `projectIssueLookupFn` and the ledger
  through `calibrationLedgerPath` (`WF_CALIB_LEDGER` override). Surfaces:
  `sdlc project forecast` (errors when there are no trains or no calibration)
  and `project show`/`status` (`forecastLine`, best-effort and quiet without
  trains).
- **Navigation.** `DiscoverInRoots` matches a record that cites the issue as a
  `[repo#N]` row ref OR carries it as a train car, and fills
  `ProjectMatch.Train/Car/Cars`; `project find` prints `(train: X, car i/n)` and
  `resolve --kind project --json` carries `train`/`car`. A train-only match has
  `RowRef=false`, and the close-time tick skips it with an info line (no row
  to edit, so no detail block is demanded).

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
                       remains the loud escape hatch. Landing validates the
                       branch's own window, merge-base(main, head)..head, via
                       landinggate.go's `landingMergeBase`, shared with the
                       duplicate-id gate (#322)
  start.go             migration stub (REMOVED in #39 — errors with
                       "use claim + change-code")
  claim.go             fresh remote status claim + local issue checkpoint (#244)
  claimdecision.go     pure open-status reservation decision
  issuepublish.go      selected documentation commit publication command
  commitpublication.go pure workflow-path eligibility policy
  synctrunk.go         new-issue reservation with per-attempt ID allocation
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
  merge.go             ← Makefile merge; durable routing before legacy cleanup
  landing.go           configured target, durable PR/merge and safe return/deletion
  ghlanding.go         typed GitHub PR identity and expected-head merge
  landingarchive.go    PR-owned remote archive and retry provenance
  milestoneclose.go    composition over close + judge milestone-review
  actualsplit.go       #234: the build/converge split's IO shell (gate ledger → rounds,
                       cutTimes, row metadata); the rule itself is activetime/split.go
  convergereport.go    #234: `sdlc converge-report` (estimate.ConvergeFactors)
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
                       anchor), RunGit,
                       IsShippedWorkSubject/ShippedWorkOnMain — #76 ship probe;
                       TrunkFile + the `runGitIn` shim — #209, see below)
    issue/             frontmatter parse/edit + plan-section regexes +
                       scaffold.go (NextID/Slugify/Render — #56)
    judge/             Category enum, prompt builder, classify, dispatch
    project/           project-file core: line-preserving typed Doc/Task parser +
                       checkbox/frontmatter/section mutations (#180 M2); model-derived
                       scaffold, pure summaries, and pure named guards (#180 M3), alongside
                       the legacy brain-residency lookup/detail-block helpers (#171
                       will lift residency)
```

## Publishing to the trunk (`gitx.TrunkFile`, #244)

Both reservation and selected-commit publication build a commit whose parent is
an observed remote tip, then push with an explicit
`--force-with-lease=refs/heads/main:<observed>` expectation. This detects forward
movement and rewinds. Plain fast-forward acceptance is not an exact-ref guard.
Each transaction permits at most three attempts and leaves caller branches,
working files and the caller index untouched.

`UpdateMany` serves reservations. Its `prepare(*TrunkView)` callback derives the
whole write set again from each fresh tip. Claim requires open status; creation
chooses another free ID for every occupied ID, even the same slug. The claim path
never renumbers. File existence is distinct from empty contents, and deletions
are explicit in `TrunkWrite`.

`SelectCommit` pins a source and its sole parent and enumerates every changed
path. Command-level eligibility accepts ordinary Markdown workflow records;
root/merge commits, symlinks and submodules refuse. `PublishCommit` uses
`git merge-tree --write-tree --merge-base=<source-parent>` to merge only the
selected change into fresh main. Git capability failures and conflicts refuse
without mutating caller state. The resulting source-message-preserving commit
carries `Source-Commit:` provenance; history queries share a 30-second deadline.

`publicationStep` owns the retry/outcome classification. Selected publication
can confirm a lost acknowledgment by source or publication reachability. A
reservation remains uncertain after an ambiguous acknowledgment: a peer can
produce an identical commit, so reachability does not prove this caller won.
Explicit rejections and up-to-date reservation responses recheck remote status
instead. Failed probes never become evidence of absence or permission to replay.

The `runGitIn` / `runGitInContext` boundary preserves separate stdout and stderr
and an explicit repository directory. Read-only degraded reads can warn and use
a stale ref; publication requires a fresh fetch. Tests use a stateful commit/tree/
ref fake at that same boundary and real bare-Git fixtures in normal test runs,
including receive-pack races, remote rewinds, lost acknowledgments, identical
candidate commits, conflicts and retries after later edits or reverts.

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
against a **per-gate signature catalog** (`internal/processmanual/gatesig.go`,
`GateCatalog` — the one owner of each command's bypass set, whose `Gate` field
help pages render via `{{GATE_FLAGS}}`; ACK grammars: G1 close·mclose `--no-X (or --force): …`,
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
push/merge preflight) resolves through `internal/judge.ResolveReview` (#409; it
replaced `ResolveAgentCLI`). The precedence is: explicit `--agent`, then `AGENT_CMD` (operator/script override,
including invalid values so dispatch validation still fails loudly), then
`PAIR_AGENT`, then conservative process/session signals such as `CODEX_CI` /
`CODEX_THREAD_ID`, then `claude`. Command flags stay empty by default so
environment defaults live in one place; call sites pass an explicit-source bit
from Cobra's `Flags().Changed("agent")`. Close dry-run uses the same boundary
review prompt/options builder as real dispatch and prints the would-be command
line, so `PAIR_AGENT=codex` is inspectable as `codex exec` before any subprocess
runs.

**The reviewer's whole run, and when it didn't run (#300).** The claude
reviewer runs `-p --output-format stream-json --verbose`; for claude only,
`judge.ReadStream` reads the event stream into an `AgentRun` (every assistant message's text, any
unparsed tail, the terminal `result`). `Dispatch` returns the messages joined
in order, so the existing last-block-wins parsers (`ParseVerdictBlock`,
`gatestate.ParseFindingsBlock`, and `ParseVerdictToken`, now last-match too)
take the latest verdict and findings: a block-less postscript after a
backgrounded job erases nothing. A run with no verdict signal anywhere
(`HasVerdict`, which recognises everything any parser reads) is dispatched once
more inside the same deadline; two such runs return both texts, labelled, and
keep the `unknown` fail-safe; an error from the retry still returns the first
attempt's text, which every caller shows. `judge.RunFailure` decides "review did
not run", never over a verdict the run already gave: a stream `result` with
`is_error`, or for codex and gemini (read as plain text) a non-zero exit with a
failure signature on stderr or in the output. A network cause is
`ErrAPIUnreachable`, naming the agent's own API host to allow in the sandbox;
any other is `ErrReviewDidNotRun`. Either is persisted as nothing and not
retried. Every contract carries `UnattendedRule` (foreground
only, no notifications). Output that isn't a stream reads as plain text, as
before.

**The reviewer runs in a disposable checkout (#204).** Every close and
milestone boundary review round runs its reviewer in a fresh shared clone of the
pinned Head (`internal/reviewbox`). It is made by `git clone --shared
--no-checkout`, then `remote remove origin`, then `checkout --detach <head>`.
Objects are read through alternates. Refs, config, index and hooks are the
clone's own, and it has no remote back to the source. It keeps the source's
main authority, read-only (#447, `copyMainAuthority`, allowlisted by the pure
`mainAuthority`): `branch.main.{remote,merge}`, each named remote's fetch
refspec, and `refs/heads/main` plus every `refs/remotes/<R>/main` at the
source's commits. A repository test that resolves main and the merge base (pair's
repolint) then runs as it does in the slot. Each copied remote's URL and push
URL is the inert `readonly://review-checkout`, so a push or fetch fails. The
source's real URL, credentials, `insteadOf` and push settings are never copied.
The live issue and plan
files are overlaid at their paths, since they can be uncommitted at close. The
prompt's `repository root:` and its `git -C` recipes name the clone, and only
such an isolated prompt (`PromptInput.Isolated`) tells the reviewer to revert
in place. `sdlc judge` and plan reviews still run in the caller's checkout and
keep the scratch-copy instruction. A missing-verdict or invalid-response retry resets the clone first
(`DispatchOptions.BeforeRetry` → `Checkout.Reset`), so attempt 2 never reads
attempt 1's mutations. The process is one argv transform shared by every
agent (`Isolation.Wrap`, applied through `judge.DispatchOptions.Wrap`):

    [sandbox-exec -p <profile> | bwrap --ro-bind …] env -C <clone> <agent> …

The write fence is a deny list over every worktree of the source repository
(the reviewed slot, `:0`, sibling slots) and its git common dir. Paths are
canonicalized, since macOS `/var` and `/tmp` are symlinks. On macOS the fence
is Seatbelt; on Linux it is `bwrap --ro-bind`. Under the fence, each agent's
own sandbox is turned off (`DispatchOptions.Fenced`: claude `--settings`
sandbox disabled, codex `--sandbox danger-full-access`, gemini
`--sandbox=false`). Seatbelt cannot nest, so one outer fence is what gives the
three agents the same writable scope. Darwin's setuid `/bin/ps` cannot execute
under even an allow-default Seatbelt profile. The profile adds exactly
`(allow process-exec (literal "/bin/ps") (with no-sandbox))`; this exempts the
system process-listing binary, while its invoking shell and other children remain
subject to protected-source write denials. No wildcard executable exemption or
broader sandbox model is introduced; Linux's fence remains unchanged.

That scope is wider than the agents' own sandboxes in one direction: HOME and
other repositories stay writable, because agents write their own state there.
It is narrower where it matters, since the reviewed repository and all its
worktrees are fenced.

`reviewbox.Probe` reports a fence only after seeing it work. Through the same
`Wrap`, it must write an allowed scratch directory and be refused a protected
one, so every `fenced` round had a live denial behind it. On Darwin it also
requires `/bin/ps` to inspect a known live PID. Live conformance covers direct
and shell-child ps plus allowed and protected writes. When the probe fails,
the round degrades to the clone alone, with a warning. Slot agents normally run
sdlc outside Claude Code's Bash sandbox (`sdlc *` is in `excludedCommands`,
#334), so the fence can apply. When sdlc is called in a form that stays
sandboxed (after `cd x &&`, piped, by full path), `sandbox_apply` is refused
and the round runs clone-only. The ledger round's `isolation:` and the
sidecar's `| isolation |` row record `fenced` or `checkout`. Degrading was the
TL's decision on #204. `--dry-run` prints the wrapped command with a
`<review-checkout>` placeholder.

**Pinned preparation and answerable responses (#364).** Boundary reviews read
`.sdlc/review-prepare.json` from the pinned HEAD's regular Git blob, not the working
tree. Absence means no preparation. Version 1 accepts `command` (a nonempty argv
array) and optional `timeout_minutes` (default 10, explicit range 1–30):

```json
{"version":1,"command":["make","runtimebundle-generate"],"timeout_minutes":10}
```

Invalid JSON, unknown fields/version, trailing data, empty argv/arguments, invalid
timeouts and nonregular entries such as symlinks are setup errors. The command
runs without an implicit shell in the disposable checkout, under the same Wrap
and confinement environment as the reviewer. The remaining dispatch deadline
also limits preparation. `--dry-run` displays the declaration and ordering without
starting either process. The live source never receives preparation writes.

Attempt order is checkout → confinement → preparation → review → response
validation. The single retry resets the checkout, recreates confinement, and
prepares again. Launch errors, nonzero preparation exits, timeout, cancellation
or tracked-tree mutations stop before that reviewer attempt. Untracked/ignored
build outputs are allowed and removed with the disposable checkout and its crash
sweep. Process-group cancellation and reaping use the existing bounded runner.

REWORK requires a new blocking finding with nonempty title/detail, or an explicit
`not-addressed` disposition of a known blocking finding in scope. An old open
finding alone, Minor-only finding or unknown ID does not qualify. New findings
use `id: new`; admission precedes BR-n assignment. All referenced dispositions are
validated against the pinned scoped ledger. Optional `kind: unverifiable` names
an environment-limited check; omission preserves ordinary defect semantics. Its
detail names the attempted command, observed restriction and evidence needed.
After `sdlc finding dispute` supplies measured host evidence, the next round must
explicitly dispose every pending unverifiable dispute with a nonempty note
answering it. Ordinary historical dispute semantics remain unchanged.

Invalid responses and missing verdicts share one two-attempt budget. The retry
gets the admission error; a second invalid response produces a protocol error.
Neither invalid attempt contributes findings, reviewed-head advancement or close
finalization. Preparation and protocol diagnostics remain separate from semantic
review text. Preparation output drains beyond a 64 KiB per-attempt retention cap;
the complete additional diagnostic envelope, including metadata and rejected
response excerpts, is bounded to 256 KiB per dispatch with truncation and retained
byte counts. The existing append-only review sidecar retains it permanently and
archives it with the issue, including failed invocations with no accepted round.
Four maximum envelopes add at most 1 MiB (Git storage overhead is additional).
Accepted review bodies retain their existing behavior outside this new budget.

Ownership: `internal/reviewprepare` parses declarations and renders diagnostics;
`reviewprepare.go` reads pinned blobs; `internal/judge` executes each prepared
attempt; `boundaryresponse.go` validates admission; `reviewsidecar.go` persists
evidence after the existing stale-state checks.

**Every reviewer runs confined (#339).** The fence covers the source
repository only, and some rounds run clone-only, so `judge.Dispatch` confines
every agent it starts: the boundary review, plan-quality, estimate-quality and
`sdlc judge` (`reviewbox.Confine`). Before the first attempt it makes a base,
`<clone>/.sdlc-review/` (excluded from the clone's status), or a fresh
`$TMPDIR/sdlc-review-*` for agents that run in the caller's checkout. It proves
`tmp/` writable; a failure means the review did not run. The agent's
environment (through the run context, so argv and the `Run` seam are
unchanged) sets `TMPDIR`, `TMP`, `TEMP` and `CLAUDE_CODE_TMPDIR` to `tmp/`, and
`GIT_CEILING_DIRECTORIES` to the root's parent and the base. `bin/` goes first
on PATH, in the process environment and in `env.sh`, which `CLAUDE_ENV_FILE`
names. Claude Code's Bash rebuilds PATH from the user's shell snapshot, so the
file it sources before each command is the only way the shims reach claude
(measured on #339). `bin/` holds a `git` shim that refuses `init` and `clone` outside the root
and the base. On macOS it also holds a `mktemp` shim: without a template, macOS
mktemp writes under `_CS_DARWIN_USER_TEMP_DIR` (`/var/folders/…/T`) and ignores
TMPDIR, and an agent sandbox denies that directory. That is how the #309
reviewer's `mktemp` failed, landing its `cd` in `$HOME`, where it ran `git
init` and `git add -A`. After the last attempt, `Confinement.Check` compares a
snapshot of `.git` in every ancestor of the root and base and of `$HOME`'s
non-hidden entries. Anything new is a dispatch error that names the paths, so
the gate halts. The shims stop accidents, not a reviewer that calls the real
binaries by path. Without an OS fence, the check is what catches the rest.

Lifecycle: `<WF_REVIEW_CHECKOUT_ROOT | user cache dir/sdlc/review-checkouts>/<id>/`
plus `<id>.lock`.
- The lock is flocked as `<id>.lock.tmp` and renamed into place, and held for
  the round.
- `dispatchBoundaryReview` removes both in a defer once the reviewer has
  exited. Success, reviewer failure, timeout and interruption (SIGINT/SIGTERM
  via the CLI context) all unwind through it. The review text is in memory by
  then.
- If the clone can't be removed, its lock file stays (only the flock drops),
  so a later sweep retries.
- Each dispatch first sweeps the root, best effort: an unheld lock or
  `.lock.tmp` means a dead owner. It removes that checkout, along with any
  lockless directory, restoring write permission on read-only trees. The
  kernel drops the flock with its holder, so there is no pid-reuse guess.
  Entries it cannot remove are warned about and skipped, never a reason to
  refuse the review.
- What a SIGKILLed sdlc cannot do is reap its reviewer, which ends with its
  own turn.

Bound: one clone per live review, plus whatever the sweep has warned about.

**Which agent reviews, and its permission mode (#310).** `workshop/review.json`
(repo-local) sets `boundary_reviewer`: `same` (default), `other` (a reviewer of
another vendor family, #409 below) or `explicit` with `reviewer_pairs`. The pure
`judge.ReviewerPolicy` maps the main agent (`judge.MainAgent`: `PAIR_AGENT`,
codex env, claude) to its reviewer; `--agent` and `AGENT_CMD` stay above it, and
`same` is the default (`judge.DefaultReviewerMode`). `loadReviewerPolicy` reads the
source checkout's file; an invalid file halts the review. The ledger round's
`agent` names the reviewer CLI and its `reviewer` field the full #409 record
(model, families, degraded). `DispatchOptions.Containment`
(none / checkout / fenced) replaced `IsSandbox` + `Fenced`, and `agentArgs` maps
it to each CLI's own mode: claude `--permission-mode dontAsk` (the tool list
is enforced, so Edit/Write are refused; Bash stays unrestricted, so shell
writes are stopped only by the #204 fence), codex `--sandbox read-only | workspace-write |
danger-full-access`, gemini without `--yolo`. No judge gets an auto-approve mode;
`--sandbox` on `judge`/`change-code` is a deprecated no-op.

**Reviewer families (#409).** The pure policy (families, change kinds, modes and orders,
`Plan`, `Resolve`, `Record`, the `workshop/review.json` loader) lives in `pkg/reviewpolicy`,
and the session-agent detection and `Agent:` trailer parsing in `pkg/agentenv`, so
`cmd/doc-review` shares them (#441); `judge` and `agentid` re-export them under their
historical names and keep the run-shaped parts. `judge.FamilyOf` maps each agent to its vendor (claude →
anthropic, codex → openai, gemini → google; anything else is unknown, never guessed), and a
review's `judge.ChangeKind` is `design-heavy` (a design: plan-quality, planning review, a
full-flow boundary) or `bounded` (a quick-flow boundary). Under `other` the pure
`judge.ReviewPlan(policy, main, authorFamilies, kind, availability)` lists every agent whose
family authored none of the window, in the kind's order (design-heavy: claude, codex, gemini;
bounded: codex, claude, gemini, so #310's claude ↔ codex pairing holds and gemini is the
fallback). `reviewer_order` and `reviewer_models` in `workshop/review.json` override the
orders and pass a model per reviewer. `same` and `explicit` stay #310's floor.
`judge.LocalAvailability` skips, before launch and with a reason, an agent whose CLI is
missing or codex under checkout containment on macOS. `judge.Unavailable` counts a run as
"could not review" only on a failed launch, claude's stream failing before any assistant or
tool event (`AgentRun.Began`), or codex/gemini exiting non-zero with nothing on stdout (codex:
no activity block on stderr, and the cause read only from its `ERROR:` lines, since it echoes
the prompt), each with an unreachable, quota, sign-in or CLI-refused cause (a 400 bad model is
a misconfiguration and halts); only that may move a review to the next candidate. Every review site (close, milestone-close, fix review,
plan-quality, estimate-quality, planning review, `sdlc judge`) resolves through
`resolveReviewChoice` (cmd/sdlc/reviewchoice.go: the window's trailer families plus the
session's, the kind, `reviewAvailability`) and runs `judge.DispatchChoice`, which announces each
skip and ends in `judge.NoReviewerError` (naming every candidate) when none can review. The
result is `judge.ReviewerRecord`: the close's `Review-Reviewer:` trailer, the ledger round's
`reviewer` and the sidecar's reviewer row. A week's tally is the git one-liner in `close
--help` (REVIEWER); the trailer's form is pinned by a test that parses it back through
`git interpret-trailers`.
`reviewer_order` needs two families with no repeats, so `Reviewer()` can never return the
main agent's own family under `other`, and a model reaches each CLI's flag (`DispatchOptions.Model`). The default for a repo with no policy is the single constant
`judge.DefaultReviewerMode` (`same` until the operator decides the fleet default).

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
closes the loop the hand-kept validation log never did.

**Ownership and replay (#335).** Tracker-backed actuals use claim→release
intervals, with a fresh origin on reclaim. Sources follow the owning worktree's
session roots and explicit child dispatch lineage, including retained roots after
relocation; the checkout running close supplies no attribution authority.
Unbound brain sessions are excluded and their count/minutes remain visible.
`actualinputs.go` resolves code/tracker pins, windows, source identities and exact
byte prefixes; `actualmeasurement.go` executes that same input for actual, close,
and milestone-close. Boundary rounds retain the input snapshot so post-close
measurement replays the original right edge and phase cuts. Incomplete required
evidence refuses adoption; legacy no-tracker repositories retain a labeled old
heuristic. Completed historical closes without snapshots are reconstructed from
their exact completion binding and diagnosed explicitly.

**Parked claims (#443).** A claim can remain open while its slot works another
issue. For a new actual candidate, `actualexclusion.go` first compares each
numeric adoption's completion-pinned measurement window with the candidate's:
a retained claimant on a done card does not extend credited time. For windows
that may overlap, it checks same-slot engagement, replays pinned sources, and
protects positively credited activity intervals before the
candidate's commit attribution. The report names each removed interval and its
adopted issue. Missing replay or slot evidence refuses a numeric actual and
names the potentially overlapping interval. A fractionally credited run protects
its full time span, conservatively, rather than claiming an invented split.
The chosen intervals are retained inside the candidate's bounded
`MeasurementInputs`; historical completed actuals replay that same protected
generation even after other cards or transcripts change.

`computeActualForCloseFn` explicitly selects `actualCloseCandidate`: a close
derives new ownership-proven inputs at current HEAD/current tracker, even when
the card holds an older completion or actual correction. `computeActualFor`
retains historical replay for preview/correction. This distinction prevents a
fix re-close from binding old measurement pins to its new reviewed head.

`issue correct-actual` measures a completed generation (or excludes a proven bad
sample), records correction provenance on its card, and appends exact-target
calibration corrections. A pure `internal/actualcorrection` transition model owns
tracker uncertainty, calibration append, mirror publication and repeat/resume.
Git refs under `refs/sdlc/actual-corrections/` retain pending evidence and objects;
`issue recovery` resumes them without another measurement. Ordinary close and
correction writers share the ledger-directory lock. Calibration readers resolve
exact raw row invalidations before trust/filter selection and reject malformed
correction evidence.

For ambiguous legacy rows, `correct-actual --target-row <full SHA256>` selects
only within the eligible effective set. The diagnostic names the ledger,
completion, candidate metadata and exact dry-run commands; TL/operator provenance
selection is required, not inferred from row order. Existing recovery target
hashes bind the choice across retries; conflicting choices refuse. An untrustworthy
numeric measurement names a same-target exclusion command for a proven invalid
sample, preserving the reason and avoiding invented hours. No snapshot or
correction schema changes are needed.

**Build vs converge (#234).** A measured actual is split at the issue's
boundary reviews.
- **Build** runs to each boundary's first review dispatch.
- **Converge** runs from that dispatch to the boundary's last `#N` commit at
  or before its final dispatch.
- **Tail** is the verification after that commit (the final test run and
  review). It is uncounted at *every* boundary (operator, 10-10).

The actual is build + converge, so milestone issues measure lower than under
the pre-#234 rule, which counted every earlier milestone's tail.
- **Where it lives:** `activetime.BuildTimeline` / `Split` (pure, `split.go`)
  encode the rule over one `Compute` result. The time-proportional cut is exact
  because runs are contiguous. `computeActualFor` (`actual.go`) and
  `actualsplit.go` are the IO shell.
- **Cut evidence:** the boundary gate ledger's rounds. Each round stamps
  `dispatched` beside its finalize `timestamp`. A round without the stamp falls
  back to the finalize time, and the row's source says so.
- **E1:** `milestone-close` adopts its own increment, which is that
  milestone's phases (`sdlc actual --milestone Mx` previews it).
- **Ledger:** rows gain `build_hours converge_hours boundary_rounds
  split_source review_recipe` (indices 23–27). `split_source` is one of
  `dispatch`, `finalize`, `no-review`, `manual` (typed `--actual`), or `-`
  (pre-#234, not comparable). The round metadata is read from the persisted
  ledger at row-write time, because the measurement runs before this close's
  dispatch.
- **E3:** a row with no estimate is not a calibration point whatever its flow.
  Its ratio cell is `-`. The historical backfill is ariadne#323.
- **Drift and k:** drift (`estimate.DriftVerdict`) compares the estimate with
  **build**, and reads only rows that carry a split. Converge is calibrated as a
  factor, k = Σconverge/Σbuild per model, flow and recipe
  (`estimate.ConvergeFactors`, `sdlc converge-report`). Grammar + closed
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

**Per-gate bypass (#67).** Each `close` gate has its own `--no-<gate>` flag.
The set is owned by `processmanual.GateCatalog` (pinned to the registered flags);
help pages render it with `{{GATE_FLAGS}}` rather than keep a copy, and
`TestGateFlagListsAreRenderedFromTheCatalog` refuses a hand-listed gate flag.
`closeFlags.skip(gate)` is the single arbiter (`Force || the field`). A
per-gate flag is an *acknowledgment* that one guard doesn't apply (e.g. a
pure bugfix → `--no-atlas`); it logs an audit `[!]` line and only fires
when the gate would actually have refused. `--force` waives all at once.
The atlas gate additionally auto-satisfies (info line, no flag needed) when
the window contains no code surface — `hasCodePath`, the single docs
classifier: `*.md` / `workshop/` / `atlas/` / `docs/` are documentation,
everything else is surface (#177; docs-only closes had an incoherent demand).

**Issue branch identity (#471, `trackerbranch.go`).** A tracker card's filename
derives its one issue branch. `abandon` and attachments to started work require
that branch before changing the checkout or publishing evidence. Whole-issue
`close` and `milestone-close` also admit a renamed recovery branch only when
its HEAD exactly equals the local canonical branch tip (or a fresh remote tip
when the local ref is absent); another head refuses. Abandon may resume from rest only when
its persisted branch still matches the card. The remote branch is deleted
only at the head held by its archive ref, with leased advancement and deletion;
a remote commit beyond the local issue tip refuses before archival.

**Integrate gate (#269, `integrategate.go`).** Before planning the window,
`close` and `milestone-close` (fixed-then-ship included) require the branch to
contain every first-parent commit of freshly fetched main whose committer date
is at or before the most recent noon (ariadne-robustness-1 requirement 6). The
definition lives in **`internal/freshness`** (`Cutoff`, `LoadLocation`, `Check`),
shared with #303's catch-up at start-plan. The zone comes from the tracked
`.sdlc/freshness.json` (absent → local; malformed refuses). Main is
`reviewMainRef` in a tracker repo, so the gate and the window share one fetch,
else `TrunkRef` after `git fetch`. The probe is tri-state:
- Stale: refuses, naming `git merge` (pushed) or `git rebase` (unpushed).
- DocsOnly (`publishGateHasCodeSurface`, the publish gate's predicate): an info line.
- Inconclusive (a fetch or git failure): a "could not check" warning, never a pass
  line.

**Catch-up before new work (#303, `catchup.go`).** `start-plan --issue N` on the
resting branch fetches origin/main for the slot and each `substrateChain` repo.
If any trails it, start-plan runs the installed `weave refresh --defer-settings` (the `weaveRefresh`
seam; TestMain stubs it), which owns the all-or-nothing fast-forward and compile.
sdlc never fast-forwards a dependency itself. A successful catch-up stops before
the branch or card changes, because the running binary predates it; the re-run
plans; so does a failed refresh that had already moved a checkout (HEADs are
compared before and after). Checkouts the probe cannot read (absent per
`substrateWalk`, no origin, a failed fetch) warn as unchecked. A failed refresh
that moved nothing refuses only for a checkout `freshness.Check` calls Stale,
and otherwise warns. `claim` keeps only its #284 fast-forward of the slot's own
rest. `couch --prepare` (pair#434) is the TL-side twin: same refresh, plus an idle
check and a context reload.

Protected Claude settings are operator-published (#411). Gate refresh excludes
every action producing `.claude/settings.json` and always adds
`plan.DeferredOutput`, even after its declaration disappears. It retains
previous inventory identities and displaced backups without publication or retirement.
When Git is current, `weaveSettingsCheck` invokes read-only `weave settings`,
which reuses composition and output observation and prints one stale notice
for changed declared outputs or retained retirement evidence.
Observation failures warn rather than fail planning. `weave refresh --substrate`
from the operator's shell performs normal publication on an issue branch;
subsequent observation clears the notice. Neither path retries unsandboxed.
`settings.go`, `plan/deferred_test.go`, `settings_notice_test.go` and the native
protected-settings start-plan conformance test own the regression coverage.

The cutoff is a past instant, so a landing after it never moves the requirement.
A fast-forwarded landing counts from its own commit date, which is stricter but
never looser.

The milestone-verdict gate demands per-milestone `Review-Verdict:` evidence,
EXCEPT for **trailing** unclosed milestones (#175): Mx rows after the last
verdict-carrying one — or all rows, when none carries a verdict (the
single-pass over-split shape) — are accepted with a loud info line, because
the issue-close boundary review's window (the branch patch) covers their
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

**Measured actuals (#68, #110, #335, #385).** `--actual` is computed, not
hand-typed. `sdlc actual --issue N` and close's missing-`--actual` adoption share
the native `internal/activetime` engine. Tracker-backed repositories use the
ownership intervals and pinned source evidence described above. No-tracker
repositories retain the explicitly labelled brain/repo cwd heuristic and the
issue's `CommitWindow` plus `activetime.WindowIssues` peers.

Source selection uses `internal/transcripts.Harness`: `Name()` and
`Sources(ctx, cwds) → (Sources{Dirs,Files,Warnings}, error)`.
`DefaultHarnesses()` registers Claude, Codex and Qoder; `Select` merges their
contributions, retaining warnings and discarding partial selections on error.
Claude supplies cwd-encoded project directories; Codex supplies date-sharded
files matched by `session_meta.cwd`. Qoder supplies files matched by recorded
workspace directories, using the same strict `DiscoverEvidence` path for legacy
selection and ownership-based actuals. Failed Qoder evidence requires complete
proof of unrelated workspace roots before exclusion; unresolved or relevant
failures refuse adoption. Proven unrelated failures remain visible as warnings
in both paths. Claude/Codex legacy selection keeps its existing heuristic
handling of malformed or missing metadata. Pure encoders/parsers are tested
directly; filesystem selection is the thin, cancellable IO seam.

The engine
returns a structured `Result.Status`: `TelemetryGap`
(commits-but-0-events → restore owning-slot transcript evidence), `EmptyWindow` (nothing to measure), or
`Measured` (`PerIssue[N]` → hours). Attribution is global-boundary based (#92):
source-scoped activity runs are claimed by nearby issue-referenced commits;
overlaps collapse within one transcript source but not across parallel sources;
only a subject's **lead** refs become claimants (#317, `issueref.Lead`: after
`area:` labels and at most one of `close`/`closes`/`file`/`issue`/`plan`, the leading ref
plus refs chained to it, as in `#174-#176` or `#284, #285:`). Later refs are
citations: they stay in the mention scope (`Commit.Refs` → `WindowIssues`) but
claim nothing, so `#304: log: … (F1, #308)` no longer hands #308 a share. A
subject with no lead is a neutral time boundary. On an issue branch its own
merge commits (branch point..HEAD) are never boundaries, and a main commit bounds the issue's runs only when
its lead names the issue. Integrating main therefore leaves the boundary set
unchanged. The tracker history read beside HEAD is filtered the same way in every
scope (#321): a card commit counts only when its lead names the measured
issue; another issue's card, or a card with no lead, is not a boundary at all. The tracker interleaves every slot's card writes, so off the issue
branch (a resting branch, or main after landing) another slot's card write
would otherwise claim the run beside it. Suspicious attribution is surfaced as `Result.Warnings` and
rendered by `actual` / `active-time`. Dir-selection is deliberately narrow (NOT all
folders/sessions) — an unrelated concurrently-edited repo inflates the count.
`WindowCapDays` is 61 (was 31) so month-long issues keep their window. The
window-**start** is the **engagement anchor** (`resolveWindowStart`) when one
exists, else `CommitWindow`'s parent-of-first-`#N`-commit. It anchors at the
cheap early `claim` so DESIGN attention (brainstorm / spec / plan / reviews) before the first
code commit is in-window instead of cut off; gap-truncation keeps a dormant
claim→work gap from inflating the actual. The anchor is resolved in robustness
order (#116): the explicit `started:` stamp (written once at the open→working
flip — on the tracker card since #252, whose claim/close commits are read beside
HEAD so a freshly claimed issue has a window; local-offset RFC3339 to match `%aI`) →
`gitx.WorkingTransitionISO` (the #113 git-log heuristic, now the legacy fallback)
→ commit-parent. The explicit stamp survives rebases/moves where the heuristic's
"best-effort" history scan could silently miss and drop design time. The claim
wins even when the first `#N` commit (the filing) is earlier (#270, folding
#254's window half): filing is not design, and with boundaries scoped to the issue branch (#270) the issue's own commits
would otherwise claim every other piece of work the slot did between filing and
claim.

Recorded suite queue waits (#420) are a separate input to owned actuals. The
local-check recorder stores exact queue start and acquisition instants beside
`queue_seconds`; legacy records without exact endpoints remain unknown and
cannot support a calibrated numeric actual. A legacy `started_at` follows an
unrecorded setup delay, so a reconstructed wait never proves non-overlap with
an engagement or measurement boundary. The actual resolver selects validated records for the pinned
repository, owner slot and engagement. An exact wait crossing the measurement's
last issue commit is clipped there before ownership is checked; only the part
inside the measured window needs one pinned owner (#430). It then retains the selected intervals,
record digests and source slots in the bounded measurement snapshot. The native
engine applies each wait to admitted activity sources mapped to the attempt's
owning slot and subtracts the entire interval. No tool call, result or task
span inside it is classified or used to restore time; overlapping real work
may therefore slightly undercount. A malformed or unreadable attempt record
whose relevance cannot be ruled out, a relevant attempt missing queue or slot
fields, or a legacy attempt without exact endpoints makes coverage unknown for
both close and correction. Unknown source ownership also refuses a numeric
measurement. The refusal retains every queue blocker: only an all-legacy set
names its attempts and the matching `close` or `milestone-close --no-actual`
command, with its exclusion reason. Mixed or malformed new-format evidence
retains the restore guidance. Old completed snapshots replay unchanged; explicit
`correct-actual` builds a new candidate against the completion generation and
can exclude an unverifiable calibration sample without inventing hours.
The retained snapshot keeps a wait's binding transcript even when all of that
source's activity was subtracted, and queue-aware corrections carry the same
ownership-proven calibration trust as engagement measurements.

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

`push` and `merge` use deterministic publish gates; LLM review belongs to
`close` and `milestone-close`. Judges report findings rather than editing files;
the implementing agent applies fixes before retrying the boundary.

The ordinary-worktree/dependency **legacy merge path** rechecks tracked code
changes and resumes interrupted cleanup using `origin/main` ancestry checks.
Its `assessDirty` helper separates Blocking / Untracked / Tracker entries:
tracker edits and untracked files do not themselves block that legacy gate.
`merge_e2e_test.go` explicitly exercises this legacy behavior with real local
Git and injected GitHub responses. Durable :0/:N landing instead refuses all
tracked dirt and uses exact PR integration plus remote archive evidence, as
described above. Reservation and documentation publication follow #244.

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
Run by hand, `sdlc judge plan-quality --issue N --dry-run` renders exactly the
prompt the gate sends — issue, durable plan and the ledger's prior findings, read
through change-code's own resolution (`planArtifactPath`, `planQualityPromptInput`)
and never written. A live `judge plan-quality` refuses and points at `change-code`:
the gate's decision is its ledger's, and an unrecorded review is the stateless
pre-#187 behavior (#189).
`planArtifactPath` finds the durable plan by issue id, not exact slug (#235): the one
`NNNNNN-*-plan.md` in the plans dir, whatever its slug (sidecars excluded by
`classifyFamily`'s suffix rule), with a note naming the canonical name when it
differs; two candidates refuse. Change-code, close's measure and review snapshot,
the review manifest and `judge plan-quality` all read through it.
Cite the marker (`ARCH-DRY`) in plans/Logs/findings.
`ARCH-EVOLVE` (#454) applies on the quick flow too: persisted formats evolve
additively, with unknown-member capture and preservation through every rewrite,
strict known-field validation, and justified unsafe-ignore version refusals.
Its opt-in live plan-quality fixture test is
`SDLC_LIVE_ARCH_EVOLVE=1 go test ./cmd/sdlc/internal/judge -run '^TestLiveArchitectureEvolve$' -count=1 -v -timeout 20m`.
Default suites exercise the registry/delivery guards and never call a model.

**Adding an `ARCH-*` entry** flows into every RUNTIME consumer with no other edit
— the block header's entry count, `{{ARCH_STAR}}`, every prompt, `start-plan`,
`arch-principles` all derive from `ArchitectureMarkers()`. The test layer keeps two intentional policy tripwires:
`TestArchitectureMarkers` pins full membership and order, and `TestQuickMarkers`
pins the quick-flow selection. Rendered-count and flow-selection expectations
derive from the registry or its shared renderer. A deliberate entry change also
updates affected golden blocks and regenerates the full process manual. See
`atlas/workflow/architecture-principles.md`.

**Deferred principles (#208).** `internal/judge/architecture-deferred.md` holds
entries written down but deliberately NOT embedded, so they carry no gate cost.
Activation is a pure MOVE into `architecture.md`;
`TestDeferredPrinciplesReachNoGate` derives both the forbidden set (scanning the
deferred file with the registry's own `markersIn`) and the gate-facing text
(walking `promptFS` plus the render helpers), so the move empties the set and the
guard stays green with no edit to it. **`sdlc start-plan`** (#75 M2)
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
manifest's read-only inspection recipes and return REWORK with an answerable
blocking unverifiable finding if the repository, objects, or commands are unavailable. This keeps agent argv bounded even for a
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
shared `boundaryReviewOptionsFor` (ARCH-DRY); the same derivation feeds the
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
validates repository/worktree/branch identity, the reviewed commit, issue,
prepared project edits, optional durable plan and gate ledgers before any write.
The shared snapshot and docs-only descendant policy are described under
[Repo transaction lock](#repo-transaction-lock-132). Stale or interrupted results
write neither a sidecar nor a ledger round.
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
artifact: every completed review whose prepared state still validates writes
its semantic final response to a durable sidecar under `workshop/plans/` — `NNNNNN-slug-close-review.md` for a whole-issue
close, `NNNNNN-slug-m<x>-review.md` for milestone `Mx`. The write lives in the
shared finalization path, after lock reacquisition and snapshot validation
(`reviewsidecar.go`: pure `sidecarMeta` + `renderReviewEntry` + `sidecarPath`
behind a thin atomic-write seam — ARCH-PURE). Both close paths share it
(ARCH-DRY). Each file carries a metadata
header (issue id/title, repo, issue file, boundary kind, milestone, base..head
window, command, reviewer, timestamp, verdict) plus that response. Harness
diagnostics, progress, prompt echo, and tool transcript remain terminal stderr
and never enter verdict/finding parsing. The bounded preparation/protocol
diagnostic envelope described above is retained separately in the review sidecar. A re-run of the same
boundary **appends** a timestamped `## Re-review` section rather than overwriting
(the §1 revision convention). The terminal still prints the semantic body + the
`Review-Verdict:` trailer; the sidecar adds a durable surface an agent can reopen
after scrollback loss or compaction (the path is echoed as `review sidecar: …`).
`--no-judge`/`--dry-run`/not-run boundaries write nothing — no body to persist.

**Window base: the branch patch and the interdiff (#304; #58, #77, #197).**
`planReviewWindow` (`reviewwindowplan.go`, pure; `planBoundaryWindow` is the IO
shell) is the single source for the atlas-coverage gate, the boundary review's
manifest, the churn report, the publish gate's quick-flow re-measure, the
`Review-Window:` trailer and the line close prints. They all provably cover the same
diff (ARCH-DRY). Every window is defined on the **branch patch**,
`diff(merge-base(main, HEAD), HEAD)`, never on a commit range. Merging or rebasing
main changes commits, not the patch, so neither widens a window.

- **Whole-issue close, and a first milestone:** the branch patch. Printed as
  `branch patch vs main@<base>: <n> issue commit(s), <f> file(s)`, where the count
  is first-parent with merges excluded.
- **Later milestones:** the **interdiff** since the last finalized review.
  - That review's head `H_r` is stamped on its gate-ledger round (`Round.Reviewed`,
    only when the round finalizes: D5) and found by `gatestate.LatestReviewed`.
  - `gitx.RebasedReviewedBase` replays the reviewed patch onto today's main with
    `git merge-tree --write-tree --merge-base=<merge-base(main,H_r)> <merge-base(main,HEAD)> H_r`
    and wraps the tree in a deterministic, unreferenced commit `S`, which gc
    collects. `diff(S, HEAD)` is the issue's own work since that review.
  - Merge attributes come from main's tree (`git --attr-source=<merge-base(main,HEAD)>`),
    not the working tree, so a `merge=union` file (#320) replays without a conflict
    and `S` stays a function of its inputs. This raises sdlc's git floor to **2.42**:
    review windows and the publish gate refuse older git.
  - Inter-milestone side-quests and fixes still land in exactly one window (#58).
  - A **conflict** keeps git's conflicted tree, so the window shows only the
    resolution and is labelled "includes conflict resolutions in …".
  - A reviewed head that can't be resolved, or a criss-cross history, over-covers
    to the branch patch with a warning.
  - A milestone closed before #304 has no stamped round, so its `Review-Verdict:`
    trailer commit (`previousReviewBoundary`) stands in. Remove that fallback once no
    open issue predates #304.
- **`close --fixed-then-ship` (#183):** the narrow re-close after a finalized close
  review. `planFixedThenShipWindow` plans the same interdiff, anchored on the close
  boundary's own last finalized round (`gatestate.LatestFinalizedAt`), so the window
  is the fixes plus any conflict resolution with main. It never over-covers: an
  unplannable window is `windowNone`, and close refuses, naming the full close. The
  `fix-review` recipe (`internal/judge/prompts/fix-review.md`) disposes the prior
  findings, judges listed conflict resolutions (`PromptInput.ReviewNotes`), and
  treats a scope change as REWORK. The window excludes the issues directory, so
  `contractChangeNotes` compares Spec/Revisions/Done-when, as the issue stood at the
  reviewed head, with now, and passes each changed section as a note (BR-1). The atlas gate and the quick-flow shell stay on
  the branch patch. A code-conflict resolution needs `--full-suite '<evidence>'`
  (`fixedThenShipRefusal`, `closefixed.go`; only the review narrows). Finalizing is
  an ordinary close: a new evidence commit rebinds the card, so the publish anchor
  advances. The publish gate's code-delta and conflict refusals name this path. A
  `fix-review` round never counts as a full review (`fullRoundIn`). Milestones have
  no such flag, because fixes after a milestone close fall in the next interdiff.
- **One fresh main (D11):** `reviewMainBase` measures against main as freshly
  fetched in an issue tracker repository, so a stale tracking ref can't push the base
  behind a main the branch already merged.
- **Pins (D4):** each finalized boundary pins the newest commit its review produced
  at `refs/sdlc/reviewed/<id>/<boundary>` (`reviewpin.go`), so `H_r` survives a
  rebase. The pins are removed when the issue ends (done, abandon, archive) and
  swept by settle and recovery reconcile.
- **Milestone evidence (#197):** in a tracker repository `milestone-close` commits
  its own evidence (`commitMilestoneEvidence`: details plus `<stem>-*` plans
  records, subject `#N Mx: close`, verdict trailers), built in a temporary index
  and swapped in by compare-and-swap. The milestone-verdict gate finds that commit,
  so nothing is hand-pasted. A legacy repository keeps the paste protocol.
- **Known limits:**
  - `Review-Window:` trailers record commit ids as of the review, and a rebase
    leaves them historical.
  - #194's in-review anchor still refuses a rebase made *during* a review.
  - A main revert of code the branch depends on merges cleanly, giving an empty
    interdiff. That is a semantic gap only CI catches.

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

Ariadne's owner `Makefile` reads effective `go env GOFLAGS`, retains those flags,
and exports `-trimpath` to its tool builds and Python test sharder. Already-present
`-trimpath` is kept once. A command-line Make `GOFLAGS=...` assignment retains
Make's explicit override semantics. Generated `construct/dev-aliases.sh`
functions pass the flag directly to every owner's `go build`; existing shell
functions need to be reloaded after the emitter changes. Direct Go invocations
can pass `-trimpath` themselves. The compiler records module/import paths rather
than worktree source paths, allowing unchanged compiled packages to share cache
artifacts across worktrees. Fixture/source discovery uses cwd, Git or module roots.

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

### Machine-wide queue drain (#481)

`sdlc queue pause --owner TOKEN [--reason TEXT] [--cancel-running]`, `resume
--owner TOKEN` and `status --json` are the pair#722 drain adapter's frozen seam.
The operator or :0 can mutate holds using the landing-grant authority rule.
Owners compose without expiry; a resume removes only its owner and does not
restart interrupted work. Snapshot v1 carries required state/completeness,
pause-owner and activity arrays, plus detail. Active operation IDs can expose
several of suite/merge/landing; waiters are excluded. Unknown/incomplete evidence
never becomes a complete empty snapshot.

`cmd/sdlc/internal/queuegate` owns pure owner/operation transitions, strict
additive JSON codecs and the portable filesystem backend. `queuecmd.go` owns
CLI framing and authority; `queue_legacy.go` is a bounded, read-only status
overlay for old machine locks, pending queue tokens/tickets, landing heads and
same-user merge processes. Held legacy waiters make completeness false with a
named detail, while remaining outside active[]. Verified new waiting bindings
stay excluded and complete.
Process observations never authorize termination. Status-only legacy scans
do not run on the admission path; counted tests protect that separation and
linear registry/binding reads.

The default root is `$XDG_STATE_HOME/ariadne/queue` or
`~/.local/state/ariadne/queue`; `WF_SDLC_QUEUE_ROOT` overrides it. A private
`WF_TEST_LOCK_FILE` alone derives `<file>.queue-state`. Conformance additionally
sets `WF_SDLC_QUEUE_FLEET_ROOT` to an isolated fleet: this scope makes no claim
about production processes. `control.lock` serializes owner writes with active
lease publication; each operation holds its own flock for its lifetime.
Paused ownership is durable even if the caller loses the mutation response.
Unresolved owner/descendant evidence is retained rather than aged into absence.
Interrupted evidence has its own bounded family, outside SuiteAttempt and
failure/flake decisions.

Rollout rebuilds central :0, then lets old runners finish/relaunch; legacy
activity remains named unknown until that boundary is proved. Managed suites
own cancellation through processgroup cleanup. A direct make watcher alone
only proves one PID ended, so an unproved descendant boundary stays unknown
for operator reconciliation. No shared process group is signalled. The
supervisor notice/ack formatter belongs to pair#722, while tl-worker describes
the checkpoint discipline. Its recorded-check continuation/restart rule (#525)
owns safe timing, hub-first context-pressure recovery, and post-restart retained
check verification; the continuation prototype points workers to that rule.

**Reconcile one unresolved operation (#501).** The operator/:0 runs
`sdlc queue reconcile --id ID --digest SHA256 --note TEXT` only after verifying
that the operation has no descendants. `queue_reconcile.go` uses the existing
kernel incarnation probe; `internal/queuegate/reconcile.go` checks exact raw
bytes and operation/lease exclusion under `control.lock`. It appends one
base64-original/digest/actor/note/time record to permanent `reconcile.jsonl`,
fsyncs it, then unlinks only that unknown tombstone. A crash before unlink
leaves a logged attempt; retry by ID/digest finishes without rewriting or
adding audit evidence. Missing/corrupt/changed evidence refuses, and a dead
owner alone never proves descendant absence. Pause owners and interruption
evidence stay intact. There is no log retention/expiry; streaming audit reads
belong only to this rare manual repair and never to admission. Use private
`WF_SDLC_QUEUE_ROOT` fixtures for conformance, never fabricated live residue.

`queue_runtime.go` enrolls each merge before waiting; `Line.tryAdmit` wraps the
actual line transaction with its global gate. Ordinary-worktree merges and
already-merged/train cleanup use the same lifetime through their bypass
admission. `localcheck_run.go` enrolls suites and supplies the exact managed
operation ID to `scripts/test-lock.py`. `QueueGate` borrows that held lease and
publishes its running phase together with weighted capacity; grant-provider
calls remain outside the global mutex. Reentrant makes reuse their holder.
Secondary allocation/line locks are attempted without blocking while the
global gate is held; contention releases the global gate and retries without
discarding FIFO position. Blocking preview/writer work stays outside it.
The forked watcher reopens its control mutex descriptor so it cannot unlock
its parent's transaction. Raw slot/head bindings identify the controller's
operation ID, kernel incarnation and checkout root.

`Lease.RetireOutcome` serializes completion against cancellation, so a
cancellation that wins cannot become a passing suite. The managed owner joins
its cancellation reader and cleans the process group before retirement.
`queue_interruption.go` retains the latest raw interruption log per check/head
as `<git common dir>/sdlc/local-checks/<head>/<check>.interrupted.log` (the normal
newest-20-head retention), while the queue's immutable JSON diagnostics keep
their separate 30-day/4096-record policy. These logs never enter SuiteAttempt,
main-red or flake evidence. A same-head manual retry after resume is ordinary
admission. Direct make has no verified workspace identity and reports it as
unknown, with its checkout named in detail; a surviving child remains unknown.

### Running the test suite (#253)

The required recorded suite is declared in `.sdlc/local-checks.json` as
`make test`, with weight 2 and a 60-minute timeout. It inherits the owner's
trimpath default and existing Go flags. Before that declaration landed, the
equivalent fallback suite command exercised the same candidate Makefile.

`make test` (`scripts/test-shard.py`) is the full run: it compiles the
`cmd/sdlc` test binary once and splits its top-level tests over at most eight
processes (fewer on smaller hosts; four if CPU count is unavailable), balanced
from the last run's per-test times
(`<git-common-dir>/sdlc-test-timings.json`), with every other package in the module
(cmd/weave, pkg/…, the other tools) alongside in one `go test` (#330; before, only
`cmd/sdlc/...` ran, and a cmd/weave hang went unseen).
`cmd/sdlc`'s tests change cwd and swap package seams, so they cannot run in
parallel inside one process; separate processes share neither. Explicit overrides
remain available, for example `make test ARGS="--shards 12"`. Five quiet-start
whole-module measurements on a 12-core M2 Max (#480) found eight shards took
584–600s versus 578s for twelve, with peak one-minute load 29–34 versus 68.
Six shards took 747–820s. These are whole-machine measurements under the
full-suite gate, not a guaranteed load ceiling. Plain `go test ./cmd/sdlc/...` still works but runs
`cmd/sdlc` serially and needs `-timeout 60m` (it exceeds go's 10-minute default).
`scripts/test-timing.py` reports per-test times from either run's `-json` output.

Suites under `sdlc local-check` (the recorded suite, merge's integration suite,
known-flake reruns) write their temp files to the TestScratch RAM disk when it
is mounted and, on Darwin, formatted as APFS (#468, #474): `runUnderSlot` (`localcheck_run.go`) gives each run
`TMPDIR`/`GOTMPDIR` in its own folder under `/Volumes/TestScratch/tmp`, removed
after, and puts the run's checkout under `/Volumes/TestScratch/local-checks`
(review checkouts under `review-checkouts`; `internal/scratchdisk` holds the
rule, and a disk that refuses the write falls back), and falls back to the normal temp folder when the disk is absent or under
2 GiB free (`testscratch.go`). Darwin checks statfs filesystem type even for
`WF_TEST_SCRATCH` overrides: HFS or an unknown type prints a reason and keeps
the normal folder; non-Darwin override fixtures retain their existing behavior.
The check performs no capability writes; cleanup still scans rejected mounted
roots for earlier disposable folders. It also sweeps stale run and go-build folders,
and warns about `t.TempDir` folders a scratched run still left on the SSD temp
folders. `pkg/golint`'s `tmp-literal` flags hardcoded temp roots that bypass it.

Fallback also owns and removes a per-run directory and explicitly overrides
both child temp variables (#477). `internal/scratchdisk.NormalTempDir` probes
actual create/remove access, preserves a usable inherited root, then tries
Darwin's per-user temp and the Unix default independently. Detached-checkout
creation uses the same repair if its os.TempDir fallback is unusable; explicit
custom fallback errors remain strict. Root enumeration also covers all repaired
normal candidates for stale-checkout and per-run cleanup, even after the inherited root
works again. No global process environment is changed.
The local-run dependency can inject the normal-temp selector for fixtures (#487);
its nil default remains the strict production selector. Fallback child-env tests
create their checkout before invalidating inherited temp values and verify real
child writes under a sandbox that denies host normal-temp roots.

`runLocalCheck` reads one bounded complete log for startup classification and
normal suite facts. `suiteenvironment.go` accepts only direct Go work-directory
creation errors and the explicit compile-before-report Python/make wrapper.
Mixed, quoted, truncated, timed-out or unknown output retains ordinary rules.
Recognized startup errors return a typed could-not-run error before flake,
attempt or status publication; known-flake reruns use the same classifier.
Raw evidence moves to `<head>/<check>.environment.log`, bounded by existing
20-head pruning, outside SuiteAttempt and legacy acceptance paths. A retention
failure names and retains the raw inflight log. No schema changes or rewrites:
a same-head environment repair can pass, while genuine failed attempts remain.
Ordinary baseline diagnosis and train `mainIsRed` propagate execution errors;
main health stays unknown and the train emits no failure/split event for them.

Every package whose tests run git calls `testfix.PreferRealGit()` from its
`TestMain`, putting `$(git --exec-path)` first on PATH: on macOS `/usr/bin/git`
is the xcrun shim, which costs about 11 ms per spawn. A new git-driving test
package needs the same `TestMain`
(`TestGitDrivingPackagesPreferRealGit` fails without it). The in-memory git fake that would make a
fast inner-loop tier is #261; the single git seam it and in-process parallelism
need is #262.

A package whose tests push into a local bare origin also calls
`testfix.QuietBackgroundGit()` (cmd/sdlc, gitx, tracker, fleet) and runs the
returned cleanup after `m.Run`. It turns off git's detached auto-gc and
auto-maintenance two ways: through `GIT_CONFIG_*` for every git the binary runs,
and through a `GIT_TEMPLATE_DIR` config for every repository it creates. The
template is needed because git unsets the config environment for the receive
side of a local push (#302). Late writers into a temp repo race `t.TempDir`'s
cleanup ("directory not empty"). The one #302 found was a cancelled git whose
local-transport far end (receive-pack) kept writing. So every context-bound
git runner that pushes or fetches, or otherwise spawns hooks or transports,
kills the whole process group (`processgroup.Terminate`). A read-only `git log`
stream has no far end and is exempt. A new runner of that kind adds one
`testfix.AssertCancelEndsReceiveSide` test.

A test known to fail for a tracked reason is quarantined by name:
`testfix.Quarantine(t, "#N", why)` as its first line skips it with a `QUARANTINED (#N)`
message naming the owning issue, whose fix deletes the call; `WF_TEST_QUARANTINED=run`
runs quarantined tests anyway. Place it at the failing step when only that step is
the known problem (a load-sensitive deadline), so the rest stays covered.
`grep -rn 'testfix.Quarantine(' cmd` lists them (#330).

Under load (several slots' suites at once) spawning a binary or running git is
many times slower, so a test's verdict must not ride on elapsed time. Bound waits
with `testgate.HangCeiling` (`pkg/testgate`, so cmd/weave shares it; cmd/sdlc's
`testfix.HangCeiling` and `hangGuard` alias it): it catches a hang, not a slow machine. Prove ordering with an event instead: `testfix.Watch`
signals when a command's output contains a needle, e.g. `repolock.WaitMessagePrefix`
to fail the moment one command waits for the repo lock behind another (#325).

Cancellation tests prove a descendant stopped with `testfix.AwaitExit(pid, ceiling)`,
which asks the kernel (sysctl `kern.proc.pid` on macOS, `/proc` on Linux) instead of
running `ps`: agent sandboxes deny exec of `/bin/ps` (#330). It waits, because a
group SIGKILL lands asynchronously after the runner returns.

A deadline test must not let a clock race the fixture's startup: a short timeout can
fire before the fixture reports its pid (#434). Hand the code under test
`testfix.DeadlineOnCue(ctx)` and call `expire()` once the fixture has signalled
readiness; the code's own `WithTimeout` then sees a real `DeadlineExceeded`. Where
the deadline lives in a child process, gate the fixture on a FIFO the test releases.

A test that needs a built binary (an end-to-end run of `sdlc`, `weave` or a vocabulary
tool) calls `testfix.BuiltBinary(t, dir, pkg, flags...)`, never `go build` itself (#438;
the `test-go-build` lint rule enforces it). It builds each package and flag set once per
test process, under a 5-minute bound, into an `sdlc-testbin-*` directory under TMPDIR that
the package's TestMain removes with `testfix.RemoveBuiltBinaries()` after `m.Run`; the next
process sweeps any such directory a killed run left behind after 24 h. Measured on the
end-to-end tests: 18 builds (310 MB of binaries) became 2 (20.9 MB).

`TestConstructScripts` asks `scripts/test/construct-scripts.test.sh --list` for
its NUL-delimited selection, preserving the runner's exclusion policy, then
runs each selected script as a named subtest in each color environment (#448).
Each script gets the unchanged two-minute hang ceiling and its own process
group cancellation; completed scripts cannot consume a later script's budget.
The shell runner's default CI invocation still executes the selected batch.

### The Go lint catalog (#407)

Suite text attribution (#476) preserves indentation before parsing controls:
`suiteIndentLevel` classifies the whole leading Unicode-whitespace prefix once,
accepting only ASCII-space quads; `suiteControlLine` derives column-zero
admission for outcomes, summaries and JSON events from that same classifier.
`failingNameLines` admits only column-zero records. Raw indented descendants
coarsen to a proven ancestor, usually the top-level test; JSON failure events
and explicit shard headers still prove exact leaves. `SuiteLogFacts.Warnings`
explains the precision loss and possible subtest-only registration miss.
`flakes.Covers` remains the matching authority; quotations cannot invent leaves.
Quoted nested package trailers cannot flush pending outer names.
The same boundary applies to unattributed failures and timing facts, including
historical logs. `TestFailingNamesLineVisits` counts one line traversal at N and
2N; `TestKnownFlakeGateQuotedNestedOutput` exercises the registered outer retry.

`pkg/golint` is the curated list of mechanical Go mistakes: each rule names the incident
that taught it (`test-landing-completion-wait` #498/#500,
`session-move-batched-reads` and `session-move-raw-fields` #504,
`suite-line-provenance` #476, `test-hardcoded-arch-count` #454, `test-batch-shell-runner` #448, `test-shrunk-deadline` #434, `tmp-literal` #342, `test-go-build` pair#602,
`test-tight-wallclock` and `test-fixed-sleep` the 2026-10-15 flakes, `test-reads-workshop`
#469 (a test reading `workshop/plans` or `workshop/issues`, which move to history on
landing), `wire-strict-decode` pair#594, and `json-recursive-scanner` #489, a hand-rolled recursive duplicate-key
JSON scanner outside `cmd/sdlc/internal/jsonstrict`, the one owner of that rule). `TestAtlasNamesEveryRule` keeps this list
in step with `Catalog`. Packages are matched by the
name the file imports them under, and generated files are skipped. Rules read only the syntax tree (stdlib
`go/parser`, no type-checking, no external linter), and `TestTrackedGoFilesPassTheCatalog`
runs them over every tracked Go file in the suite, so every landing enforces them (well under
a second over about 1,140 files). A finding that doesn't apply is suppressed on its line or the line
above with `//lint:allow <rule> — <reason>`; a suppression without a reason, naming an
unknown rule, or matching nothing is itself a finding. A fix for a mechanical mistake class
adds its rule here (the shared planning block in `tl-worker`/`tl-lead` help says so).

Mark a named type with `// wire: <why it can evolve separately>` in its type-doc
comment when it receives JSON from a peer binary, agent or toolchain that may
add fields before this binary upgrades. The lint scans all tracked Go sources
once to index marked types, then flags a direct `json.NewDecoder` →
`DisallowUnknownFields` → `Decode` chain into one. Direct calls in expression
statements, assignment RHS, `var` initializers, returns, and assignment-form
`if` initializers are covered. Expression-form `if` initializers are tracked
in #445. It reports at `Decode`;
ambiguous branches, goto-bearing functions, and helper decoders are outside this conservative
syntax-only rule. A marked target resolves in the same package, the module
root, or an imported subpackage when its declared package name or explicit
alias matches the selector; a shadowed qualifier is not an import. A decoder remains
proven only if its function-local binding is declared by its sole direct
`json.NewDecoder` assignment, with no reassignment or address or closure escape,
and no use outside the two direct method calls. Copying it or passing it to a
helper disqualifies the name, as does a second assignment even when the new
value is another decoder. A package-global decoder is outside the proof
because a helper can replace it. This can miss a strict decode after aliasing, but
cannot guess that the decoder remains strict. The current examples
are Claude stream events, `go test -json` events, and Codex/Claude usage
records (#440). Do not mark local configuration, tracker receipts, fleet policy,
or version-pinned issue observations: their unknown keys are errors, so strict
decoding remains correct. A real exception at a flagged decode uses the normal
reasoned `//lint:allow wire-strict-decode — <why>` on that line or the one above.

Follow-ups: golangci-lint for off-the-shelf linters (#439, operator), one build
per test binary (#438).

### Two full suites per machine (#319, #341)

The rule: a full `make test` takes one of `WF_TEST_SUITES` (default 2) full-suite
slots; a targeted run takes one of `WF_TEST_GATE_SLOTS` test-gate slots and never
waits for a full suite. The two caps are independent. #319 allowed one suite and
held targeted runs back behind it; at load ~2 with five slots waiting on each
other's suites, that wasted the machine, so the operator raised it (#341).

A full `make test` in any woven repo (ariadne, pair, parley.nvim, …) first takes
a machine-wide slot, so suites from parallel slots and repos queue instead of
oversubscribing the cores. `Makefile.workflow` takes it at parse time
whenever `test` is a goal, through ariadne's `scripts/test-lock.py`, resolved
beside `Makefile.workflow`'s real path (so a woven repo needs no copy, and a
missing helper warns that the run is unlocked):
`flock` on one of `${XDG_STATE_HOME:-~/.local/state}/ariadne/test-suite.lock`
(slot 0) and `test-suite.lock.<k>`, held by a forked watcher until that `make`
exits, so a crash or kill releases it at once. Slot 0 keeps #319's filename, so an
older `test-lock.py` (a checkout not yet refreshed) still counts against the cap.
Parse time matters: a prerequisite would run after the test target's own
prerequisites, which are pair's whole suite. Targeted runs (`go test -run …`, a
single `make test-*` target) never take it; they go through the test gate below. A nested `make test` under a locked
run passes through (`WF_TEST_LOCK_HELD_BY`), even while another slot is free.

**`sdlc`'s own suite runs hold the slot for the run only (#383).** A local check
(`sdlc local-check run`, and the suite `sdlc merge` runs) does not give the
helper its own pid to watch. It starts a per-run holder, a `cat` on a pipe from
`sdlc`, and passes that pid as `--watch`. Right after the suite's process group
ends, it closes the pipe and reaps the holder. The helper's watcher frees the
slot on that exit (at once on macOS's kqueue, within its 0.5 s poll elsewhere),
so the status post, the CI poll, the PR merge, archiving and cleanup hold no slot. If `sdlc` dies,
the kernel closes the pipe, and the slot is freed as well. A rerun in the same
`sdlc` (a re-integrated head, the baseline diagnosis) starts a new holder and
takes a ticket like any other waiter. The suite's `WF_TEST_LOCK_HELD_BY` is the
holder's pid, so its nested `make test` passes through.

Queued cancellation follows the same lifecycle (#416). The runner owns
SIGINT/SIGTERM shutdown only during queue acquisition and suite execution:
it cancels its helper group and reaps the holder, restoring default signal
handling before result bookkeeping. Git/network preparation retains default
termination as well. Detached checkout cleanup uses independent bounded Git
recovery IO. If the runner dies without unwinding,
the holder's pipe closes; the queued helper checks its watched PID before
admission and on each one-second queue poll, exits when it is gone, and
releases its own ticket. A killed helper's unlocked ticket is pruned by the
next waiter. Neither path discovers or stops another run (#400).

**Weighted suites (#346).** A suite heavier than one slot declares its weight:
`WF_TEST_WEIGHT ?= N` in the repo's root Makefile, before `Makefile.workflow`
is included. The default is 1. ariadne retains 2 with its eight-shard cap (#480):
the measured suites still occupied about 78–90% of whole-machine CPU. Taking
both slots keeps two full suites from overlapping; the comparison did not test
concurrent suites. Earlier overlapping suites had pushed load to 185 (#346).

**Host preflight before the lock (#333).** A repo whose suite cannot run on some
hosts ships a preflight, by convention `tests/host-preflight.sh`, or names another
with `WF_TEST_PREFLIGHT ?=` before including `Makefile.workflow`. When `test` is a
goal it runs at parse time, before the lock: a failing preflight stops `make test`
at once with its own message ("not taking the full-suite lock"), so a sandboxed run
never queues behind other suites first. Its output goes to stderr; only its exit
status is read. pair ships one (sockets, PTYs, exec of `ps`). ariadne has none: #330
made its suite pass in the agent sandbox, so the same probes would block runs that
work.

`Makefile.workflow` passes the weight as `test-lock.py acquire --slots N`,
clamped to `WF_TEST_SUITES`. The acquisition rules:
- **All or none.** The N slots are taken in one scan or not at all. A waiter
  holds no slot, so two heavy waiters cannot deadlock.
- **First in, first out (#357).** Each waiter takes a numbered ticket in
  `test-suite.lock.tickets/` and holds a flock on it while it waits; the lowest
  live ticket is the oldest waiter. A ticket whose lock the kernel released (its
  waiter died) is pruned, so a crash never blocks the line. Tickets are created and
  pruned under a short allocation lock (`alloc`), so a new ticket is never mistaken
  for a dead one.
- **Grant priority (#452).** A suite for an issue holding a currently valid
  fleet-blocking landing grant precedes waiting normal suites. FIFO holds within
  each class; no running suite is preempted. Expiry or revocation returns the
  waiter to normal FIFO at its original ticket number. A priority head needing
  two slots waits for both; light normal suites cannot backfill ahead of it.
- **One elected head.** Under `alloc`, every attempt reevaluates grant evidence,
  elects the head, takes `test-suite.lock.queue` and scans weighted capacity.
  A failed scan releases the queue token, so a normal waiter cannot reserve it
  against a later grant. The allocation counter stays numeric, independent of
  effective order. A stopped live waiter keeps its FIFO position within its class.
  Old helpers already holding the sticky queue token drain naturally; never kill
  or unlock them for rollout. Next-free priority is guaranteed after those old
  waiting helpers have drained and participating helpers are upgraded.
- **Evidence and failure.** Recorded suite identity is the original checkout's
  issue branch only when its HEAD equals the tested commit. Baselines, other
  checks and unknown identities remain normal. Direct make uses its supplying
  workflow's `bin/sdlc`; recorded checks use their own executable. Tickets carry
  repository/issue identity, never a priority boolean or executable. Snapshots
  live only in memory, invalidate on source changes or Go-derived activation/
  expiry boundaries, and start no provider process on unchanged pre-boundary
  polls. Provider calls are bounded to 2 seconds and 64 KiB. Missing, old,
  unreadable or malformed evidence emits `priority unavailable` and falls back
  to normal FIFO for every suite, including known issues; capacity still holds.
- **Records.** The holder's record goes into each slot it holds, with
  `slots: [...]`.

The wait names each holder once, then `position K in line; next in line: …` (the
waiter ahead), or `position 1 in line: next to run` for the head.
It asks for `2 of 2 full-suite slots` for a weight-2 run. The wait also names
normal FIFO or fleet-blocking issue/deadline, and reports class/position changes
immediately rather than waiting for the periodic diagnostic.
If contention or a pause prevents validating order, it reports `position unknown`
until an observation is available, then explicitly labels reused diagnostics
`last observed`. Missing order is never treated as zero predecessors. Switching
between a validated order and its cached diagnostic alone does not repeat the
log; periodic reports and fresh pause reasons still apply (#491).

While waiting it prints `[test-lock] waiting Ns for one of N full-suite slots …;
held by:` and one `<repo> (<worktree>), make pid P, since T` line per holder,
repeated every 30s. `WF_TEST_LOCK=off` skips the lock
(loudly). `WF_TEST_LOCK_TIMEOUT=<seconds>` bounds the wait and then fails make
with "full-suite lock not taken (timeout)" — a lock wait, not a test failure.
Agents running the pre-close suite should run `make test` in the background so a
wait can't trip the tool's timeout. Tests: `scripts/test/test-lock.test.sh`, which
the `20-ci-workflow-tests` merge check runs (FIFO order under a stopped waiter, a
dead waiter skipped, an older copy's head going first).
`TestSuiteQueuePriorityRealHelper` additionally drives the real Python helper
and production Go grant provider with private lock/grant files; it covers grant
lifecycle, both FIFO classes, weight, provider failure and direct make binding.

### Targeted runs share a machine budget (#324)

Heavy test packages call `testgate.Enter()` (`pkg/testgate`) first in their
`TestMain`, so every run of their tests is gated, whatever the harness, shell or
flags:
- **Covered packages:** every git-driving `cmd/sdlc` package
  (`TestGitDrivingPackagesPreferRealGit` enforces it), plus `cmd/weave` and weave's
  `refresh` and `acquire`.
- **Never behind a full suite (#341):** the gate ignores the full-suite slots.
  A targeted run is small next to a suite, and full suites have their own cap.
- **Slots:** it takes one of `WF_TEST_GATE_SLOTS` slots (default NCPU/4),
  `test-gate.<k>.lock` beside `test-suite.lock`, each holding its holder's record,
  until the binary exits.
- **Parallel cap:** it caps `-test.parallel` at `WF_TEST_PARALLEL` (default 2)
  unless the run set it.
- **Wait budget:** the wait spends the run's `-test.timeout`, because `cmd/go` kills
  a binary at that timeout plus a minute from its start, however long it queued.
  So the gate waits at most a quarter of it, then goes ahead without a slot
  (loudly), and shrinks `-test.timeout` by the wait so the binary's own alarm, which
  names the running tests, fires first. The measurement caught this: runs queued
  10+ minutes behind a pair suite were SIGQUIT-killed.
- **Hatches:** the same as #319's lock. `WF_TEST_LOCK=off` skips the gate loudly,
  and `WF_TEST_LOCK_TIMEOUT` gives up as "a lock wait, not a test failure".
- **Under `make test`:** `WF_TEST_LOCK_HELD_BY` makes the gate stand aside, because
  the suite has its own lock and shards.
- **Nested helpers (#330):** `Enter` exports `WF_TEST_GATE_HELD_BY=<pid>`, so a test
  binary the gated one starts (a re-exec'd `os.Args[0] -test.run=…` helper) stands
  aside under its ancestor's slot. Gated again, it would queue behind its own
  parent's slot with no `-test.timeout`, so with no wait budget, and hang (cmd/weave's
  `TestCompileRecoversAfterProcessDeathDuringGeneration` did, for 22 minutes).

**Waits are invisible on a pass.** In package-list mode (`go test ./pkg/…`), `go`
drops a passing binary's output, so the `[test-gate] waiting … held by …` lines
show only with `-v`, on a failure, or in local-directory mode. A run that looks
hung is usually queued: `cat ~/.local/state/ariadne/test-gate.*.lock` names the
holders (`test-suite.lock*` name the running full suites). Run long targeted
batches in the background, like the full suite.

Why in the binary and not the slot env: env is fixed when a harness launches and
differs per harness (pair#203), and a `go -exec` shim needs an executable on PATH,
which slots don't have.

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
both `push` and the legacy `merge` path. Durable landing archives remotely
through `landingarchive.go` instead. The legacy archive sweep also moves the issue's
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

Most Make targets delegate to `bin/sdlc` when built and retain shell fallbacks.
The `issue-sync` wrapper always uses Go: if needed, it compiles a temporary SDLC
binary from the workflow source directory and invokes it from the consumer cwd.

  `make close-issue` → `sdlc close`
  `make fetch <N>`   → `sdlc fetch --github-issue N` (deprecated alias →
                       `sdlc issue new --from-github N`, #56 M2)
  `make worktree`    → `sdlc change-code --worktree=yes --no-judge --no-structural
                       --no-estimate` (post-#39, #113; preserves the make target's
                       pre-existing quick-and-dirty gate-free semantics)
  `make issue-sync ISSUE=N` → `sdlc issue sync --issue N` (local checkpoint only)
  `make push`        → `sdlc push`
  `make pull-request` → `sdlc pr`
  `make merge`       → `sdlc merge`
  `make check-<cat>` → `sdlc judge <cat>`

The remaining shell fallbacks serve downstream repos that vendor
`Makefile.workflow` before building SDLC. There is no separate shell publication
implementation behind `issue-sync`.

## When to add a new verb

The rule from the pensive: **when the same drift gets caught at review
twice**, promote it from a `workshop/lessons.md` entry to an `sdlc <verb>`
check. The first time → prose. The second time → code.

Do not formalize the workflow into a state machine. Add checkpoint
guards for known commit moments where drift recurs; everything between
checkpoints stays prose-driven.

## Workspace identity (#242)

`workspace [address] --json` exposes the shared `pkg/workspace` identity contract.
`state` includes the same workspace object. Repository labels and fleet paths
derive from the primary/common Git directory; local artifacts remain in the
current worktree. See [Workspace Identity](workspace-identity.md) for address,
resting-branch, calibration and project-discovery semantics.

## Local usage (#365, #385)

`sdlc usage` reads bounded local Claude/Codex/Qoder response evidence. Its first rows
show fleet native usage in the last hour, wall-clock burn over 1–3 hours, and an
empirical account projection when two compatible meter readings and complete
spend evidence support it. `--since` selects detailed groups/hourly totals;
`--json` exposes schema version 2. Caches are input subsets and reasoning is an
output subset, so totals never add them twice. Codex normalization admits usage
only after resolving the immutable executing identity; pre-identity usage is
skipped and makes spend coverage incomplete even when a valid header arrives
later. Missing cwd separately affects attribution.

Qoder reports per-response credits; its zero token placeholders stay null.
Unexpected nonzero or invalid token placeholders produce visible format coverage
and block account projections while preserving independently valid credits.
Credits use exact bounded decimal arithmetic, are deduplicated by message.id,
and never include original_credits a second time. Reports carry parallel token
and credit columns; a mixed scope has null complete totals for unsupported
units and separately labelled known subtotals. No conversion between tokens and
credits is inferred. `internal/qodertranscript` supplies the shared typed
metadata/configuration interpretation for usage, actuals and agent identity.
Discovery reads `~/.qoder/projects/*/*.jsonl`, matches recorded
workspace-directories rather than project slugs, and preserves immutable session
identity across replays. Actuals/close use the existing ownership and pinned
prefix evidence path. Failed Qoder sources with a recorded requested workspace
block automatic actual adoption. Unclassifiable files, including native
session/subagents/agent-*.jsonl sidechains without workspace metadata, are skipped
with path/reason warnings rather than blocking unrelated issues. Identity scans
only top-level sessions. Usage retains unbound sources in its unknown-slot
coverage; skipped actual sources are not attributed by directory-name guesses.

Agent identity checks QODER_CLI before competing native-agent markers; Qoder
also exports CLAUDE_PLUGIN_* variables, which do not prove Claude invocation.
When present, Pair's launch-pinned PAIR_SESSION_ID must match the transcript
sessionId and recorded cwd. Without that binding, competing sessions remain
qoder/-/-; recency never chooses another session's model. SDLC_AGENT_ID remains
the explicit ambiguity fallback. Qoder transcript recognition does not add a
Qoder gate-reviewer CLI.

`--issue N` reuses ownership intervals and bound descendants from actual-time
attribution, but includes the review/verification tail and ends completed issues
at their exact completion evidence commit. Reviewer binding has precedence over
worker eligibility: separate worker, reviewer (rounds/attempts), and total lines
count each response once. `actual` remains an independent active-time metric
with its own right edge. The command reads the local tracker snapshot without
refreshing it; provenance names the pin. Unmapped spend is visible in
unknown-slot, and legacy missing review bindings stay coverage gaps.

`internal/usage` owns pure response normalization, deduplication, window/rate
aggregation and meter transitions. `internal/transcripts/usage_discovery.go`
adds usage-only discovery of explicitly dispatched Claude children; actual's
existing flat Claude discovery remains unchanged. `usage_inputs.go` applies
Git-verified fleet membership and shared ownership; `usage_reviewers.go` joins
recorded dispatch bindings, never guesses by repository overlap.

`internal/reviewusage` records accounting-only Requested → Started → Finished
receipts in the Git common directory's closed review-usage CAS namespace. The
request persists before launch; cancellation uses a bounded cleanup context.
An unfinished receipt is uncertainty, not evidence that the provider spent
nothing. Canonical attempts need stable, complete bound response evidence or an
explicit unresolved_reviewer_usage spend gap. These gaps follow issue/report
windows and each agent's burn/calibration windows; confirmed launch failures are
nonspending. Typed copies enter round ledgers and sidecars after safe lock
reacquisition; accounting never advances a review gate or permits a retry.

`sdlc usage calibrate --agent codex --remaining 29276 --unit credits` writes the
private bounded meter under XDG_STATE_HOME (default ~/.local/state). Two readings
of the same agent/unit calibrate meter change against local tokens between them
(reported native credits for Qoder);
refills restart calibration. One reading, stale/incomplete evidence or zero burn
produces a labelled status and null projection. Plain usage has no cache and
never interacts with the providers. Re-run opt-in identity conformance when the
agent CLIs change and before bench runs; see the usage help for coverage bounds.


### Landing-check capability (#509)

The opt-in version2 local-check declaration separates `scope: landing` from internal affected-test coverage and `resource: runner` from suite-slot weight. Its schema lives in `localchecks_v2.go`; strict known members coexist with preserved opaque future metadata. Existing v1 declarations retain their behavior. `landingchecks.go` selects enabled checks from per-car path unions and binds in-memory proofs to candidate/policy/car identity; `landingchecks_git.go` reads only regular tracked blobs on the configured base and caches immutable input reads for the landing owner's lifetime.

Runner-resource execution uses the existing exact-head worktree, managed queue/run lease, logs and attempt store, while bypassing suite-helper discovery and allocation. Its entire subprocess group gets one30s graceful cancellation budget, including after the immediate Make leader exits; forced cleanup then joins before checkout retirement. No VM lease is acquired or inspected by SDLC. Automatic train/single-landing enforcement is the next milestone; no consumer is activated by this capability issue.

For unconfirmed runner cleanup, the existing queue operation is left unknown and raw logs and Git-locked worktree inputs remain for operator reconciliation. Stale sweeps resolve each candidate's own Git lock, including scratch directories belonging to another repository. Signal success alone never authorizes retirement; escalation is followed by bounded completion confirmation.

The shared predicate in `localcheck_resource.go` protects all retained artifacts through bounded owner references inside existing scratch/in-flight directories. Both scratch's age/PID sweep and the attempt store's dead-owner/temp collectors consult it; missing/unreadable authority retains evidence. Runner launch requires an attempt writer, while suite-resource legacy fallback remains. Unknown outcomes retain everything until operator reconciliation: prove descendant absence, reconcile the queue, remove named scratch/log directories, then unlock/remove the worktree. Git metadata stays until linked artifacts are reconciled. If unknown endings are frequent, retention policy is an explicit future operator decision.

The Go lint catalog includes `test-sweep-private-root` (#509): test calls to the destructive temp collector require visible testing.TempDir provenance, preserved through safe literal relative joins. Runtime roots, parent traversal, uncertain aliases/reassignments and address escapes refuse the source check. This prevents a fake process-liveness probe from deleting another run's scratch folder.

Landing trains enter `PhaseChecking` after the successful final-prefix suite (also after docs reintegration). `RunLandingChecks` produces typed `LandingChecksDone`; only a current-prefix pass permits `WaitCI`. Check failure refuses that prefix without suite bisection or baseline diagnosis.

`landingchecks_gate.go` owns automatic check proof and invocation-local attempt memory. `conductor.execute` runs it for the exact kept final prefix; tracked `runDurableMerge` and legacy `runMerge` call the same gate after per-head requirements. Pinned original car heads, policy digests and retained log hashes bind proof. Docs reuse requires ancestor/code-surface proof and no watched new path; issue-only fast merges apply the same boundary. `headLocalChecks` projects per-head requirements without hiding red statuses. Manual-only landing entries require pinned main and explicit CLI naming; runner execution uses no suite capacity. No consumer is activated by this capability.


### Reader-first attempt compatibility (#524, prerequisite for #463)

`teststats_evidence.go` projects the bounded attempt store into supported runs,
strictly identified unsupported proof, and unscopable read warnings. A future
record retains its original bytes and cannot provide outcomes or measurements.
Saved-proof selection, republish and success guards report “needs a newer sdlc”
for its exact ID, subject or linked parent; unrelated repo/head/named-check
proof remains usable. Suite and affected share their existing same-head test
failure authority. Unreadable, malformed, ambiguous or budget-truncated input
still prevents proof of completeness. Statistics and actuals keep supported
measurements with explicit unavailable-record coverage; flake registration
cannot use unsupported authority or substitute a legacy log for it.

**Activation barrier:** pre-#524 binaries treat a v3 record as a global invalid
record. A payload cannot retrofit that old reader. This landing writes no v3
record or ResumeLink and enables no resume command. After it lands the TL must
upgrade every slot's actual sdlc; verify the binary path and build revision
across slots before #463's separate writer/activation landing. A stale or
unverified reader keeps activation held. Never hide links from older readers.
The later resume design must also name one real recovery command for crashed
pending proof; queue reconciliation alone does not repair a resume link.

The retained compatibility control is opt-in: run
`sh scripts/verify-pre524-reader.sh`. It creates a private clone at the pinned
pre-#524 revision and executes the committed old-reader fixture, asserting the
old global warning, incomplete saved proof and publication refusal. Its printed
private path is retained for inspection. The upgraded comparison is
`go test ./cmd/sdlc -run '^TestScopedSavedProofConsumers$' -count=1`.

The invoker owns that private compatibility-control directory, retained only
through evidence inspection. Afterwards remove the printed parent directory
(`ariadne-pre524-reader.*`), including its clone; never sweep the shared scratch
root. It contains no live lease or production evidence and is not a long-term
artifact store. The committed fixture/script and issue Log retain the proof.


### Exact-package known-flake batches (#524)

`internal/flakes.RerunCommands` groups validated names by package and promotes
registered subtests to whole top-level tests. Each sorted literal argv runs
one package with anchored names and `go test -json -count=1`. A name from a
different package cannot select a same-name decoy. `RerunMissingJSON` requires
package-bound exact-test, whole-top-level and package PASS events; Output text, malformed
or conflicting terminal events cannot authorize success. Existing saved
supported evidence remains readable.

`knownFlakeStep` runs the batch through `runUnderSlotBatch` and the shared
`runManagedCommandBatch` execution seam. One queue admission, weight-1 slot,
run lease, scratch directory, raw log and aggregate targeted timeout span all
package processes. Each group is stopped before the next starts; failure,
timeout or cancellation stops the batch. Argv announcements go to terminal diagnostics with quoted embedded
newlines; the raw log contains only process output, so command data cannot
become log facts or interfere with startup-error classification. The result is one immutable
`suite-rerun` attempt linked to the original failed suite/affected attempt,
never another full-suite sample. Runner-resource checks retain their existing
single-command cleanup and retention authority.


Executable conformance is opt-in:
`SDLC_PRIVATE_CONFORMANCE=1 go test ./cmd/sdlc -run '^TestLocalCheckPhase1Conformance$' -count=1 -v`.
The subprocess enters production main with private real Git/tracker/helper,
private queue/lock/scratch/evidence, and a stateful fake GitHub publisher.
It proves unrelated named-check republish beside a future record, related
refusal with zero publication, exact-package decoy exclusion, unchanged opaque
bytes, and one retained original/linked rerun pair. The test owns and removes
all fixture state via `t.TempDir`; no installed binary or live evidence changes.


PASS event identity follows actual Go output, separately from the registry's
narrow eligible-name grammar. Incidental legal subtests (punctuation, Unicode,
rewritten spaces, empty/duplicate segments) cannot invalidate a required root's
proof. One t.Run may introduce a slash-containing suffix, so slash prefixes
are not inferred ancestry and cannot require invented events. Exact required
leaf, whole top-level test and package must pass, preserving sibling-failure
blocking and required-identity terminal conflict refusal.

### Immutable resume proof core (#463 M1)

`localcheck_resume_evidence.go` defines strict v3 resume links and terminal
interruption receipts, a pure `resolveResumeProof` and the execution/recovery
transition table. Links fence one saved suite parent to one reserved rerun child;
they are coordination evidence and never count as suite samples. The common
`SuiteEvidenceSnapshot` indexes them alongside real attempts and unsupported
proof. Recognizing a pending link alone supplies no passing authority.

`localcheck_resume_store.go` holds the immutable parent inode lock across
recovery, verifies exact raw parent bytes and publishes synced proof with an
exclusive link. `localcheck_resume_retention.go` pins pending bundles, validates
terminal bundle age and retires under that lock: selectable parent unlink and
successful directory sync first, ordinary logs/inflight next, link removal and
sync next, terminal child/log or interruption receipt last. Crashes retain either
the fence or an unselectable parent with a terminal witness for cleanup retry.
Malformed/unreadable evidence retains data. Shared members/logs stay protected.
The existing pruner skips all indexed bundle members during individual-file
cleanup. This core milestone adds no user command or automatic link writer;
CLI activation and all-reader composite proof are the following milestone.

One flock on the existing attempts directory spans the protected-member snapshot
and both bundle and legacy pruning. Publication acquires parent, then directory;
pruning acquires directory, then **strictly nonblocking** parent locks, skipping
busy parents. The inverse order is safe only because pruning never waits on a
parent. Real subprocess barrier tests deadlock under a deliberate blocking-lock
mutation, and verify owner-death release and publication versus old snapshots.
Directory path/inode identity is revalidated; no new lock files or per-parent
snapshot scans are added.
