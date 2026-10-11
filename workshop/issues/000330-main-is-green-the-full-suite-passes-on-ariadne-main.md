---
id: 000330
status: codecomplete
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-11
estimate_hours:
card_mirror: 'c988e891d18f8360b73281bbb73ea7ca9c874008' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T23:14:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.60
---

# Main is green: the full suite passes on ariadne main

## Problem

ariadne has no CI on main, so nothing shows whether the full suite passes there. During batch 2 (10-10) workers judged
failures by "it also fails on main", and full suites flaked under load (ops kink 53). The operator (batch-2 recap,
10-10): "let's ensure ariadne and pair are green on main".

## Spec

- Run the full suite (`make test`) on a clean main; list every failure, every flake (run it twice), and the slowest
  packages.
- Each failure or flake: fix it at the root, or quarantine it by name with an owning issue (the exception, named in
  the Log). Overlaps the test-infra train (ariadne#325 deadline tests, #302 tempdir cleanup race): fold or point,
  don't duplicate.
- Record the suite's wall time and shard count in the Log (input for the shard-count question from the recap).
- Decide whether ariadne gets a CI job on main; if yes, add it (or file it).

## Done when

- [x] `make test` passes on main from a clean slot, twice in a row, with wall time recorded.
- [x] Every quarantined test has an owning issue.
- [x] The CI decision is recorded, and the job exists or is filed.

## Plan

- [x] Baseline: `make test` ×2 on the issue branch (= main + this body), plus the Go packages `make test` never runs.
- [x] Sandbox failure: replace the three `ps -o stat=` probes (processgroup, judge ×2) with one exec-free helper in `testfix` (`AwaitExit`).
- [x] Fix `processAlive` treating EPERM as dead (repo lock reclaimed a live peer's lock under the sandbox).
- [x] Load-sensitive deadline failures: quarantined by name under #325 (`testfix.Quarantine`), no duplicate fix here.
- [x] Coverage gap: `make test` runs only `cmd/sdlc/...` → its `packages` job now runs every module package.
- [x] Fix the testgate self-deadlock the widened run exposed (nested helper binaries).
- [x] CI decision recorded; job filed as #331.
- [x] `make test` green ×2 with wall time + shard count in the Log.

## Log

### 2026-10-10
- Claimed from ariadne:3 on the TL's dispatch (ops:0). Load before the run: 7–8 (pair:4 suite holding the machine lock).
- Run A (`make test`, 12 shards over 1047 cmd/sdlc tests + 20 other packages): waited 177s for the lock behind pair:4's
  suite, then wall 661.8s (real 839.6s with the wait). Load peaked ~58 (my own concurrent targeted run of the non-sdlc
  packages added to it). 3 failures, none a regression:
  - `processgroup.TestCancellationKillsDescendants`: `fork/exec /bin/ps: operation not permitted`. Claude Code's
    Seatbelt sandbox denies exec of `ps`, so it fails in every sandboxed agent slot. The two judge
    cancellation tests use the same `ps -o stat=` probe but ignore its error, so under the sandbox their
    "descendant still running" check silently passes. Root fix below (exec-free kernel query).
  - `TestPlanningReviewConcurrencySchedules` (plan + estimate: "unrelated mutation could not finish ... signal:
    killed") — exactly #325.
  - `TestReviewConcurrencySchedules/milestone-close/branch`: "reviewer process did not start" after 22s — same
    load-sensitive deadline family; recorded on #325 rather than fixed here.
- `make test` covers only `go list ./cmd/sdlc/...`. cmd/weave (14 pkgs), pkg/* (6), cmd/datatype, cmd/doc-review,
  cmd/vocabulary have no suite entry point; run separately here.
- Non-sdlc packages, run once by hand (`go test -count=1 -timeout 30m`, gated): all pass except **cmd/weave, which
  hung**: `TestCompileRecoversAfterProcessDeathDuringGeneration` sat 22m18s in `CombinedOutput` on its re-exec'd
  helper until the 30m timeout. Root cause in `pkg/testgate` (#324): the helper is the same test binary, so its
  `TestMain` calls `Enter()` again; it carries no `-test.timeout`, so its wait budget is 0 = unbounded, and it queues
  behind full suites and slots, including the slot its own parent holds (reproduced with `WF_TEST_GATE_SLOTS=1`:
  the child waits on the parent's pid). Under `make test` `WF_TEST_LOCK_HELD_BY` bypasses the gate, so only targeted
  runs (agents' inner loop) hang. Fix: `Enter` exports `WF_TEST_GATE_HELD_BY=<pid>`; a descendant stands aside under
  its ancestor's slot. New `TestNestedHelperStandsAside` failed before (child killed at 10s, waiting on the parent),
  passes after; the weave test then passes (151s, of which ~150s was the gate's quarter-timeout wait behind a suite).
  Other re-exec helpers covered by the same fix: cmd/weave generator_lifetime, staging, weavefs, release, plan;
  cmd/sdlc cli_signal, judge cancellation.
- `ps` probes → `testfix.AwaitExit(pid, 2s)`: kernel query (sysctl `kern.proc.pid` via `golang.org/x/sys/unix` on
  darwin, `/proc/<pid>/stat` on linux, `ps` fallback elsewhere). Making the probe actually work exposed a second
  latent bug: the two judge checks failed immediately, because a group SIGKILL lands asynchronously after `Run`
  returns and the old `ps` spawn (~10ms) hid that race. So it is a condition wait with a ceiling (#325's pattern),
  not an instant probe. Mutation check: `Kill(pid)` instead of `Kill(-pgid)` in processgroup fails all three sites
  ("pid N still running after 2s"), sandboxed — the judge ones used to pass silently there.
- New dependency `golang.org/x/sys` (v0.47.0, test-only use): the stdlib exposes no kinfo_proc sysctl, and
  exec-free inspection is the point.
- `make test` widened from `go list ./cmd/sdlc/...` (21 pkgs) to `go list ./...` minus the sharded cmd/sdlc
  package: the cmd/weave hang above was invisible to every "green" suite. Cost measured below.
- Run B (`make test`, tree with the fixes above + the widened package list): waited 1452s for the lock (pair:4,
  ariadne:4, ariadne:1 suites ahead), then **wall 487.0s, 12 shards over 1047 cmd/sdlc tests + 44 other packages**
  (real 1940s with the wait). The widened `packages` job (now incl. cmd/weave, pkg/*) passed in 232.6s, inside the
  shard wall: whole-module coverage costs no wall time at 12 shards. Failures:
  - `TestPlanningReviewConcurrencySchedules` again (#325).
  - shard11: the hermeticity guard (#149/#165) reported "leaked .git/sdlc.lock" in a passing shard. Root cause is a
    real lock-safety bug, not a test leak: `processAlive` was `kill(pid, 0) == nil`, and from inside Claude Code's
    Seatbelt sandbox `kill -0` on any process outside it answers EPERM (verified: a live parley.nvim `make` pid →
    "operation not permitted"). So a peer slot's live sdlc holding the shared `.git/sdlc.lock` (the common dir is
    shared by all slots) read as dead: the guard counted it as ours, and in production `repolock.Acquire`
    renames a "dead" holder's lock away and takes it — two slots' verbs inside one transaction lock. Fix: EPERM
    counts as alive (ESRCH is the only "gone"); `TestProcessAliveCountsUnsignalableProcesses` (pid 1) failed
    before, passes after. The other liveness checks were already right: weave staging tests `!ESRCH`,
    `test-lock.py` treats `PermissionError` as alive.
- CI decision: yes, main should run the full suite in CI, but not added here. The suite has never run on Linux
  (darwin-tuned helpers), and a 4-core hosted runner would run ~3× the 12-shard wall. Filed as #331.
- Quarantine (the spec's named exception), owner #325, both unowned at the time:
  - `TestPlanningReviewConcurrencySchedules`: failed in 3 of 3 full runs here (runs A, B, and the TL-reported 4/4).
  - `TestReviewConcurrencySchedules`: failed in 1 of 2 (milestone-close/branch, "reviewer process did not start").
  Mechanism: new `testfix.Quarantine(t, issue, why)`, a loud `t.Skipf` naming the issue; `WF_TEST_QUARANTINED=run`
  runs them. #325's fix must delete both calls (reported to the TL to fold into #325's Done when).
- #302 (TempDir cleanup race): not seen in any run here; nothing to fold.

### 2026-10-11
- 2026-10-11: closed — make test green twice back to back on b4b918a3 (run C wall 469.0s, run D wall 448.8s; 12 shards, 1048 cmd/sdlc tests + 44 other packages, now whole-module). Fixes: testgate nested-helper deadlock (TestNestedHelperStandsAside red→green; cmd/weave hung test passes), repolock processAlive only-ESRCH-is-gone (TestProcessAliveCountsUnsignalableProcesses red→green), exec-free AwaitExit replacing ps probes (mutation: non-group kill fails all 3 sites). #325 quarantines owned by #325 Done-when; CI filed as #331. Round-1 fixes re-verified: TestReviewConcurrencySchedules + TestProcessAlive pass targeted.; review verdict: SHIP
- 2026-10-11: flow upgraded quick → full — 148 added lines in code files (limit 100); an earlier round of this close already ran the full review
- **Green ×2, back to back**, on the final tree (b4b918a3 = main + this branch), `make test` from ariadne:3:
  - Run C: lock wait 1134s (pair:5, ariadne:2 suites ahead), then **wall 469.0s**, exit 0.
  - Run D: no wait, **wall 448.8s**, exit 0 (load ~16–28 at start).
  - Both: **12 shards** (one per core) over 1048 cmd/sdlc tests + 44 other packages, 2 quarantined (#325).
- Shard-count input for the recap: at 12 shards the cmd/sdlc shards finish at 400–480s each, and the whole-module
  `packages` job (~230s) fits inside them. The wall is set by the slowest shard, so fewer shards cost roughly
  linearly. The "about 4 minutes" in the atlas predates the current test count and the load from parallel slots.
  The suite lock queue costs more than the suite: 177–1452s of waiting per run tonight.

- Close review round 1: FIX-THEN-SHIP, 1 Important + 3 Minor, all fixed in round: #325's Done when now requires
  deleting both quarantine calls (BR-1); the `TestReviewConcurrencySchedules` quarantine narrowed to its 20s
  start-deadline step, so the schedules stay covered (it passed in full, 68s, at lower load); `processAlive` reads
  only ESRCH as gone, the same reading as weave staging; the atlas and AGENTS.local timings now say about 8 minutes.
- Landing: main moved twice under the PR. On the second merge, #310's side-quest (3c84da0a) had independently fixed
  the same `processAlive` EPERM bug. Resolution: kept this branch's only-ESRCH-is-gone reading (close-review BR-3)
  with main's comment, and dropped this branch's duplicate test for main's `TestProcessAliveTreatsEPERMAsAlive`.
